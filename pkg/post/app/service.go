package app

import (
	"github.com/callicoder/go-docker/pkg/common/app/event"
	"github.com/callicoder/go-docker/pkg/common/uuid"
	"github.com/pkg/errors"
)

type Post struct {
	ID       uuid.UUID
	AuthorID uuid.UUID
	Title    string
	Text     string
}

type NewPost struct {
	ID       uuid.UUID
	AuthorID uuid.UUID
	Title    string
}

type UserFriend struct {
	UserID   uuid.UUID
	FriendID uuid.UUID
}

type User struct {
	UserID   uuid.UUID
	Username string
}

type UserNotifier interface {
	Notify(userIDs []uuid.UUID, postID uuid.UUID, author, title string) error
}

type UserProvider interface {
	ListUserSubscribers(userID uuid.UUID) ([]uuid.UUID, error)
}

type PostRepository interface {
	NewID() uuid.UUID
	Store(post Post) error
}

type UserFriendRepository interface {
	AddFriend(userFriend *UserFriend) error
	RemoveFriend(userFriend *UserFriend) error
}

type UserRepository interface {
	AddUser(user *User) error
	RemoveUser(userID uuid.UUID) error
	GetUser(userID uuid.UUID) (*User, error)
}

type NewsLineStore interface {
	AddNewPost(post NewPost) error
}

type NewsLineCache interface {
	InvalidateUsers(userIDs []uuid.UUID) error
}

type PostService interface {
	CreatePost(authorID uuid.UUID, title, text string) (uuid.UUID, error)
	AddNewPost(postID, authorID uuid.UUID, title string) error
}

type UserService interface {
	AddUser(userID uuid.UUID, username string) error
	RemoveUser(userID uuid.UUID) error
	GetUserName(userID uuid.UUID) (string, error)
	AddUserFriend(userID, friendID uuid.UUID) error
	RemoveUserFriend(userID, friendID uuid.UUID) error
}

type postService struct {
	postRepository  PostRepository
	newsLineStore   NewsLineStore
	eventDispatcher event.Dispatcher
}

func (s postService) CreatePost(authorID uuid.UUID, title, text string) (uuid.UUID, error) {
	id := s.postRepository.NewID()
	post := Post{
		ID:       id,
		AuthorID: authorID,
		Title:    title,
		Text:     text,
	}
	err := s.postRepository.Store(post)
	if err != nil {
		return uuid.Nil, errors.WithStack(err)
	}
	err = s.eventDispatcher.Dispatch(event.PostCreated{PostID: id.String(), AuthorID: authorID.String(), Title: title, Text: text})
	if err != nil {
		return uuid.Nil, errors.WithStack(err)
	}
	return id, nil
}

func (s postService) AddNewPost(postID, authorID uuid.UUID, title string) error {
	return s.newsLineStore.AddNewPost(NewPost{ID: postID, AuthorID: authorID, Title: title})
}

func NewPostService(postRepository PostRepository, newsLineStore NewsLineStore, eventDispatcher event.Dispatcher) PostService {
	return &postService{
		postRepository:  postRepository,
		newsLineStore:   newsLineStore,
		eventDispatcher: eventDispatcher,
	}
}

type userService struct {
	userFriendRepository UserFriendRepository
	userRepository       UserRepository
}

func (s userService) AddUser(userID uuid.UUID, username string) error {
	user := User{
		UserID:   userID,
		Username: username,
	}
	return s.userRepository.AddUser(&user)
}

func (s userService) RemoveUser(userID uuid.UUID) error {
	return s.userRepository.RemoveUser(userID)
}

func (s userService) GetUserName(userID uuid.UUID) (string, error) {
	user, err := s.userRepository.GetUser(userID)
	if user == nil || err != nil {
		return "", err
	}
	return user.Username, nil
}

func (s userService) AddUserFriend(userID, friendID uuid.UUID) error {
	userFriend := UserFriend{
		UserID:   userID,
		FriendID: friendID,
	}
	return s.userFriendRepository.AddFriend(&userFriend)
}

func (s userService) RemoveUserFriend(userID, friendID uuid.UUID) error {
	userFriend := UserFriend{
		UserID:   userID,
		FriendID: friendID,
	}
	return s.userFriendRepository.RemoveFriend(&userFriend)
}

func NewUserService(userFriendRepository UserFriendRepository, userRepository UserRepository) UserService {
	return &userService{
		userFriendRepository: userFriendRepository, userRepository: userRepository,
	}
}
