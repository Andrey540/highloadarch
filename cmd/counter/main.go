package main

import (
	"context"
	stdlog "log"
	"time"

	api "github.com/callicoder/go-docker/pkg/common/api"
	commonapp "github.com/callicoder/go-docker/pkg/common/app"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/event"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/metrics"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/server"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/vitess"
	"github.com/callicoder/go-docker/pkg/counter/app"
	"github.com/callicoder/go-docker/pkg/counter/infrastructure"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

const (
	appID = "counter"
)

func main() {
	logger := server.InitLogger()
	errorLogger := server.InitErrorLogger()
	cnf, err := parseEnv()
	if err != nil {
		errorLogger.Println(err)
	}
	err = runService(cnf, logger, errorLogger)
	if err == server.ErrStopped {
		logger.Println("service is successfully stopped")
	} else if err != nil {
		errorLogger.Println(err)
	}
}

func runService(cnf *config, logger, errorLogger *stdlog.Logger) error {
	metricsHandler, err := metrics.NewPrometheusMetricsHandler(appID)
	if err != nil {
		errorLogger.Println(err)
		return err
	}

	connector := vitess.NewConnector()
	err = connector.Open(cnf.dbDsn(), vitess.Config{MaxConnections: cnf.DBMaxConn, ConnectionLifetime: time.Duration(cnf.DBConnectionLifetime) * time.Second}, "@primary")
	if err != nil {
		errorLogger.Println(err)
		return err
	}
	// noinspection GoUnhandledErrorResult
	defer connector.Close()

	schemaLoader := vitess.NewSchemaLoader(logger)
	err = schemaLoader.Migrate(cnf.schemaDsn(), cnf.VSchemaPath, cnf.DBName)
	if err != nil {
		errorLogger.Println(err)
		return err
	}

	mysqlClient := connector.TransactionalClient()
	eventDispatcherErrorsCh := make(chan error)
	go func() {
		for err := range eventDispatcherErrorsCh {
			errorLogger.Println(err)
		}
	}()

	commonUnitOfWorkFactory := infrastructure.NewUnitOfWorkFactory(mysqlClient, cnf.DBName)
	transports, connections, err1 := server.InitEventTransport(nil, cnf.amqpConf(), logger, errorLogger)

	if err1 != nil {
		errorLogger.Println(err1)
		return err1
	}

	eventDispatcher := commonapp.NewStoredEventDispatcher(commonUnitOfWorkFactory, transports, eventDispatcherErrorsCh)
	unitOfWorkFactory := infrastructure.NewNotifyingUnitOfWorkFactory(commonUnitOfWorkFactory, eventDispatcher.Activate)

	handlerFactory := app.NewEventHandlerFactory(errorLogger)
	eventsHandler := event.NewEventsHandler(unitOfWorkFactory, handlerFactory, logger, errorLogger)

	for _, transport := range transports {
		transport.(event.Transport).SetHandler(eventsHandler)
	}

	defer func() {
		for _, connection := range connections {
			// noinspection GoUnhandledErrorResult
			connection.Stop() // nolint:errcheck
		}
	}()

	eventDispatcher.Start()
	defer eventDispatcher.Stop()

	queryService := infrastructure.NewUnreadMessagesQueryService(connector.TransactionalClient(), cnf.DBName)

	stopChan := make(chan struct{})
	server.ListenOSKillSignals(stopChan)
	serverHub := server.NewHub(stopChan)

	// Serve grpc
	grpcServer := infrastructure.NewGRPCServer(queryService)
	baseServer := grpc.NewServer(grpc.UnaryInterceptor(server.MakeGrpcUnaryInterceptor(logger, errorLogger)))
	api.RegisterCounterServer(baseServer, grpcServer)
	server.ServeGRPC(cnf.ServeGRPCAddress, serverHub, baseServer)

	// Serve http endpoints and grpc-gateway proxy
	server.ServeHTTP(
		cnf.ServeGRPCAddress,
		cnf.ServeRESTAddress,
		appID,
		serverHub,
		metricsHandler,
		func(ctx context.Context, grpcGatewayMux *runtime.ServeMux, address string, opts []grpc.DialOption) error {
			return api.RegisterCounterHandlerFromEndpoint(ctx, grpcGatewayMux, cnf.ServeGRPCAddress, opts)
		},
		logger, errorLogger)

	return serverHub.Wait()
}
