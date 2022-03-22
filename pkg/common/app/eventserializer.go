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
	return nil, errors.New("Undefined event type " + eventType)
}

func NewSerializer() Serializer {
	return &serializer{}
}
