package main

import (
	"context"
	"encoding/json"
	"html/template"
	stdlog "log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	api "github.com/callicoder/go-docker/pkg/common/api"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/metrics"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/realtime"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/redis"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/request"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/response"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/server"
	"github.com/callicoder/go-docker/pkg/socialnetwork/inrastructure"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	satoriuuid "github.com/satori/go.uuid"
	"google.golang.org/grpc"
)

const (
	appID       = "socialnetwork"
	sessionName = "otussid"

	signInURL   = "/api/v1/signin"
	registerURL = "/api/v1/register"

	loginPageURL     = "/app"
	registerPageURL  = "/app/register"
	myProfilePageURL = "/app/profile"
)

type UserCreatedResponse struct {
	UserID      string `json:"userId"`
	RedirectURL string `json:"redirect_url"`
}

type ListUserItem struct {
	ID       string
	Username string
}

type UserConversationItem struct {
	UserID         string
	Username       string
	UnreadMessages int
}

type UserProfilePage struct { // nolint: maligned
	IsSelfProfile bool
	Profile       UserData
	Friends       []FriendData
	IsFriend      bool
}

type UserConversationsPage struct { // nolint: maligned
	Conversations []UserConversationItem
}

type FriendData struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type UserData struct {
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

type MessageData struct {
	ID       string
	UserName string
	Text     string
}

type PostData struct {
	ID     string
	Author string
	Title  string
	Text   string
}

type NewPostData struct {
	ID     string
	Author string
	Title  string
}

type ConversationPage struct { // nolint: maligned
	ID       string
	UserName string
	Messages []MessageData
}

type MyPostsPage struct { // nolint: maligned
	Posts []PostData
}

type NewPostsPage struct { // nolint: maligned
	UserID       string
	Posts        []NewPostData
	RealtimeHost string
}

type PostPage struct { // nolint: maligned
	Post PostData
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

	eventDispatcherErrorsCh := make(chan error)
	go func() {
		for err := range eventDispatcherErrorsCh {
			errorLogger.Println(err)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userGRPCConn, err := grpc.DialContext(ctx, cnf.UserServiceGRPCAddress, grpc.WithInsecure())
	if err != nil {
		return errors.WithStack(err)
	}
	// noinspection GoUnhandledErrorResult
	defer userGRPCConn.Close()

	conversationGRPCConn, err := grpc.DialContext(ctx, cnf.ConversationServiceGRPCAddress, grpc.WithInsecure())
	if err != nil {
		return errors.WithStack(err)
	}
	// noinspection GoUnhandledErrorResult
	defer conversationGRPCConn.Close()

	counterGRPCConn, err := grpc.DialContext(ctx, cnf.CounterServiceGRPCAddress, grpc.WithInsecure())
	if err != nil {
		return errors.WithStack(err)
	}
	// noinspection GoUnhandledErrorResult
	defer counterGRPCConn.Close()

	postGRPCConn, err := grpc.DialContext(ctx, cnf.PostServiceGRPCAddress, grpc.WithInsecure())
	if err != nil {
		return errors.WithStack(err)
	}
	// noinspection GoUnhandledErrorResult
	defer postGRPCConn.Close()

	userService := inrastructure.NewUserService(userGRPCConn)
	conversationService := inrastructure.NewConversationService(conversationGRPCConn)
	counterService := inrastructure.NewCounterService(counterGRPCConn)
	postService := inrastructure.NewPostService(postGRPCConn)
	realtimeHosts, err := cnf.realtimeHosts()
	if err != nil {
		return err
	}
	realtimeClientService := realtime.NewClientService(realtimeHosts)

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
	serveHTTP(cnf, serverHub, userService, conversationService, counterService, postService, sessionService, realtimeClientService, logger, errorLogger, metricsHandler)

	return serverHub.Wait()
}

func serveHTTP(config *config, serverHub *server.Hub, userService inrastructure.UserService,
	conversationService inrastructure.ConversationService, counterService inrastructure.CounterService, postService inrastructure.PostService,
	sessionService redis.SessionService, realtimeClientService realtime.ClientService, logger, errorLogger *stdlog.Logger, metricsHandler metrics.PrometheusMetricsHandler) {
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
		conversationTpl := template.Must(template.ParseFiles(getTemplateFiles("/socialnetwork/data/tpl/conversation.page.html")...))
		conversationsTpl := template.Must(template.ParseFiles(getTemplateFiles("/socialnetwork/data/tpl/user_conversations.page.html")...))
		myPostsTpl := template.Must(template.ParseFiles(getTemplateFiles("/socialnetwork/data/tpl/my_posts.page.html")...))
		newPostsTpl := template.Must(template.ParseFiles(getTemplateFiles("/socialnetwork/data/tpl/new_posts.page.html")...))
		postTpl := template.Must(template.ParseFiles(getTemplateFiles("/socialnetwork/data/tpl/post.page.html")...))

		router.HandleFunc(signInURL, authUser(userService, sessionService)).Methods(http.MethodPost)
		router.HandleFunc("/api/v1/signout", logoutUser(sessionService)).Methods(http.MethodPost)
		router.HandleFunc(registerURL, registerUser(userService)).Methods(http.MethodPost)

		router.HandleFunc(loginPageURL, renderTemplate(signInTpl)).Methods(http.MethodGet)
		router.HandleFunc(registerPageURL, renderTemplate(registerUserTpl)).Methods(http.MethodGet)
		router.HandleFunc(myProfilePageURL, getMyProfile(userService, viewUserTpl)).Methods(http.MethodGet)
		router.HandleFunc("/app/profile/{id}", getUserProfile(userService, viewUserTpl)).Methods(http.MethodGet)
		router.HandleFunc("/app/logout", logoutUserWithRedirect(sessionService)).Methods(http.MethodGet)
		router.HandleFunc("/app/user/list", listUsers(userService, listUsersTpl)).Methods(http.MethodGet)
		router.HandleFunc("/app/conversation/user/{id}", getConversation(userService, conversationService, conversationTpl)).Methods(http.MethodGet)
		router.HandleFunc("/app/conversation/list", listConversations(userService, conversationService, counterService, conversationsTpl)).Methods(http.MethodGet)
		router.HandleFunc("/app/post/list", getMyPosts(postService, myPostsTpl)).Methods(http.MethodGet)
		router.HandleFunc("/app/post/news", getNewPosts(postService, userService, realtimeClientService, newPostsTpl)).Methods(http.MethodGet)
		router.HandleFunc("/app/post/{id}", getPost(postService, userService, postTpl)).Methods(http.MethodGet)

		router.PathPrefix("/user/api/").HandlerFunc(proxyRequest(config.UserServiceRESTAddress))
		router.PathPrefix("/conversation/api/").HandlerFunc(proxyRequest(config.ConversationServiceRESTAddress))
		router.PathPrefix("/post/api/").HandlerFunc(proxyRequest(config.PostServiceRESTAddress))
		router.PathPrefix("/counter/api/").HandlerFunc(proxyRequest(config.CounterServiceRESTAddress))

		nextRequestID := func() string {
			return satoriuuid.NewV1().String()
		}

		metricsHandler.AddMetricsMiddleware(router)
		router.Use(server.AuthAPIMiddleware(sessionService, inrastructure.UserCtxKey, sessionName, []string{"/api/"},
			[]string{signInURL, registerURL}))
		router.Use(server.AuthAppMiddleware(sessionService, inrastructure.UserCtxKey, sessionName, loginPageURL, []string{"/app/"},
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

func proxyRequest(serviceURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		proxyURL, err := url.Parse(serviceURL)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}

		loggedUserID := inrastructure.GetUserIDFromContext(r)
		requestID := server.GetRequestIDFromContext(r)

		r.Header.Set(request.UserIDHeader, loggedUserID)
		if requestID != "" {
			r.Header.Set(request.RequestIDHeader, requestID)
		}

		proxy := httputil.NewSingleHostReverseProxy(proxyURL)
		proxy.ServeHTTP(w, r)
	}
}

func authUser(userService inrastructure.UserService, sessionService redis.SessionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authResponse, err := userService.AuthUser(r)
		if err != nil {
			response.WriteUnauthorizedResponse(err.Error(), w)
			return
		}
		if authResponse == nil {
			response.WriteUnauthorizedResponse("Unauthorized", w)
			return
		}
		session, err := sessionService.SaveSession(authResponse.UserID)
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

func registerUser(userService inrastructure.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		registerResponse, err := userService.RegisterUser(r)
		if err != nil {
			response.WriteUnauthorizedResponse(err.Error(), w)
			return
		}
		writeUserCreatedResponse(registerResponse.UserID, w)
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

func getMyProfile(userService inrastructure.UserService, tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := userService.GetMyProfile(r)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		friends, err := userService.ListMyFriends(r)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		profile := UserProfilePage{
			IsSelfProfile: true,
			Profile:       getProfile(user),
			Friends:       getFriends(friends),
			IsFriend:      false,
		}
		err = tpl.Execute(w, profile)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
	}
}

func getUserProfile(userService inrastructure.UserService, tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, err := userService.GetUserProfile(r)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		friends, err := userService.ListUserFriends(r)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		loggedUserID := inrastructure.GetUserIDFromContext(r)
		userID := request.GetIDFromRequest(r)

		myFriends, err := userService.ListMyFriends(r)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		isUserFriend := isUserFriend(userID, myFriends)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		profile := UserProfilePage{
			IsSelfProfile: loggedUserID == userID,
			Profile:       getProfile(user),
			Friends:       getFriends(friends),
			IsFriend:      isUserFriend,
		}
		err = tpl.Execute(w, profile)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
	}
}

func listUsers(userService inrastructure.UserService, tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users, err := userService.ListUsers(r, []string{})
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		var result = make([]ListUserItem, len(users))
		for _, item := range users {
			result = append(result, listUserResponse(item))
		}
		err = tpl.Execute(w, result)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
	}
}

func getConversation(userService inrastructure.UserService, conversationService inrastructure.ConversationService, tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conversationID, err := conversationService.GetConversationID(r)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		messages, err := conversationService.ListMessages(r, conversationID)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		var userIDs []string
		var messageIDs []string
		for _, message := range messages {
			userIDs = append(userIDs, message.UserID)
			messageIDs = append(messageIDs, message.Id)
		}
		loggedUserID := inrastructure.GetUserIDFromContext(r)
		userIDs = append(userIDs, loggedUserID)
		usersMap, err := getUsersMap(r, userService, userIDs)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		loggedUser := usersMap[loggedUserID]
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		messagesData, err := getMessages(usersMap, messages)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		err = conversationService.ReadMessages(r, conversationID, messageIDs)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		page := ConversationPage{ID: conversationID, UserName: loggedUser.UserName, Messages: messagesData}
		err = tpl.Execute(w, page)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
	}
}

func listConversations(userService inrastructure.UserService, conversationService inrastructure.ConversationService,
	counterService inrastructure.CounterService, tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conversations, err := conversationService.ListConversations(r)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		unreadMessagesMap, err := getUnreadMessagesMap(r, counterService)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		var userIDs []string
		for _, conversation := range conversations {
			userIDs = append(userIDs, conversation.UserID)
		}
		usersMap, err := getUsersMap(r, userService, userIDs)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		conversationsData, err := getConversations(usersMap, unreadMessagesMap, conversations)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		page := UserConversationsPage{Conversations: conversationsData}
		err = tpl.Execute(w, page)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
	}
}

func getMyPosts(postService inrastructure.PostService, tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		posts, err := postService.ListPosts(r)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		postsData, err := getPosts(posts)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		page := MyPostsPage{Posts: postsData}
		err = tpl.Execute(w, page)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
	}
}

