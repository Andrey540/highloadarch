package main

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	stdlog "log"
	"net/http"
	"time"

	commonapp "github.com/callicoder/go-docker/pkg/common/app"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/metrics"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/mysql"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/redis"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/request"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/response"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/server"
	"github.com/callicoder/go-docker/pkg/common/uuid"
	"github.com/callicoder/go-docker/pkg/socialnetwork/app"
	"github.com/callicoder/go-docker/pkg/socialnetwork/app/command"
	"github.com/callicoder/go-docker/pkg/socialnetwork/infrastructure"
	"github.com/gorilla/mux"
)

const (
	appID       = "socialnetwork"
	sessionName = "otussid"

	signInURL   = "/v1/signin"
	registerURL = "/v1/register"

	loginPageURL     = "/app"
	registerPageURL  = "/app/register"
	myProfilePageURL = "/app/profile"
)

type UserCreatedResponse struct {
	UserID      string `json:"userId"`
	RedirectURL string `json:"redirect_url"`
}

type User struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Age       int    `json:"age"`
	Sex       int    `json:"sex"`
	Interests string `json:"interests"`
	City      string `json:"city"`
	Password  string `json:"password"`
}

type ListUserItem struct {
	ID       string
	Username string
}

type UserFriend struct {
	ID       string
	Username string
}

type UserProfilePage struct { // nolint: maligned
	IsSelfProfile bool
	Profile       User
	Friends       []UserFriend
	IsFriend      bool
}

type RegisterUserRequest struct {
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Age       int    `json:"age"`
	Sex       int    `json:"sex"`
	Interests string `json:"interests"`
	City      string `json:"city"`
	Password  string `json:"password"`
}

type UpdateUserRequest struct {
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Age       int    `json:"age"`
	Sex       int    `json:"sex"`
	Interests string `json:"interests"`
	City      string `json:"city"`
	Password  string `json:"password"`
}

type AuthRequest struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

var userCtxKey = &contextKey{"user"}

type contextKey struct {
	name string
}

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
	commonUnitOfWorkFactory := infrastructure.NewUnitOfWorkFactory(mysqlClient)

	transports := []commonapp.Transport{}
	eventDispatcher := commonapp.NewStoredEventDispatcher(commonUnitOfWorkFactory, transports, eventDispatcherErrorsCh)
	unitOfWorkFactory := infrastructure.NewNotifyingUnitOfWorkFactory(commonUnitOfWorkFactory, eventDispatcher.Activate)

	commandHandlerFactory := command.NewCommandHandlerFactory()
	commandsHandler := commonapp.NewCommandsHandler(unitOfWorkFactory, commandHandlerFactory)

	userQueryService := infrastructure.NewUserQueryService(connector.TransactionalClient())

	sessionService, err := redis.NewSessionService(&redis.Config{
		Password: cnf.RedisPassword,
		Host:     cnf.RedisHost + ":" + cnf.RedisPort,
	})
	if err != nil {
		return err
	}
	// nolint: errcheck
	defer sessionService.Stop()

	stopChan := make(chan struct{})
	server.ListenOSKillSignals(stopChan)
	serverHub := server.NewHub(stopChan)
	serveHTTP(cnf, serverHub, userQueryService, commandsHandler, sessionService, logger, errorLogger, metricsHandler)

	return serverHub.Wait()
}

