package inrastructure

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/callicoder/go-docker/pkg/common/infrastructure/httpclient"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/request"
	userrequest "github.com/callicoder/go-docker/pkg/common/infrastructure/request/user"
	userresponse "github.com/callicoder/go-docker/pkg/common/infrastructure/response/user"
)

type UserService struct {
	baseURL string
	client  httpclient.HTTPClient
}

func (s UserService) AuthUser(r *http.Request) (userresponse.User, error) {
	decoder := json.NewDecoder(r.Body)
	var authRequest userrequest.Auth
	err := decoder.Decode(&authRequest)
	if err != nil {
		return userresponse.User{}, err
	}
	var response userresponse.User
	err = s.client.MakeJSONRequest(authRequest, &response, http.MethodPost, s.baseURL+userrequest.SignInURL, nil)
	return response, err
}

func (s UserService) RegisterUser(r *http.Request) (userresponse.User, error) {
	decoder := json.NewDecoder(r.Body)
	var registerRequest userrequest.RegisterUser
	err := decoder.Decode(&registerRequest)
	if err != nil {
		return userresponse.User{}, err
	}
	var response userresponse.User
	err = s.client.MakeJSONRequest(registerRequest, &response, http.MethodPost, s.baseURL+userrequest.RegisterURL, nil)
	return response, err
}

func (s UserService) GetMyProfile(r *http.Request) (userresponse.Data, error) {
	var response userresponse.Data
	headers := getHeaders(r)
	loggedUserID := getUserIDFromContext(r)
	url := strings.ReplaceAll(s.baseURL+userrequest.ProfileURL, "{id}", loggedUserID)
	err := s.client.MakeJSONRequest(nil, &response, http.MethodGet, url, headers)
	return response, err
}

func (s UserService) GetUserProfile(r *http.Request) (userresponse.Data, error) {
	var response userresponse.Data
	headers := getHeaders(r)
	userID := request.GetIDFromRequest(r)
	url := strings.ReplaceAll(s.baseURL+userrequest.ProfileURL, "{id}", userID)
	err := s.client.MakeJSONRequest(nil, &response, http.MethodGet, url, headers)
	return response, err
}

func (s UserService) ListMyFriends(r *http.Request) ([]userresponse.Friend, error) {
	var response []userresponse.Friend
	headers := getHeaders(r)
	loggedUserID := getUserIDFromContext(r)
	url := strings.ReplaceAll(s.baseURL+userrequest.ListUserFriendsURL, "{id}", loggedUserID)
	fmt.Println(url)
	err := s.client.MakeJSONRequest(nil, &response, http.MethodGet, url, headers)
	return response, err
}

func (s UserService) ListUserFriends(r *http.Request) ([]userresponse.Friend, error) {
	var response []userresponse.Friend
	headers := getHeaders(r)
	userID := request.GetIDFromRequest(r)
	url := strings.ReplaceAll(s.baseURL+userrequest.ListUserFriendsURL, "{id}", userID)
	err := s.client.MakeJSONRequest(nil, &response, http.MethodGet, url, headers)
	return response, err
}

func (s UserService) ListUsers(r *http.Request, ids []string) ([]userresponse.ListItemDTO, error) {
	var response []userresponse.ListItemDTO
	listUsersRequest := userrequest.ListUsers{UserIds: ids}
	headers := getHeaders(r)
	err := s.client.MakeJSONRequest(listUsersRequest, &response, http.MethodGet, s.baseURL+userrequest.ListUsersURL, headers)
	return response, err
}

func NewUserService(baseURL string, client httpclient.HTTPClient) UserService {
	return UserService{
		client:  client,
		baseURL: baseURL,
	}
}