func getNewPosts(postService inrastructure.PostService, userService inrastructure.UserService, realtimeClientService realtime.ClientService, tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		loggedUserID := inrastructure.GetUserIDFromContext(r)
		posts, err := postService.ListNews(r)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		var userIDs []string
		for _, post := range posts {
			userIDs = append(userIDs, post.AuthorID)
		}
		usersMap, err := getUsersMap(r, userService, userIDs)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		postsData, err := getNews(usersMap, posts)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		page := NewPostsPage{UserID: loggedUserID, Posts: postsData, RealtimeHost: realtimeClientService.GetHost(loggedUserID)}
		err = tpl.Execute(w, page)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
	}
}

func getPost(postService inrastructure.PostService, userService inrastructure.UserService, tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		post, err := postService.GetPost(r)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		usersMap, err := getUsersMap(r, userService, []string{post.AuthorID})
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
		if _, found := usersMap[post.AuthorID]; !found {
			response.WriteNotFoundResponse(errors.New("User not found"), w)
			return
		}
		user := usersMap[post.AuthorID]
		data := PostData{ID: post.Id, Author: user.UserName, Title: post.Title, Text: post.Text}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		page := PostPage{Post: data}
		err = tpl.Execute(w, page)
		if err != nil {
			response.WriteErrorResponse(err, w)
			return
		}
	}
}

