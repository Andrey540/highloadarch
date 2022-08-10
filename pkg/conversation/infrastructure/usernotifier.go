package infrastructure

import (
	"encoding/json"

	"github.com/callicoder/go-docker/pkg/common/infrastructure/realtime"
	"github.com/callicoder/go-docker/pkg/common/uuid"
	"github.com/pkg/errors"
)

type Message struct {
	Subscriber     string `json:"subscriber"`
	ConversationID string `json:"conversation_id"`
	MessageID      string `json:"message_id"`
	Message        string `json:"message"`
}

type UserNotifier struct {
	realtimeService realtime.Service
}

func (n UserNotifier) Notify(userIDs []uuid.UUID, conversationID, messageID, author uuid.UUID, message string) error {
	realtimeMessages := []realtime.Message{}
	for _, userID := range userIDs {
		if userID == author {
			continue
		}
		message := Message{Subscriber: userID.String(), ConversationID: conversationID.String(), MessageID: messageID.String(), Message: message}
		data, err := json.Marshal(message)
		if err != nil {
			return errors.WithStack(err)
		}
		realtimeMessages = append(realtimeMessages, realtime.Message{ChannelID: userID.String(), Data: data})
	}
	return n.realtimeService.Publish(realtimeMessages)
}

func (n UserNotifier) Close() error {
	if n.realtimeService == nil {
		return nil
	}
	return n.realtimeService.Close()
}

func NewUserNotifier(hosts []string, channel string) (UserNotifier, error) {
	realtimeService, err := realtime.NewService(hosts, channel)
	if err != nil {
		return UserNotifier{}, err
	}
	return UserNotifier{realtimeService: realtimeService}, nil
}
