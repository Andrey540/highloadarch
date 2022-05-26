package infrastructure

import (
	"github.com/callicoder/go-docker/pkg/common/infrastructure/mysql"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/tarantool"
	"github.com/callicoder/go-docker/pkg/common/uuid"
	"github.com/callicoder/go-docker/pkg/post/app"
	"github.com/pkg/errors"
)

const maxNewsCount = 1000

type newsLineQueryService struct {
	client          mysql.Client
	tarantoolClient tarantool.Client
}

type newsLineItem struct {
	ID       int
	UserID   string
	PostID   string
	AuthorID string
	Title    string
}

func (s newsLineQueryService) ListPosts(userID uuid.UUID) ([]*app.PostDTO, error) {
	const sqlQuery = `SELECT id, author_id, title, text FROM post WHERE author_id=? ORDER BY created_at LIMIT 1000`
	rows, err := s.client.Query(sqlQuery, mysql.BinaryUUID(userID))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if rows.Err() != nil {
		return nil, errors.WithStack(rows.Err())
	}
	var result []*app.PostDTO
	for rows.Next() {
		var post app.PostDTO
		err1 := rows.Scan(&post.ID, &post.Author, &post.Title, &post.Text)
		if err1 != nil {
			return []*app.PostDTO{}, errors.WithStack(err)
		}
		result = append(result, &post)
	}
	defer rows.Close()
	return result, nil
}

func (s newsLineQueryService) ListNews(userID uuid.UUID) (*[]app.NewsLineItem, error) {
	return s.listNewsTarantool(userID)
}

func (s newsLineQueryService) listNewsTarantool(userID uuid.UUID) (*[]app.NewsLineItem, error) {
	var newsItems []newsLineItem
	err := s.tarantoolClient.Select("mysqldata", "user_idx", 0, maxNewsCount, userID.String(), &newsItems)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	result := []app.NewsLineItem{}
	for _, news := range newsItems {
		result = append(result, app.NewsLineItem{ID: news.PostID, Author: news.AuthorID, Title: news.Title})
	}
	return &result, err
}

// nolint: unused
func (s newsLineQueryService) listNewsSQL(userID uuid.UUID) (*[]app.NewsLineItem, error) {
	const sqlQuery = `SELECT post_id, author_id, title FROM news_line WHERE user_id=? ORDER BY id DESC LIMIT ?`
	rows, err := s.client.Query(sqlQuery, userID.String(), maxNewsCount)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if rows.Err() != nil {
		return nil, errors.WithStack(rows.Err())
	}
	var result []app.NewsLineItem
	for rows.Next() {
		var post app.NewsLineItem
		err1 := rows.Scan(&post.ID, &post.Author, &post.Title)
		if err1 != nil {
			return nil, errors.WithStack(err)
		}
		result = append(result, post)
	}
	defer rows.Close()
	return &result, nil
}

func (s newsLineQueryService) GetPost(postID uuid.UUID) (*app.PostDTO, error) {
	const sqlQuery = `SELECT id, author_id, title, text FROM post WHERE id=?`
	rows, err := s.client.Query(sqlQuery, mysql.BinaryUUID(postID))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if rows.Err() != nil {
		return nil, errors.WithStack(rows.Err())
	}
	if !rows.Next() {
		return nil, nil
	}
	var post app.PostDTO
	err = rows.Scan(&post.ID, &post.Author, &post.Title, &post.Text)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()
	return &post, nil
}

func NewNewsLineQueryService(client mysql.Client, tarantoolClient tarantool.Client) app.NewsLineQueryService {
	return &newsLineQueryService{
		client:          client,
		tarantoolClient: tarantoolClient,
	}
}
