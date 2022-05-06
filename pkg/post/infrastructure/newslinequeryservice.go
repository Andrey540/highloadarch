package infrastructure

import (
	"github.com/callicoder/go-docker/pkg/common/infrastructure/mysql"
	"github.com/callicoder/go-docker/pkg/common/uuid"
	"github.com/callicoder/go-docker/pkg/post/app"
	"github.com/pkg/errors"
)

type newsLineQueryService struct {
	client        mysql.Client
	newsLineCache NewsLineCache
}

func (s newsLineQueryService) ListPosts(userID uuid.UUID) ([]*app.PostDTO, error) {
	const sqlQuery = `SELECT id, author_id, title, text FROM post WHERE author_id=?`
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
	news, err := s.newsLineCache.GetUserNews(userID)
	if err != nil {
		return nil, err
	}
	if news != nil {
		return news, nil
	}
	const sqlQuery = `SELECT p.id, p.author_id, p.title FROM post p INNER JOIN news_line nl ON nl.post_id = p.id WHERE nl.user_id=?`
	rows, err := s.client.Query(sqlQuery, mysql.BinaryUUID(userID))
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
	return &result, s.newsLineCache.SaveUserNews(userID, &result)
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

func NewNewsLineQueryService(client mysql.Client, newsLineCache NewsLineCache) app.NewsLineQueryService {
	return &newsLineQueryService{
		client:        client,
		newsLineCache: newsLineCache,
	}
}