func listUserResponse(user *api.UserListItem) ListUserItem {
	return ListUserItem{
		ID:       user.UserID,
		Username: user.UserName,
	}
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

func isUserFriend(loggedUserID string, friends []*api.Friend) bool {
	for _, friend := range friends {
		if friend.UserID == loggedUserID {
			return true
		}
	}
	return false
}

func getProfile(user *api.UserData) UserData {
	return UserData{
		ID:        user.Id,
		Username:  user.UserName,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Age:       int(user.Age),
		Sex:       int(user.Sex),
		Interests: user.Interests,
		City:      user.City,
		Password:  user.Password,
	}
}

func getFriends(friends []*api.Friend) []FriendData {
	result := make([]FriendData, len(friends))
	for _, friend := range friends {
		result = append(result, FriendData{
			ID:       friend.UserID,
			Username: friend.UserName,
		})
	}
	return result
}

func getUnreadMessagesMap(r *http.Request, counterService inrastructure.CounterService) (map[string]int, error) {
	result := make(map[string]int)
	unreadMessages, err := counterService.ListUnreadMessages(r)
	if err != nil {
		return nil, err
	}
	for _, messages := range unreadMessages {
		result[messages.ConversationID] = int(messages.Count)
	}
	return result, nil
}

func getUsersMap(r *http.Request, userService inrastructure.UserService, ids []string) (map[string]*api.UserListItem, error) {
	result := make(map[string]*api.UserListItem)
	users, err := userService.ListUsers(r, ids)
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		result[user.UserID] = user
	}
	return result, nil
}

