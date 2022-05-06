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
	"github.com/callicoder/go-docker/pkg/common/infrastructure/request"
	userrequest "github.com/callicoder/go-docker/pkg/common/infrastructure/request/user"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/response"
	userresponse "github.com/callicoder/go-docker/pkg/common/infrastructure/response/user"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/server"
	"github.com/callicoder/go-docker/pkg/common/uuid"
	"github.com/callicoder/go-docker/pkg/user/app"
	"github.com/callicoder/go-docker/pkg/user/app/command"
	infrastructure "github.com/callicoder/go-docker/pkg/user/infrastructure"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

const (
	appID = userrequest.AppID
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

	masterConnector := mysql.NewConnector()
	err = masterConnector.MigrateUp(cnf.masterDSN(), cnf.MigrationsDir)
	if err != nil {
		return err
	}
	err = masterConnector.Open(cnf.masterDSN(), mysql.Config{MaxConnections: cnf.DBMaxConn, ConnectionLifetime: time.Duration(cnf.DBConnectionLifetime) * time.Second})
	if err != nil {
		errorLogger.Println(err)
		return err
	}
	// noinspection GoUnhandledErrorResult
	defer masterConnector.Close()

	slaveConnector := mysql.NewConnector()
	err = slaveConnector.Open(cnf.slaveDSN(), mysql.Config{MaxConnections: cnf.DBMaxConn, ConnectionLifetime: time.Duration(cnf.DBConnectionLifetime) * time.Second})
	if err != nil {
		errorLogger.Println(err)
		return err
	}
	// noinspection GoUnhandledErrorResult
	defer slaveConnector.Close()

	eventDispatcherErrorsCh := make(chan error)
	go func() {
		for err := range eventDispatcherErrorsCh {
			errorLogger.Println(err)
		}
	}()

	mysqlClient := masterConnector.TransactionalClient()
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

	handlerFactory := app.NewEventHandlerFactory()
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

	userQueryService := infrastructure.NewUserQueryService(slaveConnector.TransactionalClient())

	stopChan := make(chan struct{})
	server.ListenOSKillSignals(stopChan)
	serverHub := server.NewHub(stopChan)
	serveHTTP(cnf, serverHub, userQueryService, commandsHandler, logger, errorLogger, metricsHandler)

	return serverHub.Wait()
}

func serveHTTP(config *config, serverHub *server.Hub, queryService app.UserQueryService, commandsHandler commonapp.CommandHandler, logger, errorLogger *stdlog.Logger,
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

		router.HandleFunc(userrequest.SignInURL, authUser(queryService)).Methods(http.MethodPost)
		router.HandleFunc(userrequest.RegisterURL, registerUser(commandsHandler)).Methods(http.MethodPost)
		router.HandleFunc(userrequest.ProfileURL, getUser(queryService)).Methods(http.MethodGet)
		router.HandleFunc(userrequest.FindUserURL, findUsers(queryService)).Methods(http.MethodGet)
		router.HandleFunc(userrequest.UpdateUserURL, updateUser(commandsHandler)).Methods(http.MethodPut)
		router.HandleFunc(userrequest.DeleteURL, deleteUser(commandsHandler)).Methods(http.MethodDelete)
		router.HandleFunc(userrequest.AddFriendURL, addUserFriend(commandsHandler)).Methods(http.MethodPost)
		router.HandleFunc(userrequest.RemoveFriendURL, removeUserFriend(commandsHandler)).Methods(http.MethodPost)
		router.HandleFunc(userrequest.ListUserFriendsURL, listUserFriends(queryService)).Methods(http.MethodGet)
		router.HandleFunc(userrequest.ListUsersURL, listUsers(queryService)).Methods(http.MethodGet)

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

func authUser(service app.UserQueryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var authRequest userrequest.Auth
		err := decoder.Decode(&authRequest)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		user, err := service.GetUserByNameAndPassword(authRequest.UserName, authRequest.Password)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		if user == nil {
			response.WriteNotFoundResponse(errors.New("User not found"), w)
			return
		}
		writeUserResponse(user.ID.String(), w)
	}
}

func registerUser(commandsHandler commonapp.CommandHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		registerUserRequest, err := getRegisterUserRequest(r)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		registerUserCommand := command.RegisterUser{
			ID:        request.GetRequestIDFromRequest(r),
			Username:  registerUserRequest.Username,
			FirstName: registerUserRequest.FirstName,
			LastName:  registerUserRequest.LastName,
			Age:       registerUserRequest.Age,
			Sex:       registerUserRequest.Sex,
			Password:  registerUserRequest.Password,
			City:      registerUserRequest.City,
			Interests: registerUserRequest.Interests,
		}
		id, err := commandsHandler.Handle(registerUserCommand)
		if err != nil {
			processError(err, w)
			return
		}
		writeUserResponse(id.(uuid.UUID).String(), w)
	}
}

func getUser(service app.UserQueryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDStr := request.GetIDFromRequest(r)
		userID, err := uuid.FromString(userIDStr)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		user, err := service.GetUserProfile(userID)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		writeGetUserResponse(user, w)
	}
}

