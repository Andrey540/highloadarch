package inrastructure

import (
	"context"
	"encoding/json"
	"net/http"

	api "github.com/callicoder/go-docker/pkg/common/api"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/request"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

type RegisterUser struct {
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Age       int    `json:"age"`
	Sex       int    `json:"sex"`
	Interests string `json:"interests"`
	City      string `json:"city"`
	Password  string `json:"password"`
}

type UpdateUser struct {
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Age       int    `json:"age"`
	Sex       int    `json:"sex"`
	Interests string `json:"interests"`
	City      string `json:"city"`
	Password  string `json:"password"`
}

type Auth struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

type UserService struct {
	client api.UserClient
}

func (s UserService) AuthUser(r *http.Request) (*api.SignInResponse, error) {
	decoder := json.NewDecoder(r.Body)
	var authRequest Auth
	err := decoder.Decode(&authRequest)
	if err != nil {
		return nil, err
	}
	req := &api.SignInRequest{
		UserName: authRequest.UserName,
		Password: authRequest.Password,
	}
	ctx := getGRPCContext(context.Background(), r)
	return s.client.SignIn(ctx, req)
}

func (s UserService) RegisterUser(r *http.Request) (*api.RegisterUserResponse, error) {
	decoder := json.NewDecoder(r.Body)
	var registerRequest RegisterUser
	err := decoder.Decode(&registerRequest)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	req := &api.RegisterUserRequest{
		UserName:  registerRequest.Username,
		FirstName: registerRequest.FirstName,
		LastName:  registerRequest.LastName,
		Age:       uint32(registerRequest.Age),
		Sex:       uint32(registerRequest.Sex),
		Interests: registerRequest.Interests,
		City:      registerRequest.City,
		Password:  registerRequest.Password,
	}
	ctx := getGRPCContext(context.Background(), r)
	return s.client.RegisterUser(ctx, req)
}

func (s UserService) GetMyProfile(r *http.Request) (*api.UserData, error) {
	loggedUserID := GetUserIDFromContext(r)
	req := &api.GetProfileRequest{
		UserID: loggedUserID,
	}
	ctx := getGRPCContext(context.Background(), r)
	resp, err := s.client.GetProfile(ctx, req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return resp.User, nil
}

func (s UserService) GetUserProfile(r *http.Request) (*api.UserData, error) {
	userID := request.GetIDFromRequest(r)
	req := &api.GetProfileRequest{
		UserID: userID,
	}
	ctx := getGRPCContext(context.Background(), r)
	resp, err := s.client.GetProfile(ctx, req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return resp.User, nil
}

func (s UserService) ListMyFriends(r *http.Request) ([]*api.Friend, error) {
	loggedUserID := GetUserIDFromContext(r)
	req := &api.ListFriendsRequest{
		UserID: loggedUserID,
	}
	ctx := getGRPCContext(context.Background(), r)
	resp, err := s.client.ListFriends(ctx, req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return resp.Friends, nil
}

func (s UserService) ListUserFriends(r *http.Request) ([]*api.Friend, error) {
	userID := request.GetIDFromRequest(r)
	req := &api.ListFriendsRequest{
		UserID: userID,
	}
	ctx := getGRPCContext(context.Background(), r)
	resp, err := s.client.ListFriends(ctx, req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return resp.Friends, nil
}

func (s UserService) ListUsers(r *http.Request, ids []string) ([]*api.UserListItem, error) {
	req := &api.ListUsersRequest{
		UserIDs: ids,
	}
	ctx := getGRPCContext(context.Background(), r)
	resp, err := s.client.ListUsers(ctx, req)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return resp.Users, nil
}

func NewUserService(conn grpc.ClientConnInterface) UserService {
	return UserService{
		client: api.NewUserClient(conn),
	}
}
