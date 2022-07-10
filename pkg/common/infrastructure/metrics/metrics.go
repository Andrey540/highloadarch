package metrics

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewPrometheusMetricsHandler() (PrometheusMetricsHandler, error) {
	labelNames := []string{"endpoint", "method"}

	latencyHistogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "app_request_latency_seconds",
		Help:    "Application Request Latency.",
		Buckets: prometheus.DefBuckets,
	}, labelNames)
	err := prometheus.Register(latencyHistogram)
	if err != nil {
		return nil, err
	}

	requestCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "app_request_count",
		Help: "Application Request Count.",
	}, labelNames)
	err = prometheus.Register(requestCounter)
	if err != nil {
		return nil, err
	}

	serverErrorCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "app_server_error_count",
		Help: "Application Server Error (5xx) Count.",
	}, labelNames)
	err = prometheus.Register(serverErrorCounter)
	if err != nil {
		return nil, err
	}

	labelNames = []string{"methodName", "code"}

	grpcLatencyHistogram := prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "app_grpc_request_latency_seconds",
		Help:    "Application GRPC Request Latency.",
		Buckets: prometheus.DefBuckets,
	}, labelNames)
	err = prometheus.Register(grpcLatencyHistogram)
	if err != nil {
		return nil, err
	}

	grpcRequestCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "app_grpc_request_count",
		Help: "Application GRPC Request Count.",
	}, labelNames)
	err = prometheus.Register(grpcRequestCounter)
	if err != nil {
		return nil, err
	}

	grpcServerErrorCounter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "app_grpc_server_error_count",
		Help: "Application GRPC Server Error (5xx) Count.",
	}, labelNames)
	err = prometheus.Register(grpcServerErrorCounter)
	if err != nil {
		return nil, err
	}

	return &prometheusMetricsHandler{
		latencyHistogram:       latencyHistogram,
		requestCounter:         requestCounter,
		serverErrorCounter:     serverErrorCounter,
		grpcLatencyHistogram:   grpcLatencyHistogram,
		grpcRequestCounter:     grpcRequestCounter,
		grpcServerErrorCounter: grpcServerErrorCounter,
	}, nil
}

type PrometheusMetricsHandler interface {
	AddMetricsMiddleware(router *mux.Router)
	AddGRPCMetricsMiddleware(baseInterceptor grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor
}

type prometheusMetricsHandler struct {
	latencyHistogram       *prometheus.HistogramVec
	requestCounter         *prometheus.CounterVec
	serverErrorCounter     *prometheus.CounterVec
	grpcLatencyHistogram   *prometheus.HistogramVec
	grpcRequestCounter     *prometheus.CounterVec
	grpcServerErrorCounter *prometheus.CounterVec
}

func (p *prometheusMetricsHandler) AddMetricsMiddleware(router *mux.Router) {
	router.Handle("/metrics", promhttp.Handler())

	router.Use(p.prometheusMiddleware())
}

func (p *prometheusMetricsHandler) AddGRPCMetricsMiddleware(baseInterceptor grpc.UnaryServerInterceptor) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		start := time.Now()

		resp, err = baseInterceptor(ctx, req, info, handler)

		duration := time.Since(start)
		grpcStatus, _ := status.FromError(err)
		labels := []string{extractMethodName(info), fmt.Sprintf("%d", grpcStatus.Code())}
		p.grpcLatencyHistogram.WithLabelValues(labels...).Observe(duration.Seconds())
		p.grpcRequestCounter.WithLabelValues(labels...).Inc()
		if grpcStatus.Code() == codes.Internal || grpcStatus.Code() == codes.Unavailable || grpcStatus.Code() == codes.Unknown {
			p.serverErrorCounter.WithLabelValues(labels...).Inc()
		}

		return resp, err
	}
}

func (p *prometheusMetricsHandler) prometheusMiddleware() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			start := time.Now()
			code := http.StatusOK

			defer func() {
				httpDuration := time.Since(start)
				var labels []string
				re := regexp.MustCompile("(^*)/[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$")
				url := re.ReplaceAllString(req.URL.Path, `$1`)
				labels = append(labels, url)
				labels = append(labels, req.Method)
				p.latencyHistogram.WithLabelValues(labels...).Observe(httpDuration.Seconds())
				p.requestCounter.WithLabelValues(labels...).Inc()
				if code >= 500 {
					p.serverErrorCounter.WithLabelValues(labels...).Inc()
				}
			}()

			rw := &responseWriter{w, http.StatusOK}
			next.ServeHTTP(rw, req)
			code = rw.statusCode
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func extractMethodName(info *grpc.UnaryServerInfo) string {
	method := info.FullMethod
	return method[strings.LastIndex(method, "/")+1:]
}
