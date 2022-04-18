package app

import (
	"github.com/callicoder/go-docker/pkg/common/uuid"
)

type ConversationQueryService interface {
	GetUsersConversation(userIDs []uuid.UUID) (*Conversation, error)
	ListMessages(conversationID uuid.UUID) ([]*Message, error)
}
