package post

type Post struct {
	PostID string `json:"post_id"`
}

type NewsListItem struct {
	ID       string `json:"id"`
	AuthorID string `json:"authorID"`
	Title    string `json:"title"`
}

type Data struct {
	ID       string `json:"id"`
	AuthorID string `json:"authorID"`
	Title    string `json:"title"`
	Text     string `json:"text"`
}
