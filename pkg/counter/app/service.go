package app

import (
	"github.com/callicoder/go-docker/pkg/common/uuid"
)

type UserConversation struct {
	ConversationID uuid.UUID
	Count          int
}

type UnreadMessagesQueryService interface {
	ListConversations(userID uuid.UUID) ([]*UserConversation, error)
}

type Store interface {
	IncreaseUnreadMessages(conversationID, userID uuid.UUID) error
	DecreaseUnreadMessages(conversationID, userID uuid.UUID, count int) error
}
