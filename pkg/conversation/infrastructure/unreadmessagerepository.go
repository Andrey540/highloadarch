package infrastructure

import (
	"github.com/callicoder/go-docker/pkg/common/infrastructure/sql"
	"github.com/callicoder/go-docker/pkg/common/uuid"
	"github.com/callicoder/go-docker/pkg/conversation/app"
	"github.com/pkg/errors"
	satoriuuid "github.com/satori/go.uuid"

	"strings"
)

type unreadMessageRepository struct {
	client sql.Client
	dbName string
}

func (r unreadMessageRepository) NewID() uuid.UUID {
	return uuid.UUID(satoriuuid.NewV1())
}

func (r unreadMessageRepository) Store(message *app.ConversationUnreadMessage) error {
	if len(message.UserIDs) == 0 {
		return nil
	}
	sqlQuery := `INSERT INTO ` + r.dbName + `.user_unread_message (conversation_id, user_id, message_id) VALUES`
	values := []string{}
	params := []interface{}{}
	for _, userID := range message.UserIDs {
		params = append(params, sql.BinaryUUID(message.ConversationID), sql.BinaryUUID(userID), sql.BinaryUUID(message.MessageID))
		values = append(values, `(?, ?, ?)`)
	}
	_, err := r.client.Exec(sqlQuery+strings.Join(values, ","), params...)
	return errors.WithStack(err)
}

func (r unreadMessageRepository) FindUserUnreadMessages(conversationID, userID uuid.UUID, messageIDs []uuid.UUID) (*app.UserUnreadMessages, error) {
	if len(messageIDs) == 0 {
		return nil, nil
	}
	sqlQuery := `SELECT message_id FROM ` + r.dbName + `.user_unread_message WHERE conversation_id = ? AND user_id = ? AND message_id IN `
	values := []string{}
	params := []interface{}{sql.BinaryUUID(conversationID), sql.BinaryUUID(userID)}
	for _, messageID := range messageIDs {
		params = append(params, sql.BinaryUUID(messageID))
		values = append(values, `?`)
	}
	rows, err := r.client.Query(sqlQuery+`(`+strings.Join(values, ",")+`)`, params...)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if rows.Err() != nil {
		return nil, errors.WithStack(rows.Err())
	}
	result := app.UserUnreadMessages{ConversationID: conversationID, UserID: userID}
	messages := []uuid.UUID{}
	for rows.Next() {
		var messageID uuid.UUID
		err1 := rows.Scan(&messageID)
		if err1 != nil {
			return nil, errors.WithStack(err)
		}
		messages = append(messages, messageID)
	}
	result.MessageIDs = messages
	defer rows.Close()
	return &result, nil
}

func (r unreadMessageRepository) Remove(messages *app.UserUnreadMessages) error {
	if len(messages.MessageIDs) == 0 {
		return nil
	}
	sqlQuery := `DELETE FROM ` + r.dbName + `.user_unread_message WHERE conversation_id = ? AND user_id = ? AND message_id IN `
	values := []string{}
	params := []interface{}{sql.BinaryUUID(messages.ConversationID), sql.BinaryUUID(messages.UserID)}
	for _, messageID := range messages.MessageIDs {
		params = append(params, sql.BinaryUUID(messageID))
		values = append(values, `?`)
	}
	_, err := r.client.Exec(sqlQuery+`(`+strings.Join(values, ",")+`)`, params...)
	return errors.WithStack(err)
}

func NewUnreadMessageRepository(client sql.Client, dbName string) app.UnreadMessagesRepository {
	return &unreadMessageRepository{
		client: client,
		dbName: dbName,
	}
}