func serveHTTP(config *config, serverHub *server.Hub, queryService app.UserQueryService, commandsHandler commonapp.CommandHandler,
	sessionService redis.SessionService, logger, errorLogger *stdlog.Logger,
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

		registerUserTpl := template.Must(template.ParseFiles(getTemplateFiles("/socialnetwork/data/tpl/register.page.html")...))
		viewUserTpl := template.Must(template.ParseFiles(getTemplateFiles("/socialnetwork/data/tpl/user_profile.page.html")...))
		signInTpl := template.Must(template.ParseFiles(getTemplateFiles("/socialnetwork/data/tpl/login.page.html")...))
		listUsersTpl := template.Must(template.ParseFiles(getTemplateFiles("/socialnetwork/data/tpl/user_list.page.html")...))

		router.HandleFunc(signInURL, authUser(queryService, sessionService)).Methods(http.MethodPost)
		router.HandleFunc("/v1/signout", logoutUser(sessionService)).Methods(http.MethodPost)
		router.HandleFunc(registerURL, registerUser(commandsHandler)).Methods(http.MethodPost)
		router.HandleFunc("/v1/profile/{id}", getUser(queryService)).Methods(http.MethodGet)
		router.HandleFunc("/v1/update/{id}", updateUser(commandsHandler)).Methods(http.MethodPut)
		router.HandleFunc("/v1/delete/{id}", deleteUser(commandsHandler)).Methods(http.MethodDelete)
		router.HandleFunc("/v1/user/friend/add/{id}", addUserFriend(commandsHandler)).Methods(http.MethodPost)
		router.HandleFunc("/v1/user/friend/remove/{id}", removeUserFriend(commandsHandler)).Methods(http.MethodPost)

		router.HandleFunc(loginPageURL, renderTemplate(signInTpl)).Methods(http.MethodGet)
		router.HandleFunc(registerPageURL, renderTemplate(registerUserTpl)).Methods(http.MethodGet)
		router.HandleFunc(myProfilePageURL, getMyProfile(queryService, viewUserTpl)).Methods(http.MethodGet)
		router.HandleFunc("/app/profile/{id}", getUserProfile(queryService, viewUserTpl)).Methods(http.MethodGet)
		router.HandleFunc("/app/logout", logoutUserWithRedirect(sessionService)).Methods(http.MethodGet)
		router.HandleFunc("/app/user/list", listUsers(queryService, listUsersTpl)).Methods(http.MethodGet)

		nextRequestID := func() string {
			return fmt.Sprintf("%d", time.Now().UnixNano())
		}

		metricsHandler.AddMetricsMiddleware(router)
		router.Use(server.AuthAPIMiddleware(sessionService, userCtxKey, sessionName, []string{"/v1/"},
			[]string{signInURL, registerURL}))
		router.Use(server.AuthAppMiddleware(sessionService, userCtxKey, sessionName, loginPageURL, []string{"/app/"},
			[]string{registerPageURL}))
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

func getTemplateFiles(filename string) []string {
	return []string{
		filename,
		"/socialnetwork/data/tpl/authorized.layout.html",
		"/socialnetwork/data/tpl/unauthorized.layout.html",
	}
}

func authUser(service app.UserQueryService, sessionService redis.SessionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		decoder := json.NewDecoder(r.Body)
		var authRequest AuthRequest
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
			response.WriteNotFoundResponse(err, w)
			return
		}
		session, err := sessionService.SaveSession(user.ID.String())
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		writeUserAuthResponse(session, r, w)
	}
}

func logoutUser(sessionService redis.SessionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionName)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		if cookie == nil {
			response.WriteSuccessResponse(w)
			return
		}
		err = sessionService.RemoveSession(cookie.Value)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		response.WriteSuccessResponse(w)
	}
}

func logoutUserWithRedirect(sessionService redis.SessionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionName)
		if err != nil || cookie == nil {
			response.WriteErrorResponse(err, w)
			return
		}
		err = sessionService.RemoveSession(cookie.Value)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		http.Redirect(w, r, loginPageURL, http.StatusSeeOther)
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
		writeUserCreatedResponse(id.(uuid.UUID).String(), w)
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
		writeUserResponse(user, w)
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
		userID := getUserIDFromContext(r)
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
		userID := getUserIDFromContext(r)
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

func renderTemplate(tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err := tpl.Execute(w, nil)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
	}
}

