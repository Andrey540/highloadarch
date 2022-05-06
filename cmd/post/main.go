package main

import (
	"context"
	"encoding/json"
	"fmt"
	stdlog "log"
	"net/http"
	"time"

	commonapp "github.com/callicoder/go-docker/pkg/common/app"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/event"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/metrics"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/mysql"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/redis"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/request"
	postrequest "github.com/callicoder/go-docker/pkg/common/infrastructure/request/post"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/response"
	postresponse "github.com/callicoder/go-docker/pkg/common/infrastructure/response/post"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/server"
	"github.com/callicoder/go-docker/pkg/common/uuid"
	"github.com/callicoder/go-docker/pkg/post/app"
	"github.com/callicoder/go-docker/pkg/post/app/command"
	infrastructure "github.com/callicoder/go-docker/pkg/post/infrastructure"
	"github.com/gorilla/mux"
)

const (
	appID = postrequest.AppID
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
		return err
	}

	connector := mysql.NewConnector()
	err = connector.MigrateUp(cnf.dsn(), cnf.MigrationsDir)
	if err != nil {
		return err
	}
	err = connector.Open(cnf.dsn(), mysql.Config{MaxConnections: cnf.DBMaxConn, ConnectionLifetime: time.Duration(cnf.DBConnectionLifetime) * time.Second})
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
	newsLineCache, err := infrastructure.NewNewsLineCache(&redis.Config{
		Password: cnf.RedisPassword,
		Host:     cnf.RedisHost + ":" + cnf.RedisPort,
	})
	if err != nil {
		return err
	}
	// nolint: errcheck
	defer newsLineCache.Stop()

	commonUnitOfWorkFactory := infrastructure.NewUnitOfWorkFactory(mysqlClient)

	transports, connections, err1 := server.InitEventTransport(nil, cnf.amqpConf(), logger, errorLogger)

	if err1 != nil {
		errorLogger.Println(err1)
		return err1
	}

	eventDispatcher := commonapp.NewStoredEventDispatcher(commonUnitOfWorkFactory, transports, eventDispatcherErrorsCh)
	unitOfWorkFactory := infrastructure.NewNotifyingUnitOfWorkFactory(commonUnitOfWorkFactory, eventDispatcher.Activate)

	commandHandlerFactory := command.NewCommandHandlerFactory()
	commandsHandler := commonapp.NewCommandsHandler(unitOfWorkFactory, commandHandlerFactory)

	handlerFactory := app.NewEventHandlerFactory(newsLineCache)
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

	newsLineQueryService := infrastructure.NewNewsLineQueryService(connector.TransactionalClient(), *newsLineCache)

	stopChan := make(chan struct{})
	server.ListenOSKillSignals(stopChan)
	serverHub := server.NewHub(stopChan)
	serveHTTP(cnf, serverHub, newsLineQueryService, commandsHandler, logger, errorLogger, metricsHandler)

	return serverHub.Wait()
}

func serveHTTP(config *config, serverHub *server.Hub, queryService app.NewsLineQueryService, commandsHandler commonapp.CommandHandler, logger, errorLogger *stdlog.Logger,
	metricsHandler metrics.PrometheusMetricsHandler) {
	ctx := context.Background()
	_, cancel := context.WithCancel(ctx)
	var httpServer *http.Server
	serverHub.Serve(func() error {
		router := mux.NewRouter()
		router.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
			response.WriteSuccessResponse(w)
		}).Methods(http.MethodGet)
		router.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
			response.WriteSuccessResponse(w)
		}).Methods(http.MethodGet)

		router.HandleFunc(postrequest.CreatePostURL, createPost(commandsHandler)).Methods(http.MethodPost)
		router.HandleFunc(postrequest.ListPostsURL, listPosts(queryService)).Methods(http.MethodGet)
		router.HandleFunc(postrequest.ListNewsURL, listNews(queryService)).Methods(http.MethodGet)
		router.HandleFunc(postrequest.GetPostURL, getPost(queryService)).Methods(http.MethodGet)

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

func createPost(commandsHandler commonapp.CommandHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		createPostRequest, err := getCreatePostRequest(r)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		userID := request.GetUserIDFromHeader(r)
		createPostCommand := command.CreatePost{
			ID:       request.GetRequestIDFromRequest(r),
			AuthorID: userID,
			Title:    createPostRequest.Title,
			Text:     createPostRequest.Text,
		}
		id, err := commandsHandler.Handle(createPostCommand)
		if err != nil {
			processError(err, w)
			return
		}
		writePostResponse(id.(uuid.UUID).String(), w)
	}
}

func listPosts(service app.NewsLineQueryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDStr := request.GetUserIDFromHeader(r)
		userID, err := uuid.FromString(userIDStr)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		posts, err := service.ListPosts(userID)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		result := []postresponse.Data{}
		for _, post := range posts {
			result = append(result, postresponse.Data{ID: post.ID.String(), AuthorID: post.Author.String(), Title: post.Title, Text: post.Text})
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

func listNews(service app.NewsLineQueryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDStr := request.GetUserIDFromHeader(r)
		userID, err := uuid.FromString(userIDStr)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		posts, err := service.ListNews(userID)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		result := []postresponse.NewsListItem{}
		for _, post := range *posts {
			result = append(result, postresponse.NewsListItem{ID: post.ID.String(), AuthorID: post.Author.String(), Title: post.Title})
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

func getPost(service app.NewsLineQueryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		postIDStr := request.GetIDFromRequest(r)
		postID, err := uuid.FromString(postIDStr)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		post, err := service.GetPost(postID)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		if post == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		data, err := json.Marshal(postresponse.Data{ID: post.ID.String(), AuthorID: post.Author.String(), Title: post.Title, Text: post.Text})
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(data)
	}
}

func writePostResponse(id string, w http.ResponseWriter) {
	data, err := json.Marshal(postresponse.Post{PostID: id})
	if err != nil {
		response.WriteErrorResponse(err, w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func getCreatePostRequest(r *http.Request) (postrequest.CreatePost, error) {
	decoder := json.NewDecoder(r.Body)
	var createPostRequest postrequest.CreatePost
	err := decoder.Decode(&createPostRequest)
	return createPostRequest, err
}

func processError(err error, w http.ResponseWriter) {
	if err == commonapp.ErrCommandAlreadyProcessed {
		response.WriteDuplicateRequestResponse(err, w)
		return
	}
	response.WriteErrorResponse(err, w)
}
