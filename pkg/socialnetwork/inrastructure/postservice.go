package inrastructure

import (
	"context"
	"net/http"

	api "github.com/callicoder/go-docker/pkg/common/api"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/request"
	"github.com/golang/protobuf/ptypes/empty"
	"google.golang.org/grpc"
)

type PostService struct {
	client api.PostClient
}

func (s PostService) ListPosts(r *http.Request) ([]*api.PostItem, error) {
	req := &empty.Empty{}
	ctx := getGRPCContext(context.Background(), r)
	res, err := s.client.ListPosts(ctx, req)
	if err != nil {
		return []*api.PostItem{}, err
	}
	return res.Posts, nil
}

func (s PostService) ListNews(r *http.Request) ([]*api.NewsItem, error) {
	req := &empty.Empty{}
	ctx := getGRPCContext(context.Background(), r)
	res, err := s.client.ListNews(ctx, req)
	if err != nil {
		return []*api.NewsItem{}, err
	}
	return res.News, nil
}

func (s PostService) GetPost(r *http.Request) (*api.PostItem, error) {
	postID := request.GetIDFromRequest(r)
	req := &api.GetPostRequest{
		PostID: postID,
	}
	ctx := getGRPCContext(context.Background(), r)
	res, err := s.client.GetPost(ctx, req)
	if err != nil {
		return nil, err
	}
	return res.Post, nil
}

func NewPostService(conn grpc.ClientConnInterface) PostService {
	return PostService{
		client: api.NewPostClient(conn),
	}
}
