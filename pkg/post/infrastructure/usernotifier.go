package infrastructure

import (
	"encoding/json"

	"github.com/callicoder/go-docker/pkg/common/infrastructure/realtime"
	"github.com/callicoder/go-docker/pkg/common/uuid"
	"github.com/pkg/errors"
)

type Message struct {
	Subscriber string `json:"subscriber"`
	PostID     string `json:"post_id"`
	Author     string `json:"author"`
	Title      string `json:"title"`
}

type UserNotifier struct {
	realtimeService realtime.Service
}

func (n UserNotifier) Notify(userIDs []uuid.UUID, postID uuid.UUID, author, title string) error {
	realtimeMessages := []realtime.Message{}
	for _, userID := range userIDs {
		message := Message{Subscriber: userID.String(), PostID: postID.String(), Author: author, Title: title}
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
