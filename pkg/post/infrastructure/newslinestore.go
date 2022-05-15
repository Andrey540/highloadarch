package infrastructure

import (
	"github.com/callicoder/go-docker/pkg/common/infrastructure/mysql"
	"github.com/callicoder/go-docker/pkg/post/app"
	"github.com/pkg/errors"
)

type newsLineStore struct {
	client mysql.Client
}

func (s newsLineStore) AddNewPost(post app.NewPost) error {
	sqlQuery := `INSERT INTO news_line (user_id, post_id, author_id, title)
				 SELECT uuid_from_bin(user_id), ?, ?, ?
				 FROM user_friend
				 WHERE friend_id = ?`
	_, err := s.client.Exec(sqlQuery, post.ID.String(), post.AuthorID.String(), post.Title, mysql.BinaryUUID(post.AuthorID))
	return errors.WithStack(err)
}

func NewNewsLineStore(client mysql.Client) app.NewsLineStore {
	return &newsLineStore{
		client: client,
	}
}
