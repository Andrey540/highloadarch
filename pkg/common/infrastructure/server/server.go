package server

import (
	"github.com/callicoder/go-docker/pkg/common/app"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/amqp"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/event"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/kafka"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/metrics"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/request"
	"github.com/gorilla/mux"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pkg/errors"
	satoriuuid "github.com/satori/go.uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"context"
	"io"
	stdlog "log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

// ServeFunc - runs server
type ServeFunc func() error

// StopFunc - stops server
type StopFunc func() error

const (
	serverIsCreated int32 = iota
	serverIsRunning
	serverIsStopped
)

type server struct {
	serveFunc ServeFunc
	stopFunc  StopFunc
	state     int32
}

func newServer(serve ServeFunc, stop StopFunc) *server {
	return &server{
		serveFunc: serve,
		stopFunc:  stop,
		state:     serverIsCreated,
	}
}

func (s *server) serve() error {
	if !atomic.CompareAndSwapInt32(&s.state, serverIsCreated, serverIsRunning) {
		if atomic.LoadInt32(&s.state) == serverIsRunning {
			return errAlreadyRun
		}
		return errTryRunStoppedServer
	}
	return s.serveFunc()
}

func (s *server) stop() error {
	stopped := atomic.CompareAndSwapInt32(&s.state, serverIsCreated, hubIsStopped) ||
		atomic.CompareAndSwapInt32(&s.state, serverIsRunning, serverIsStopped)

	if !stopped {
		return errAlreadyStopped
	}
	return s.stopFunc()
}

func ListenOSKillSignals(stopChan chan<- struct{}) {
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT)
		<-ch
		stopChan <- struct{}{}
	}()
}

func ServeHTTP(
	serveGRPCAddress, serveRESTAddress, appID string,
	serverHub *Hub,
	metricsHandler metrics.PrometheusMetricsHandler,
	registerFunc func(context.Context, *runtime.ServeMux, string, []grpc.DialOption) error,
	logger, errorLogger *stdlog.Logger,
) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	var httpServer *http.Server
	serverHub.Serve(func() error {
		grpcGatewayMux := runtime.NewServeMux(
			runtime.WithMetadata(metadataConverter),
			runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
				OrigName:     true,
				EmitDefaults: true,
			}),
		)
		opts := []grpc.DialOption{grpc.WithInsecure(), grpc.WithDefaultCallOptions()}
		err := registerFunc(ctx, grpcGatewayMux, serveGRPCAddress, opts)
		if err != nil {
			return err
		}

		router := mux.NewRouter()
		router.PathPrefix("/" + appID + "/api/").Handler(http.TimeoutHandler(grpcGatewayMux, 15*time.Second, ""))

		// Implement healthcheck for Kubernetes
		router.HandleFunc("/resilience/ready", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, http.StatusText(http.StatusOK))
		}).Methods(http.MethodGet)
		router.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, http.StatusText(http.StatusOK))
		}).Methods(http.MethodGet)
		router.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, http.StatusText(http.StatusOK))
		}).Methods(http.MethodGet)

		nextRequestID := func() string {
			return satoriuuid.NewV1().String()
		}

		metricsHandler.AddMetricsMiddleware(router)
		router.Use(RecoverMiddleware(errorLogger))
		router.Use(TracingMiddleware(nextRequestID))

		httpServer = &http.Server{
			Handler:      router,
			Addr:         serveRESTAddress,
			ReadTimeout:  time.Hour,
			WriteTimeout: time.Hour,
		}

		return httpServer.ListenAndServe()
	}, func() error {
		cancel()
		return httpServer.Shutdown(context.Background())
	})
}

func ServeGRPC(serveGRPCAddress string, serverHub *Hub, baseServer *grpc.Server) {
	serverHub.Serve(func() error {
		grpcListener, grpcErr := net.Listen("tcp", serveGRPCAddress)
		if grpcErr != nil {
			return errors.Wrapf(grpcErr, "failed to listen port %s", serveGRPCAddress)
		}
		grpcErr = baseServer.Serve(grpcListener)
		return errors.Wrap(grpcErr, "failed to serve GRPC")
	}, func() error {
		baseServer.GracefulStop()
		return nil
	})
}

