package infrastructure

import (
	api "github.com/callicoder/go-docker/pkg/common/api"
	commonapp "github.com/callicoder/go-docker/pkg/common/app"
	commonserver "github.com/callicoder/go-docker/pkg/common/infrastructure/server"
	"github.com/callicoder/go-docker/pkg/common/uuid"
	"github.com/callicoder/go-docker/pkg/post/app"
	"github.com/callicoder/go-docker/pkg/post/app/command"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/pkg/errors"

	"context"
)

type server struct {
	queryService    app.NewsLineQueryService
	commandsHandler commonapp.CommandHandler
}

// NewGRPCServer creates GRPC server which accesses to Service
func NewGRPCServer(queryService app.NewsLineQueryService, commandsHandler commonapp.CommandHandler) api.PostServer {
	return &server{
		queryService:    queryService,
		commandsHandler: commandsHandler,
	}
}

func (s *server) CreatePost(ctx context.Context, request *api.CreatePostRequest) (*api.CreatePostResponse, error) {
	err := commonserver.Authenticate(ctx)
	if err != nil {
		return nil, err
	}
	userID := commonserver.GetUserIDFromGRPCMetadata(ctx)
	createPostCommand := command.CreatePost{
		ID:       commonserver.GetRequestIDFromGRPCMetadata(ctx),
		AuthorID: userID,
		Title:    request.Title,
		Text:     request.Text,
	}
	id, err := s.commandsHandler.Handle(createPostCommand)
	if err != nil {
		return nil, err
	}
	return &api.CreatePostResponse{PostID: id.(uuid.UUID).String()}, err
}

func (s *server) ListPosts(ctx context.Context, _ *empty.Empty) (*api.ListPostsResponse, error) {
	err := commonserver.Authenticate(ctx)
	if err != nil {
		return nil, err
	}
	userID := commonserver.GetUserIDFromGRPCMetadata(ctx)
	userUID, err := uuid.FromString(userID)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	posts, err := s.queryService.ListPosts(userUID)
	if err != nil {
		return nil, err
	}
	result := make([]*api.PostItem, 0, len(posts))
	for _, post := range posts {
		result = append(result, &api.PostItem{Id: post.ID.String(), AuthorID: post.Author.String(), Title: post.Title, Text: post.Text})
	}
	return &api.ListPostsResponse{Posts: result}, nil
}

func (s *server) GetPost(ctx context.Context, request *api.GetPostRequest) (*api.GetPostResponse, error) {
	err := commonserver.Authenticate(ctx)
	if err != nil {
		return nil, err
	}
	postID, err := uuid.FromString(request.PostID)
	if err != nil {
		return nil, err
	}
	post, err := s.queryService.GetPost(postID)
	if err != nil {
		return nil, err
	}
	return &api.GetPostResponse{Post: &api.PostItem{Id: post.ID.String(), AuthorID: post.Author.String(), Title: post.Title, Text: post.Text}}, nil
}

func (s *server) ListNews(ctx context.Context, _ *empty.Empty) (*api.ListNewsResponse, error) {
	err := commonserver.Authenticate(ctx)
	if err != nil {
		return nil, err
	}
	userID := commonserver.GetUserIDFromGRPCMetadata(ctx)
	userUID, err := uuid.FromString(userID)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	news, err := s.queryService.ListNews(userUID)
	if err != nil {
		return nil, err
	}
	result := make([]*api.NewsItem, 0, len(news))
	for _, newsItem := range news {
		result = append(result, &api.NewsItem{Id: newsItem.ID, AuthorID: newsItem.Author, Title: newsItem.Title})
	}
	return &api.ListNewsResponse{News: result}, nil
}
