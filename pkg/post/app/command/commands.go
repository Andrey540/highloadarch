package command

const (
	CreatePostCommand = "post.create"
)

type CreatePost struct {
	ID       string `json:"id"`
	AuthorID string `json:"authorID"`
	Title    string `json:"title"`
	Text     string `json:"text"`
}

func (command CreatePost) CommandType() string {
	return CreatePostCommand
}

func (command CreatePost) CommandID() string {
	return command.ID
}
