package app

import (
	"github.com/callicoder/go-docker/pkg/common/uuid"
)

type UserConversation struct {
	ID     uuid.UUID
	UserID uuid.UUID
}

type UserMessage struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	ConversationID uuid.UUID
	Text           string
	Unread         bool
}

type ConversationQueryService interface {
	ListConversations(userID uuid.UUID) ([]*UserConversation, error)
	ListMessages(userID, conversationID uuid.UUID) ([]*UserMessage, error)
}
