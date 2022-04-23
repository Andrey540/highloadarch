package app

import (
	"github.com/callicoder/go-docker/pkg/common/uuid"
)

type ConversationQueryService interface {
	ListMessages(conversationID uuid.UUID) ([]*Message, error)
}