func findUsers(service app.UserQueryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		userName := vars["username"]
		users, err := service.ListUserProfiles(userName)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		var result = make([]userresponse.Data, len(users))
		for _, user := range users {
			result = append(result, convertToUser(user))
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

func updateUser(commandsHandler commonapp.CommandHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		updateUserRequest, err := getUpdateUserRequest(r)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		userID := request.GetIDFromRequest(r)
		updateUserCommand := command.UpdateUser{
			ID:        request.GetRequestIDFromRequest(r),
			UserID:    userID,
			Username:  updateUserRequest.Username,
			FirstName: updateUserRequest.FirstName,
			LastName:  updateUserRequest.LastName,
			Age:       updateUserRequest.Age,
			Sex:       updateUserRequest.Sex,
			Password:  updateUserRequest.Password,
			City:      updateUserRequest.City,
			Interests: updateUserRequest.Interests,
		}
		_, err = commandsHandler.Handle(updateUserCommand)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		response.WriteSuccessResponse(w)
	}
}

func deleteUser(commandsHandler commonapp.CommandHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := request.GetIDFromRequest(r)
		if userID != request.GetUserIDFromHeader(r) {
			response.WriteForbiddenResponse(w)
			return
		}
		removeUserCommand := command.RemoveUser{ID: request.GetRequestIDFromRequest(r), UserID: userID}
		_, err := commandsHandler.Handle(removeUserCommand)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		response.WriteSuccessResponse(w)
	}
}

func addUserFriend(commandsHandler commonapp.CommandHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := request.GetUserIDFromHeader(r)
		friendID := request.GetIDFromRequest(r)
		if userID == friendID {
			response.WriteSuccessResponse(w)
			return
		}
		addUserFriendCommand := command.AddUserFriend{ID: request.GetRequestIDFromRequest(r), UserID: userID, FriendID: friendID}
		_, err := commandsHandler.Handle(addUserFriendCommand)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		response.WriteSuccessResponse(w)
	}
}

func removeUserFriend(commandsHandler commonapp.CommandHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := request.GetUserIDFromHeader(r)
		friendID := request.GetIDFromRequest(r)
		addUserFriendCommand := command.RemoveUserFriend{ID: request.GetRequestIDFromRequest(r), UserID: userID, FriendID: friendID}
		_, err := commandsHandler.Handle(addUserFriendCommand)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		response.WriteSuccessResponse(w)
	}
}

func listUserFriends(service app.UserQueryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDStr := request.GetIDFromRequest(r)
		userID, err := uuid.FromString(userIDStr)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		friends, err := service.ListUserFriends(userID)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		result := make([]userresponse.Friend, len(friends))
		for _, friend := range friends {
			result = append(result, userresponse.Friend{ID: friend.ID.String(), Username: friend.Username})
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

func listUsers(service app.UserQueryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var listUsersRequest userrequest.ListUsers
		err := decoder.Decode(&listUsersRequest)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		uuids, err := uuid.FromStrings(listUsersRequest.UserIds)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		users, err := service.ListUsers(uuids)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		result := make([]userresponse.ListItemDTO, len(users))
		for _, user := range users {
			result = append(result, userresponse.ListItemDTO{
				ID:       user.ID.String(),
				Username: user.Username,
				IsFriend: user.IsFriend,
			})
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

func convertToUser(user *app.UserProfileDTO) userresponse.Data {
	return userresponse.Data{
		ID:        user.ID.String(),
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Age:       user.Age,
		Sex:       user.Sex,
		Password:  user.Password,
		Interests: user.Interests,
		City:      user.City,
	}
}

func writeUserResponse(id string, w http.ResponseWriter) {
	data, err := json.Marshal(userresponse.User{UserID: id})
	if err != nil {
		response.WriteErrorResponse(err, w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func writeGetUserResponse(user *app.UserProfileDTO, w http.ResponseWriter) {
	if user == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	userResponse := userresponse.Data{
		ID:        user.ID.String(),
		Username:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Age:       user.Age,
		Sex:       user.Sex,
		Password:  user.Password,
		Interests: user.Interests,
		City:      user.City,
	}
	data, err := json.Marshal(userResponse)
	if err != nil {
		response.WriteErrorResponse(err, w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func getRegisterUserRequest(r *http.Request) (userrequest.RegisterUser, error) {
	decoder := json.NewDecoder(r.Body)
	var registerUserRequest userrequest.RegisterUser
	err := decoder.Decode(&registerUserRequest)
	return registerUserRequest, err
}

func getUpdateUserRequest(r *http.Request) (userrequest.UpdateUser, error) {
	decoder := json.NewDecoder(r.Body)
	var updateUserRequest userrequest.UpdateUser
	err := decoder.Decode(&updateUserRequest)
	return updateUserRequest, err
}

func processError(err error, w http.ResponseWriter) {
	if err == app.ErrInvalidUserAge || err == app.ErrInvalidUserSex {
		response.WriteBadRequestResponse(err, w)
		return
	}
	if err == commonapp.ErrCommandAlreadyProcessed {
		response.WriteDuplicateRequestResponse(err, w)
		return
	}
	response.WriteErrorResponse(err, w)
}
