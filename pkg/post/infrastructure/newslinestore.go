package infrastructure

import (
	"fmt"

	"github.com/callicoder/go-docker/pkg/common/infrastructure/mysql"
	"github.com/callicoder/go-docker/pkg/post/app"
	"github.com/pkg/errors"
)

const MaxNewsCount = 1000

type newsLineStore struct {
	client mysql.Client
}

func (s newsLineStore) AddNewPost(post app.NewPost) error {
	sqlQuery := `INSERT INTO news_line (user_id, post_id)
				 SELECT user_id, ?
				 FROM user_friend
				 WHERE friend_id = ?`
	fmt.Println(sqlQuery, post.ID.String(), post.AuthorID.String())
	_, err := s.client.Exec(sqlQuery, mysql.BinaryUUID(post.ID), mysql.BinaryUUID(post.AuthorID))
	if err != nil {
		return errors.WithStack(err)
	}
	sqlQuery = `UPDATE news_line nl
                INNER JOIN user_friend uf ON uf.user_id = nl.user_id
				SET nl.post_index = nl.post_index + 1
				WHERE uf.friend_id = ?`
	_, err = s.client.Exec(sqlQuery, mysql.BinaryUUID(post.AuthorID))
	if err != nil {
		return errors.WithStack(err)
	}
	sqlQuery = `DELETE FROM news_line WHERE post_index > ?`
	_, err = s.client.Exec(sqlQuery, MaxNewsCount)
	return errors.WithStack(err)
}

func NewNewsLineStore(client mysql.Client) app.NewsLineStore {
	return &newsLineStore{
		client: client,
	}
}
