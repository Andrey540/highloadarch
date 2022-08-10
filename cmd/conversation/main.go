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
	"github.com/callicoder/go-docker/pkg/conversation/app"
	"github.com/callicoder/go-docker/pkg/conversation/app/command"
	"github.com/callicoder/go-docker/pkg/conversation/infrastructure"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"google.golang.org/grpc"
)

const (
	appID = "conversation"
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
	metricsHandler, err := metrics.NewPrometheusMetricsHandler()
	if err != nil {
		errorLogger.Println(err)
		return err
	}

	if cnf.HTTPServerEnabled == 1 {
		restPort := cnf.ServeRESTAddress[1:len(cnf.ServeRESTAddress)]
		grpcPort := cnf.ServeGRPCAddress[1:len(cnf.ServeGRPCAddress)]

		err = server.ServiceRegistryWithConsul(appID, cnf.ServiceID, restPort, ":"+restPort+"/health", []string{"urlprefix-" + appID + "/"})
		if err != nil {
			errorLogger.Println(err)
			return err
		}
		err = server.ServiceRegistryWithConsul(appID+"-grpc", cnf.ServiceID, grpcPort, ":"+restPort+"/health", []string{"urlprefix-/api.Conversation" + " proto=grpc grpcservername=" + cnf.ServiceID})
		if err != nil {
			errorLogger.Println(err)
			return err
		}
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

	realtimeHosts, err := cnf.realtimeHosts()
	if err != nil {
		return err
	}
	userNotifier, err := infrastructure.NewUserNotifier(realtimeHosts, "message")
	if err != nil {
		return err
	}
	// noinspection GoUnhandledErrorResult
	defer userNotifier.Close()

	commonUnitOfWorkFactory := infrastructure.NewUnitOfWorkFactory(mysqlClient, cnf.DBName)
	transports, connections, err1 := server.InitEventTransport(nil, cnf.amqpConf(), logger, errorLogger)

	if err1 != nil {
		errorLogger.Println(err1)
		return err1
	}

	eventDispatcher := commonapp.NewStoredEventDispatcher(commonUnitOfWorkFactory, transports, eventDispatcherErrorsCh)
	unitOfWorkFactory := infrastructure.NewNotifyingUnitOfWorkFactory(commonUnitOfWorkFactory, eventDispatcher.Activate)

	commandHandlerFactory := command.NewCommandHandlerFactory()
	commandsHandler := commonapp.NewCommandsHandler(unitOfWorkFactory, commandHandlerFactory)

	handlerFactory := app.NewEventHandlerFactory(userNotifier, errorLogger)
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

	conversationQueryService := infrastructure.NewConversationQueryService(connector.TransactionalClient(), cnf.DBName)

	stopChan := make(chan struct{})
	server.ListenOSKillSignals(stopChan)
	serverHub := server.NewHub(stopChan)

	// Serve grpc
	grpcServer := infrastructure.NewGRPCServer(conversationQueryService, commandsHandler)
	baseServer := grpc.NewServer(grpc.UnaryInterceptor(server.MakeGrpcUnaryInterceptor(metricsHandler, logger, errorLogger)))
	api.RegisterConversationServer(baseServer, grpcServer)
	server.ServeGRPC(cnf.ServeGRPCAddress, serverHub, baseServer)

	// Serve http endpoints and grpc-gateway proxy
	server.ServeHTTP(
		cnf.ServeGRPCAddress,
		cnf.ServeRESTAddress,
		appID,
		connector,
		serverHub,
		metricsHandler,
		func(ctx context.Context, grpcGatewayMux *runtime.ServeMux, address string, opts []grpc.DialOption) error {
			return api.RegisterConversationHandlerFromEndpoint(ctx, grpcGatewayMux, cnf.ServeGRPCAddress, opts)
		},
		logger, errorLogger)

	return serverHub.Wait()
}
