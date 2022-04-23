package app

import (
	"github.com/callicoder/go-docker/pkg/common/app/event"
	"github.com/callicoder/go-docker/pkg/common/uuid"
	"github.com/pkg/errors"
)

var ErrInvalidUsers = errors.New("Invalid number of users")

type Conversation struct {
	ID      uuid.UUID
	UserIDs []uuid.UUID
}

type Message struct {
	ID             uuid.UUID
	UserID         uuid.UUID
	ConversationID uuid.UUID
	Text           string
}

type ConversationRepository interface {
	NewID() uuid.UUID
	GetUserConversation(userID, target uuid.UUID) (*Conversation, error)
	Store(conversation *Conversation) error
}

type MessageRepository interface {
	NewID() uuid.UUID
	Store(message *Message) error
}

type ConversationService interface {
	StartUserConversation(userID, target uuid.UUID) (uuid.UUID, error)
	AddMessage(conversationID, userID uuid.UUID, text string) (uuid.UUID, error)
}

type service struct {
	conversationRepository ConversationRepository
	messageRepository      MessageRepository
	eventDispatcher        event.Dispatcher
}

func (s service) StartUserConversation(userID, target uuid.UUID) (uuid.UUID, error) {
	existingConversation, err := s.conversationRepository.GetUserConversation(userID, target)
	if err != nil {
		return uuid.Nil, errors.WithStack(err)
	}
	if existingConversation != nil {
		return existingConversation.ID, nil
	}
	userIDs := []uuid.UUID{userID, target}
	conversation := Conversation{
		ID:      s.conversationRepository.NewID(),
		UserIDs: userIDs,
	}
	err = s.conversationRepository.Store(&conversation)
	if err != nil {
		return uuid.Nil, errors.WithStack(err)
	}
	err = s.eventDispatcher.Dispatch(event.ConversationCreated{ConversationID: conversation.ID.String(), UserIDs: uuid.ToStrings(userIDs)})
	return conversation.ID, errors.WithStack(err)
}

func (s service) AddMessage(conversationID, userID uuid.UUID, text string) (uuid.UUID, error) {
	messageID := s.messageRepository.NewID()
	message := Message{
		ID:             messageID,
		ConversationID: conversationID,
		UserID:         userID,
		Text:           text,
	}
	err := s.messageRepository.Store(&message)
	if err != nil {
		return uuid.Nil, errors.WithStack(err)
	}
	err = s.eventDispatcher.Dispatch(event.MessageAdded{MessageID: messageID.String(), ConversationID: conversationID.String(), UserID: userID.String(), Text: text})
	return messageID, errors.WithStack(err)
}

func NewConversationService(conversationRepository ConversationRepository, messageRepository MessageRepository, eventDispatcher event.Dispatcher) ConversationService {
	return &service{
		conversationRepository: conversationRepository,
		messageRepository:      messageRepository,
		eventDispatcher:        eventDispatcher,
	}
}
