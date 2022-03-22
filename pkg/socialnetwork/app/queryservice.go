package app

import (
	"github.com/callicoder/go-docker/pkg/common/uuid"
)

type UserProfileDTO struct {
	ID        uuid.UUID
	Username  string
	FirstName string
	LastName  string
	Age       int
	Sex       int
	Interests string
	City      string
	Password  string
}

type UserListItemDTO struct {
	ID       uuid.UUID
	Username string
	IsFriend bool
}

type UserFriendDTO struct {
	ID       uuid.UUID
	Username string
}

type UserQueryService interface {
	GetUserByNameAndPassword(userName, password string) (*UserProfileDTO, error)
	GetUserProfile(id uuid.UUID) (*UserProfileDTO, error)
	ListUsers() ([]*UserListItemDTO, error)
	ListUserFriends(userID uuid.UUID) ([]*UserFriendDTO, error)
}