func MakeGrpcUnaryInterceptor(logger, errorLogger *stdlog.Logger) grpc.UnaryServerInterceptor {
	loggerInterceptor := makeLoggerServerInterceptor(logger, errorLogger)
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		resp, err = loggerInterceptor(ctx, req, info, handler)
		return resp, translateGRPCError(err)
	}
}

func InitLogger() *stdlog.Logger {
	return stdlog.New(os.Stdout, "http: ", stdlog.LstdFlags)
}

func InitErrorLogger() *stdlog.Logger {
	return stdlog.New(os.Stderr, "http: ", stdlog.LstdFlags)
}

func InitEventTransport(kafkaCnf *kafka.Config, amqpCnf *amqp.Config, logger, errorLogger *stdlog.Logger) ([]app.Transport, []event.Connection, error) {
	var transports []app.Transport
	var connections []event.Connection
	if kafkaCnf != nil {
		transport, connection, err := kafka.CreateTransport(*kafkaCnf, logger, errorLogger)
		if err != nil {
			return nil, nil, err
		}
		transports = append(transports, transport)
		connections = append(connections, connection)
	}
	if amqpCnf != nil {
		transport, connection, err := amqp.CreateTransport(*amqpCnf, logger, errorLogger)
		if err != nil {
			return nil, nil, err
		}
		transports = append(transports, transport)
		connections = append(connections, connection)
	}
	return transports, connections, nil
}

func GetRequestIDFromContext(r *http.Request) string {
	requestID, ok := r.Context().Value(RequestIDKey).(string)
	if !ok {
		return ""
	}
	return requestID
}

func GetRequestIDFromGRPCMetadata(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	values := md.Get(request.RequestIDHeader) // metadata.Get returns an array of values for the key
	if len(values) > 0 {
		return values[0]
	}
	return ""
}

func GetUserIDFromGRPCMetadata(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	values := md.Get(request.UserIDHeader) // metadata.Get returns an array of values for the key
	if len(values) > 0 {
		return values[0]
	}
	return ""
}

func Authenticate(ctx context.Context) error {
	userID := GetUserIDFromGRPCMetadata(ctx)
	if userID == "" {
		return errors.WithStack(app.ErrNotAuthenticated)
	}
	return nil
}

func metadataConverter(_ context.Context, r *http.Request) metadata.MD {
	userID := request.GetUserIDFromRequest(r)
	requestID := request.GetRequestIDFromRequest(r)
	return metadata.New(map[string]string{request.UserIDHeader: userID, request.RequestIDHeader: requestID})
}

func makeLoggerServerInterceptor(logger, errorLogger *stdlog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		start := time.Now()

		resp, err = handler(ctx, req)

		duration := time.Since(start).String()
		fields := logrus.Fields{
			"requestID": GetRequestIDFromGRPCMetadata(ctx),
			"args":      req,
			"duration":  duration,
			"method":    extractMethodName(info),
		}

		if err != nil {
			errorLogger.Println(err, "call failed")
		} else {
			logger.Println(fields, "call finished")
		}
		return resp, err
	}
}

func extractMethodName(info *grpc.UnaryServerInfo) string {
	method := info.FullMethod
	return method[strings.LastIndex(method, "/")+1:]
}

func translateGRPCError(err error) error {
	// if already a GRPC error return unchanged
	if _, ok := status.FromError(err); ok {
		return err
	}
	switch errors.Cause(err) {
	case app.ErrCommandAlreadyProcessed:
		return status.Errorf(codes.AlreadyExists, err.Error())
	case app.ErrNotAuthenticated:
		return status.Errorf(codes.Unauthenticated, err.Error())
	case app.ErrPermissionDenied:
		return status.Errorf(codes.PermissionDenied, err.Error())
	case app.ErrNotFound:
		return status.Errorf(codes.NotFound, err.Error())
	case nil:
		return nil
	default:
		return status.Errorf(codes.Internal, err.Error())
	}
}

var errAlreadyStopped = errors.New("server is not running, can't change server state")
var errAlreadyRun = errors.New("server is running, can't change server state to running")
var errTryRunStoppedServer = errors.New("server is stopped, can't change server state to running")
