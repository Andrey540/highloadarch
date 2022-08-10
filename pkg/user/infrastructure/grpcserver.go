package infrastructure

import (
	api "github.com/callicoder/go-docker/pkg/common/api"
	commonapp "github.com/callicoder/go-docker/pkg/common/app"
	commonserver "github.com/callicoder/go-docker/pkg/common/infrastructure/server"
	"github.com/callicoder/go-docker/pkg/common/uuid"
	"github.com/callicoder/go-docker/pkg/user/app"
	"github.com/callicoder/go-docker/pkg/user/app/command"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/types/known/emptypb"

	"context"
)

type server struct {
	queryService    app.UserQueryService
	commandsHandler commonapp.CommandHandler
}

// NewGRPCServer creates GRPC server which accesses to Service
func NewGRPCServer(queryService app.UserQueryService, commandsHandler commonapp.CommandHandler) api.UserServer {
	return &server{
		queryService:    queryService,
		commandsHandler: commandsHandler,
	}
}

func (s *server) SignIn(_ context.Context, request *api.SignInRequest) (*api.SignInResponse, error) {
	user, err := s.queryService.GetUserByNameAndPassword(request.UserName, request.Password)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.WithStack(commonapp.ErrNotFound)
	}
	return &api.SignInResponse{UserID: user.ID.String()}, err
}

func (s *server) RegisterUser(ctx context.Context, request *api.RegisterUserRequest) (*api.RegisterUserResponse, error) {
	registerUserCommand := command.RegisterUser{
		ID:        commonserver.GetRequestIDFromGRPCMetadata(ctx),
		Username:  request.UserName,
		FirstName: request.FirstName,
		LastName:  request.LastName,
		Age:       int(request.Age),
		Sex:       int(request.Sex),
		Interests: request.Interests,
		City:      request.City,
		Password:  request.Password,
	}
	id, err := s.commandsHandler.Handle(registerUserCommand)
	if err != nil {
		return nil, err
	}
	return &api.RegisterUserResponse{UserID: id.(uuid.UUID).String()}, err
}

func (s *server) GetProfile(ctx context.Context, request *api.GetProfileRequest) (*api.GetProfileResponse, error) {
	err := commonserver.Authenticate(ctx)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.FromString(request.UserID)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	user, err := s.queryService.GetUserProfile(userID)
	if err != nil || user == nil {
		return nil, err
	}
	return &api.GetProfileResponse{User: &api.UserData{
		Id:        user.ID.String(),
		UserName:  user.Username,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Age:       uint32(user.Age),
		Sex:       uint32(user.Sex),
		Interests: user.Interests,
		City:      user.City,
		Password:  user.Password,
	}}, err
}

func (s *server) UpdateUser(ctx context.Context, request *api.UpdateUserRequest) (*emptypb.Empty, error) {
	err := commonserver.Authenticate(ctx)
	if err != nil {
		return nil, err
	}
	updateUserCommand := command.UpdateUser{
		ID:        commonserver.GetRequestIDFromGRPCMetadata(ctx),
		UserID:    commonserver.GetUserIDFromGRPCMetadata(ctx),
		Username:  request.UserName,
		FirstName: request.FirstName,
		LastName:  request.LastName,
		Age:       int(request.Age),
		Sex:       int(request.Sex),
		Interests: request.Interests,
		City:      request.City,
		Password:  request.Password,
	}
	_, err = s.commandsHandler.Handle(updateUserCommand)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, err
}

func (s *server) DeleteUser(ctx context.Context, request *api.DeleteUserRequest) (*emptypb.Empty, error) {
	err := commonserver.Authenticate(ctx)
	if err != nil {
		return nil, err
	}
	removeUserCommand := command.RemoveUser{
		ID:     commonserver.GetRequestIDFromGRPCMetadata(ctx),
		UserID: request.UserID,
	}
	_, err = s.commandsHandler.Handle(removeUserCommand)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, err
}

func (s *server) FindProfiles(ctx context.Context, request *api.FindProfilesRequest) (*api.FindProfilesResponse, error) {
	err := commonserver.Authenticate(ctx)
	if err != nil {
		return nil, err
	}
	users, err := s.queryService.ListUserProfiles(request.UserName)
	if err != nil {
		return nil, err
	}
	usersData := []*api.UserData{}
	for _, user := range users {
		usersData = append(usersData, &api.UserData{
			Id:        user.ID.String(),
			UserName:  user.Username,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Age:       uint32(user.Age),
			Sex:       uint32(user.Sex),
			Interests: user.Interests,
			City:      user.City,
			Password:  user.Password,
		})
	}
	return &api.FindProfilesResponse{Users: usersData}, err
}

func (s *server) AddFriend(ctx context.Context, request *api.AddFriendRequest) (*emptypb.Empty, error) {
	err := commonserver.Authenticate(ctx)
	if err != nil {
		return nil, err
	}
	addFriendCommand := command.AddUserFriend{
		ID:       commonserver.GetRequestIDFromGRPCMetadata(ctx),
		UserID:   commonserver.GetUserIDFromGRPCMetadata(ctx),
		FriendID: request.UserID,
	}
	_, err = s.commandsHandler.Handle(addFriendCommand)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, err
}

func (s *server) RemoveFriend(ctx context.Context, request *api.RemoveFriendRequest) (*emptypb.Empty, error) {
	err := commonserver.Authenticate(ctx)
	if err != nil {
		return nil, err
	}
	removeFriendCommand := command.RemoveUserFriend{
		ID:       commonserver.GetRequestIDFromGRPCMetadata(ctx),
		UserID:   commonserver.GetUserIDFromGRPCMetadata(ctx),
		FriendID: request.UserID,
	}
	_, err = s.commandsHandler.Handle(removeFriendCommand)
	if err != nil {
		return nil, err
	}
	return &emptypb.Empty{}, err
}

func (s *server) ListFriends(ctx context.Context, request *api.ListFriendsRequest) (*api.ListFriendsResponse, error) {
	err := commonserver.Authenticate(ctx)
	if err != nil {
		return nil, err
	}
	userID, err := uuid.FromString(request.UserID)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	friends, err := s.queryService.ListUserFriends(userID)
	if err != nil {
		return nil, err
	}
	friendsData := []*api.Friend{}
	for _, friend := range friends {
		friendsData = append(friendsData, &api.Friend{
			UserID:   friend.ID.String(),
			UserName: friend.Username,
		})
	}
	return &api.ListFriendsResponse{Friends: friendsData}, err
}

func (s *server) ListUsers(ctx context.Context, request *api.ListUsersRequest) (*api.ListUsersResponse, error) {
	err := commonserver.Authenticate(ctx)
	if err != nil {
		return nil, err
	}
	uuids, err := uuid.FromStrings(request.UserIDs)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	users, err := s.queryService.ListUsers(uuids)
	if err != nil {
		return nil, err
	}
	usersData := []*api.UserListItem{}
	for _, user := range users {
		usersData = append(usersData, &api.UserListItem{
			UserID:   user.ID.String(),
			UserName: user.Username,
			IsFriend: user.IsFriend,
		})
	}
	return &api.ListUsersResponse{Users: usersData}, err
}
