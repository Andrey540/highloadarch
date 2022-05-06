package infrastructure

import (
	"github.com/callicoder/go-docker/pkg/common/infrastructure/mysql"
	"github.com/callicoder/go-docker/pkg/common/uuid"
	"github.com/callicoder/go-docker/pkg/post/app"
	"github.com/pkg/errors"
	satoriuuid "github.com/satori/go.uuid"
)

type postRepository struct {
	client mysql.Client
}

func (r postRepository) NewID() uuid.UUID {
	return uuid.UUID(satoriuuid.NewV1())
}

func (r postRepository) Store(post app.Post) error {
	sqlQuery := `INSERT INTO post (id, author_id, title, text) VALUES(?, ?, ?, ?);`
	_, err := r.client.Exec(sqlQuery, mysql.BinaryUUID(post.ID), mysql.BinaryUUID(post.AuthorID), post.Title, post.Text)
	return errors.WithStack(err)
}

func NewPostRepository(client mysql.Client) app.PostRepository {
	return &postRepository{
		client: client,
	}
}
