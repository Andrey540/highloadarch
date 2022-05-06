package app

import (
	"encoding/json"

	"github.com/callicoder/go-docker/pkg/common/app/event"

	"github.com/pkg/errors"
)

type eventBody struct {
	Type    string `json:"type"`
	Payload []byte `json:"payload"`
}

type Serializer interface {
	Serialize(event event.Event) (msg string, err error)
	Deserialize(eventType, msg string) (event.Event, error)
}

type serializer struct {
}

func (s *serializer) Serialize(e event.Event) (string, error) {
	payload, err := json.Marshal(e)
	if err != nil {
		return "", err
	}

	body := eventBody{Type: e.EventType(), Payload: payload}
	messageBody, err := json.Marshal(body)
	return string(messageBody), err
}

func (s *serializer) Deserialize(eventType, msg string) (result event.Event, err error) {
	body := eventBody{}
	err = json.Unmarshal([]byte(msg), &body)
	if err != nil {
		return nil, err
	}
	result = nil
	switch t := eventType; t {
	case event.UserCreatedEvent:
		currentEvent := event.UserCreated{}
		err = json.Unmarshal(body.Payload, &currentEvent)
		result = currentEvent
	case event.UserUpdatedEvent:
		currentEvent := event.UserUpdated{}
		err = json.Unmarshal(body.Payload, &currentEvent)
		result = currentEvent
	case event.UserRemovedEvent:
		currentEvent := event.UserRemoved{}
		err = json.Unmarshal(body.Payload, &currentEvent)
		result = currentEvent
	case event.UserFriendAddedEvent:
		currentEvent := event.UserFriendAdded{}
		err = json.Unmarshal(body.Payload, &currentEvent)
		result = currentEvent
	case event.UserFriendRemovedEvent:
		currentEvent := event.UserFriendRemoved{}
		err = json.Unmarshal(body.Payload, &currentEvent)
		result = currentEvent
	case event.ConversationCreatedEvent:
		currentEvent := event.ConversationCreated{}
		err = json.Unmarshal(body.Payload, &currentEvent)
		result = currentEvent
	case event.MessageAddedEvent:
		currentEvent := event.MessageAdded{}
		err = json.Unmarshal(body.Payload, &currentEvent)
		result = currentEvent
	case event.PostCreatedEvent:
		currentEvent := event.PostCreated{}
		err = json.Unmarshal(body.Payload, &currentEvent)
		result = currentEvent
	default:
		return nil, errors.New("Undefined event type " + eventType)
	}
	return result, err
}

func NewSerializer() Serializer {
	return &serializer{}
}
