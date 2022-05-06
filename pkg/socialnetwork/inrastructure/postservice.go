package inrastructure

import (
	"net/http"
	"strings"

	"github.com/callicoder/go-docker/pkg/common/infrastructure/httpclient"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/request"
	postrequest "github.com/callicoder/go-docker/pkg/common/infrastructure/request/post"
	postresponse "github.com/callicoder/go-docker/pkg/common/infrastructure/response/post"
)

type PostService struct {
	baseURL string
	client  httpclient.HTTPClient
}

func (s PostService) ListPosts(r *http.Request) ([]postresponse.Data, error) {
	var response []postresponse.Data
	headers := getHeaders(r)
	err := s.client.MakeJSONRequest(nil, &response, http.MethodGet, s.baseURL+postrequest.ListPostsURL, headers)
	return response, err
}

func (s PostService) ListNews(r *http.Request) ([]postresponse.NewsListItem, error) {
	var response []postresponse.NewsListItem
	headers := getHeaders(r)
	err := s.client.MakeJSONRequest(nil, &response, http.MethodGet, s.baseURL+postrequest.ListNewsURL, headers)
	return response, err
}

func (s PostService) GetPost(r *http.Request) (postresponse.Data, error) {
	var response postresponse.Data
	headers := getHeaders(r)
	postID := request.GetIDFromRequest(r)
	url := strings.ReplaceAll(s.baseURL+postrequest.GetPostURL, "{id}", postID)
	err := s.client.MakeJSONRequest(nil, &response, http.MethodGet, url, headers)
	return response, err
}

func NewPostService(baseURL string, client httpclient.HTTPClient) PostService {
	return PostService{
		client:  client,
		baseURL: baseURL,
	}
}
