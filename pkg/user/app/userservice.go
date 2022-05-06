package app

import (
	"github.com/callicoder/go-docker/pkg/common/app/event"
	"github.com/callicoder/go-docker/pkg/common/uuid"
	"github.com/pkg/errors"
)

const (
	Male int = iota
	Female
)

var ErrUserAlreadyExists = errors.New("User already exists")
var ErrInvalidUserSex = errors.New("Invalid user sex")
var ErrInvalidUserAge = errors.New("Invalid user age")

type User struct {
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

type UserFriend struct {
	UserID   uuid.UUID
	FriendID uuid.UUID
}

type UserRepository interface {
	NewID() uuid.UUID
	GetUserByNameAndPassword(userName, password string) (*User, error)
	GetUserByName(userName string) (*User, error)
	Delete(id uuid.UUID) error
	GetUser(id uuid.UUID) (*User, error)
	Store(user User) error
}

type UserFriendRepository interface {
	AddFriend(userFriend *UserFriend) error
	RemoveFriend(userFriend *UserFriend) error
}

type UserService interface {
	RegisterUser(username, firstName, lastName, interests, city, password string, age, sex int) (uuid.UUID, error)
	DeleteUser(id uuid.UUID) error
	UpdateUser(id uuid.UUID, username, firstName, lastName, interests, city, password string, age, sex int) (*User, error)
	AddUserFriend(userID, friendID uuid.UUID) error
	RemoveUserFriend(userID, friendID uuid.UUID) error
}

type service struct {
	userRepository       UserRepository
	userFriendRepository UserFriendRepository
	eventDispatcher      event.Dispatcher
	passwordEncoder      PasswordEncoder
}

func (s service) RegisterUser(username, firstName, lastName, interests, city, password string, age, sex int) (uuid.UUID, error) {
	if sex != Male && sex != Female {
		return uuid.Nil, ErrInvalidUserSex
	}
	if age <= 0 {
		return uuid.Nil, ErrInvalidUserAge
	}
	duplicatedUser, err := s.userRepository.GetUserByName(username)
	if err != nil {
		return uuid.Nil, errors.WithStack(err)
	}
	if duplicatedUser != nil {
		return uuid.Nil, ErrUserAlreadyExists
	}
	user := User{
		ID:        s.userRepository.NewID(),
		Username:  username,
		FirstName: firstName,
		LastName:  lastName,
		Age:       age,
		Sex:       sex,
		Interests: interests,
		City:      city,
		Password:  s.passwordEncoder.Encode(password),
	}
	err = s.userRepository.Store(user)
	if err != nil {
		return uuid.Nil, errors.WithStack(err)
	}
	err = s.eventDispatcher.Dispatch(event.UserCreated{UserID: user.ID.String(), Username: user.Username})
	if err != nil {
		return uuid.Nil, errors.WithStack(err)
	}
	return user.ID, nil
}

func (s service) DeleteUser(id uuid.UUID) error {
	err := s.userRepository.Delete(id)
	if err != nil {
		return errors.WithStack(err)
	}
	return s.eventDispatcher.Dispatch(event.UserRemoved{UserID: id.String()})
}

func (s service) UpdateUser(id uuid.UUID, username, firstName, lastName, interests, city, password string, age, sex int) (*User, error) {
	if sex != Male && sex != Female {
		return nil, ErrInvalidUserSex
	}
	if age <= 0 {
		return nil, ErrInvalidUserAge
	}
	user, err := s.userRepository.GetUser(id)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if user == nil {
		return nil, nil
	}
	if user.Username != username {
		duplicatedUser, err1 := s.userRepository.GetUserByName(username)
		if err1 != nil {
			return nil, errors.WithStack(err1)
		}
		if duplicatedUser != nil {
			return nil, ErrUserAlreadyExists
		}
	}
	user.Username = username
	user.FirstName = firstName
	user.LastName = lastName
	user.Interests = interests
	user.Age = age
	user.Password = s.passwordEncoder.Encode(password)
	user.Sex = sex
	user.City = city
	err = s.userRepository.Store(*user)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	err = s.eventDispatcher.Dispatch(event.UserUpdated{UserID: user.ID.String(), Username: user.Username})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return user, nil
}

func (s service) AddUserFriend(userID, friendID uuid.UUID) error {
	userFriend := UserFriend{
		UserID:   userID,
		FriendID: friendID,
	}
	err := s.userFriendRepository.AddFriend(&userFriend)
	if err != nil {
		return errors.WithStack(err)
	}
	err = s.eventDispatcher.Dispatch(event.UserFriendAdded{UserID: userID.String(), FriendID: friendID.String()})
	return errors.WithStack(err)
}

func (s service) RemoveUserFriend(userID, friendID uuid.UUID) error {
	userFriend := UserFriend{
		UserID:   userID,
		FriendID: friendID,
	}
	err := s.userFriendRepository.RemoveFriend(&userFriend)
	if err != nil {
		return errors.WithStack(err)
	}
	err = s.eventDispatcher.Dispatch(event.UserFriendRemoved{UserID: userID.String(), FriendID: friendID.String()})
	return errors.WithStack(err)
}

func NewUserService(userRepository UserRepository, userFriendRepository UserFriendRepository, eventDispatcher event.Dispatcher) UserService {
	return &service{
		userRepository:       userRepository,
		userFriendRepository: userFriendRepository,
		eventDispatcher:      eventDispatcher,
		passwordEncoder:      NewPasswordEncoder(),
	}
}
