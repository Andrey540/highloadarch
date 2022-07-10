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
	AuthorID       uuid.UUID
	ConversationID uuid.UUID
	Text           string
}

type ConversationUnreadMessage struct {
	UserIDs        []uuid.UUID
	ConversationID uuid.UUID
	MessageID      uuid.UUID
}

type UserUnreadMessages struct {
	UserID         uuid.UUID
	ConversationID uuid.UUID
	MessageIDs     []uuid.UUID
}

type ConversationRepository interface {
	NewID() uuid.UUID
	Find(conversationID uuid.UUID) (*Conversation, error)
	GetUserConversation(userID, target uuid.UUID) (*Conversation, error)
	Store(conversation *Conversation) error
}

type MessageRepository interface {
	NewID() uuid.UUID
	Store(message *Message) error
}

type UnreadMessagesRepository interface {
	Store(message *ConversationUnreadMessage) error
	FindUserUnreadMessages(conversationID, userID uuid.UUID, messageIDs []uuid.UUID) (*UserUnreadMessages, error)
	Remove(messages *UserUnreadMessages) error
}

type ConversationService interface {
	StartUserConversation(userID, target uuid.UUID) (uuid.UUID, error)
	AddMessage(conversationID, userID uuid.UUID, text string) (uuid.UUID, error)
	ReadMessages(conversationID, userID uuid.UUID, messageIDs []uuid.UUID) error
}

type service struct {
	conversationRepository   ConversationRepository
	messageRepository        MessageRepository
	unreadMessagesRepository UnreadMessagesRepository
	eventDispatcher          event.Dispatcher
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

func (s service) AddMessage(conversationID, authorID uuid.UUID, text string) (uuid.UUID, error) {
	conversation, err := s.conversationRepository.Find(conversationID)
	if err != nil {
		return uuid.Nil, errors.WithStack(err)
	}
	messageID := s.messageRepository.NewID()
	message := Message{
		ID:             messageID,
		ConversationID: conversationID,
		AuthorID:       authorID,
		Text:           text,
	}
	err = s.messageRepository.Store(&message)
	if err != nil {
		return uuid.Nil, errors.WithStack(err)
	}
	err = s.eventDispatcher.Dispatch(event.MessageAdded{
		MessageID:      messageID.String(),
		ConversationID: conversationID.String(),
		AuthorID:       authorID.String(),
		Text:           text,
		UserIDs:        uuid.ToStrings(conversation.UserIDs),
	})
	if err != nil {
		return uuid.Nil, errors.WithStack(err)
	}
	userIDs := []uuid.UUID{}
	for _, conversationUserID := range conversation.UserIDs {
		if conversationUserID != authorID {
			userIDs = append(userIDs, conversationUserID)
		}
	}
	unreadMessage := ConversationUnreadMessage{
		ConversationID: conversationID,
		UserIDs:        userIDs,
		MessageID:      messageID,
	}
	err = s.unreadMessagesRepository.Store(&unreadMessage)
	if err != nil {
		return uuid.Nil, errors.WithStack(err)
	}
	for _, userID := range userIDs {
		err = s.eventDispatcher.Dispatch(event.UnreadMessageAdded{
			ConversationID: conversationID.String(),
			UserID:         userID.String(),
		})
		if err != nil {
			return uuid.Nil, errors.WithStack(err)
		}
	}
	return messageID, nil
}

func (s service) ReadMessages(conversationID, userID uuid.UUID, messageIDs []uuid.UUID) error {
	if len(messageIDs) == 0 {
		return nil
	}
	messages, err := s.unreadMessagesRepository.FindUserUnreadMessages(conversationID, userID, messageIDs)
	if err != nil || len(messages.MessageIDs) == 0 {
		return errors.WithStack(err)
	}
	err = s.unreadMessagesRepository.Remove(messages)
	if err != nil {
		return errors.WithStack(err)
	}
	err = s.eventDispatcher.Dispatch(event.MessagesRead{UserID: messages.UserID.String(), ConversationID: messages.ConversationID.String(), MessageIDs: uuid.ToStrings(messages.MessageIDs)})
	return errors.WithStack(err)
}

func NewConversationService(conversationRepository ConversationRepository, messageRepository MessageRepository, unreadMessageRepository UnreadMessagesRepository, eventDispatcher event.Dispatcher) ConversationService {
	return &service{
		conversationRepository:   conversationRepository,
		messageRepository:        messageRepository,
		unreadMessagesRepository: unreadMessageRepository,
		eventDispatcher:          eventDispatcher,
	}
}
