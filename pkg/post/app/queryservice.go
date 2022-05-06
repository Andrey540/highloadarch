package app

import (
	"github.com/callicoder/go-docker/pkg/common/uuid"
)

type NewsLineItem struct {
	ID     uuid.UUID
	Author uuid.UUID
	Title  string
}

type PostDTO struct {
	ID     uuid.UUID
	Author uuid.UUID
	Title  string
	Text   string
}

type NewsLineQueryService interface {
	ListPosts(userID uuid.UUID) ([]*PostDTO, error)
	ListNews(userID uuid.UUID) (*[]NewsLineItem, error)
	GetPost(postID uuid.UUID) (*PostDTO, error)
}