func getConversations(usersMap map[string]*api.UserListItem, unreadMessagesMap map[string]int, conversations []*api.UserConversation) ([]UserConversationItem, error) {
	var result []UserConversationItem
	for _, conversation := range conversations {
		if user, found := usersMap[conversation.UserID]; found {
			unreadMessagesCount := 0
			if count, found1 := unreadMessagesMap[conversation.Id]; found1 {
				unreadMessagesCount = count
			}
			result = append(result, UserConversationItem{UserID: conversation.UserID, Username: user.UserName, UnreadMessages: unreadMessagesCount})
		}
	}
	return result, nil
}

func getMessages(usersMap map[string]*api.UserListItem, messages []*api.Message) ([]MessageData, error) {
	var result []MessageData
	for _, message := range messages {
		if user, found := usersMap[message.UserID]; found {
			result = append(result, MessageData{ID: message.Id, UserName: user.UserName, Text: message.Text})
		}
	}
	return result, nil
}

func getPosts(posts []*api.PostItem) ([]PostData, error) {
	result := []PostData{}
	for _, post := range posts {
		result = append(result, PostData{ID: post.Id, Title: post.Title, Text: post.Text})
	}
	return result, nil
}

func getNews(usersMap map[string]*api.UserListItem, posts []*api.NewsItem) ([]NewPostData, error) {
	var result []NewPostData
	for _, post := range posts {
		if user, found := usersMap[post.AuthorID]; found {
			result = append(result, NewPostData{ID: post.Id, Author: user.UserName, Title: post.Title})
		}
	}
	return result, nil
}
