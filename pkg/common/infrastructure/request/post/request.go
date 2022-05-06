package post

const (
	AppID = "post"

	urlPrefix = "/" + AppID

	CreatePostURL = urlPrefix + "/api/v1/post"
	ListPostsURL  = urlPrefix + "/api/v1/post/list"
	ListNewsURL   = urlPrefix + "/api/v1/news/list"
	GetPostURL    = urlPrefix + "/api/v1/post/{id}"
)

type CreatePost struct {
	Title string `json:"title"`
	Text  string `json:"text"`
}
