package main

import (
	"context"
	"encoding/json"

	"fmt"
	stdlog "log"
	"net/http"
	"time"

	commonapp "github.com/callicoder/go-docker/pkg/common/app"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/metrics"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/request"
	conversationrequest "github.com/callicoder/go-docker/pkg/common/infrastructure/request/conversation"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/response"
	conversationresponse "github.com/callicoder/go-docker/pkg/common/infrastructure/response/conversation"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/server"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/vitess"
	"github.com/callicoder/go-docker/pkg/common/uuid"
	"github.com/callicoder/go-docker/pkg/conversation/app"
	"github.com/callicoder/go-docker/pkg/conversation/app/command"
	infrastructure "github.com/callicoder/go-docker/pkg/conversation/infrastructure"
	"github.com/gorilla/mux"
)

const (
	appID = conversationrequest.AppID
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
	/*err = connector.MigrateUp(cnf.dbDsn(), cnf.MigrationsDir, cnf.DBName)
	if err != nil {
		errorLogger.Println(err)
		return err
	}*/
	schemaLoader := vitess.NewSchemaLoader(logger)
	err = schemaLoader.Migrate(cnf.schemaDsn(), cnf.VSchemaPath, cnf.DBName)
	if err != nil {
		errorLogger.Println(err)
		return err
	}

	if err != nil {
		errorLogger.Println(err)
		return err
	}
	err = connector.Open(cnf.dbDsn(), vitess.Config{MaxConnections: cnf.DBMaxConn, ConnectionLifetime: time.Duration(cnf.DBConnectionLifetime) * time.Second}, "@primary")
	if err != nil {
		errorLogger.Println(err)
		return err
	}
	// noinspection GoUnhandledErrorResult
	defer connector.Close()

	eventDispatcherErrorsCh := make(chan error)
	go func() {
		for err := range eventDispatcherErrorsCh {
			errorLogger.Println(err)
		}
	}()

	mysqlClient := connector.TransactionalClient()
	commonUnitOfWorkFactory := infrastructure.NewUnitOfWorkFactory(mysqlClient)

	transports := []commonapp.Transport{}
	eventDispatcher := commonapp.NewStoredEventDispatcher(commonUnitOfWorkFactory, transports, eventDispatcherErrorsCh)
	unitOfWorkFactory := infrastructure.NewNotifyingUnitOfWorkFactory(commonUnitOfWorkFactory, eventDispatcher.Activate)

	commandHandlerFactory := command.NewCommandHandlerFactory()
	commandsHandler := commonapp.NewCommandsHandler(unitOfWorkFactory, commandHandlerFactory)

	conversationQueryService := infrastructure.NewConversationQueryService(connector.TransactionalClient())

	stopChan := make(chan struct{})
	server.ListenOSKillSignals(stopChan)
	serverHub := server.NewHub(stopChan)

	serveHTTP(cnf, serverHub, conversationQueryService, commandsHandler, logger, errorLogger, metricsHandler)

	return serverHub.Wait()
}

func serveHTTP(config *config, serverHub *server.Hub, queryService app.ConversationQueryService, commandsHandler commonapp.CommandHandler, logger, errorLogger *stdlog.Logger,
	metricsHandler metrics.PrometheusMetricsHandler) {
	ctx := context.Background()
	_, cancel := context.WithCancel(ctx)
	var httpServer *http.Server
	serverHub.Serve(func() error {
		router := mux.NewRouter()
		router.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
			response.WriteSuccessResponse(w)
		})
		router.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
			response.WriteSuccessResponse(w)
		})

		router.HandleFunc(conversationrequest.StartConversationURL, startConversation(commandsHandler)).Methods(http.MethodPost)
		router.HandleFunc(conversationrequest.AddMessageURL, addMessage(commandsHandler)).Methods(http.MethodPost)
		router.HandleFunc(conversationrequest.GetConversationURL, getConversation(queryService)).Methods(http.MethodGet)

		nextRequestID := func() string {
			return fmt.Sprintf("%d", time.Now().UnixNano())
		}

		metricsHandler.AddMetricsMiddleware(router)
		router.Use(server.RecoverMiddleware(errorLogger))
		router.Use(server.TracingMiddleware(nextRequestID))
		router.Use(server.LoggingMiddleware(logger))

		httpServer = &http.Server{
			Handler:           router,
			Addr:              config.ServeRESTAddress,
			ReadHeaderTimeout: 10 * time.Second,
			ReadTimeout:       time.Hour,
			WriteTimeout:      time.Hour,
			ErrorLog:          errorLogger,
		}
		return httpServer.ListenAndServe()
	}, func() error {
		cancel()
		return httpServer.Shutdown(context.Background())
	})
}

func startConversation(commandsHandler commonapp.CommandHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var startConversationRequest conversationrequest.StartUserConversation
		err := decoder.Decode(&startConversationRequest)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		startConversationCommand := command.StartUserConversation{
			ID:     request.GetRequestIDFromRequest(r),
			User:   startConversationRequest.User,
			Target: startConversationRequest.Target,
		}
		id, err := commandsHandler.Handle(startConversationCommand)
		if err != nil {
			processError(err, w)
			return
		}
		writeConversationResponse(id.(uuid.UUID).String(), w)
	}
}

func addMessage(commandsHandler commonapp.CommandHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var addMessageRequest conversationrequest.AddMessage
		err := decoder.Decode(&addMessageRequest)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		userID := request.GetUserIDFromHeader(r)
		addMessageCommand := command.AddMessage{
			ID:             request.GetRequestIDFromRequest(r),
			ConversationID: addMessageRequest.ConversationID,
			UserID:         userID,
			Text:           addMessageRequest.Text,
		}
		id, err := commandsHandler.Handle(addMessageCommand)
		if err != nil {
			processError(err, w)
			return
		}
		writeMessageResponse(id.(uuid.UUID).String(), w)
	}
}

func getConversation(service app.ConversationQueryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conversationIDStr := request.GetIDFromRequest(r)
		conversationID, err := uuid.FromString(conversationIDStr)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		messages, err := service.ListMessages(conversationID)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		result := make([]conversationresponse.MessageData, 0, len(messages))
		for _, message := range messages {
			result = append(result, conversationresponse.MessageData{ID: message.ID.String(), ConversationID: message.ConversationID.String(), UserID: message.UserID.String(), Text: message.Text})
		}
		data, err := json.Marshal(result)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}
}

func writeConversationResponse(id string, w http.ResponseWriter) {
	data, err := json.Marshal(conversationresponse.Conversation{ConversationID: id})
	if err != nil {
		response.WriteErrorResponse(err, w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func writeMessageResponse(id string, w http.ResponseWriter) {
	data, err := json.Marshal(conversationresponse.Message{MessageID: id})
	if err != nil {
		response.WriteErrorResponse(err, w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func processError(err error, w http.ResponseWriter) {
	if err == commonapp.ErrCommandAlreadyProcessed {
		response.WriteDuplicateRequestResponse(err, w)
		return
	}
	response.WriteErrorResponse(err, w)
}