func getMyProfile(service app.UserQueryService, tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := getUserUUIDFromContext(r)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		user, err := service.GetUserProfile(userID)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		friends, err := service.ListUserFriends(userID)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		profile := UserProfilePage{
			IsSelfProfile: true,
			Profile:       convertToUser(user),
			Friends:       convertToFriends(friends),
			IsFriend:      false,
		}
		err = tpl.Execute(w, profile)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
	}
}

func getUserProfile(service app.UserQueryService, tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		loggedUserID, err := getUserUUIDFromContext(r)
		if err != nil {
			response.WriteErrorResponse(err, w)
		}
		userID, err := uuid.FromString(request.GetIDFromRequest(r))
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		user, err := service.GetUserProfile(userID)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		friends, err := service.ListUserFriends(userID)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		isUserFriend, err := isUserFriend(service, loggedUserID, userID)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		profile := UserProfilePage{
			IsSelfProfile: loggedUserID == userID,
			Profile:       convertToUser(user),
			Friends:       convertToFriends(friends),
			IsFriend:      isUserFriend,
		}
		err = tpl.Execute(w, profile)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
	}
}

func listUsers(service app.UserQueryService, tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := service.ListUsers()
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		var usersRes = make([]ListUserItem, len(users))
		for _, user := range users {
			usersRes = append(usersRes, listUserResponse(user))
		}
		err = tpl.Execute(w, usersRes)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
	}
}

func listUserResponse(user *app.UserListItemDTO) ListUserItem {
	return ListUserItem{
		ID:       user.ID.String(),
		Username: user.Username,
	}
}

func convertToUser(user *app.UserProfileDTO) User {
	return User{
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

func convertToFriends(friends []*app.UserFriendDTO) []UserFriend {
	result := make([]UserFriend, len(friends))
	for _, friend := range friends {
		result = append(result, UserFriend{ID: friend.ID.String(), Username: friend.Username})
	}
	return result
}

func writeUserAuthResponse(session string, r *http.Request, w http.ResponseWriter) {
	expiration := time.Now().Add(24 * time.Hour)
	cookie := http.Cookie{Name: sessionName, Value: session, Expires: expiration, Domain: r.Host, Path: "/"}
	http.SetCookie(w, &cookie)
	response.WriteSuccessWithRedirectResponse(myProfilePageURL, w)
}

func writeUserCreatedResponse(id string, w http.ResponseWriter) {
	data, err := json.Marshal(UserCreatedResponse{UserID: id, RedirectURL: loginPageURL})
	if err != nil {
		response.WriteErrorResponse(err, w)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func writeUserResponse(user *app.UserProfileDTO, w http.ResponseWriter) {
	if user == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	userResponse := User{
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

func getRegisterUserRequest(r *http.Request) (RegisterUserRequest, error) {
	decoder := json.NewDecoder(r.Body)
	var registerUserRequest RegisterUserRequest
	err := decoder.Decode(&registerUserRequest)
	return registerUserRequest, err
}

func getUpdateUserRequest(r *http.Request) (UpdateUserRequest, error) {
	decoder := json.NewDecoder(r.Body)
	var updateUserRequest UpdateUserRequest
	err := decoder.Decode(&updateUserRequest)
	return updateUserRequest, err
}

func isUserFriend(service app.UserQueryService, loggedUserID, userID uuid.UUID) (bool, error) {
	friends, err := service.ListUserFriends(loggedUserID)
	if err != nil {
		return false, nil
	}
	for _, friend := range friends {
		if friend.ID == userID {
			return true, nil
		}
	}
	return false, nil
}

func getUserUUIDFromContext(r *http.Request) (uuid.UUID, error) {
	ctx := r.Context()
	userSession := ctx.Value(userCtxKey).(*redis.UserSession)
	return uuid.FromString(userSession.UserID)
}

func getUserIDFromContext(r *http.Request) string {
	ctx := r.Context()
	userSession := ctx.Value(userCtxKey).(*redis.UserSession)
	return userSession.UserID
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
