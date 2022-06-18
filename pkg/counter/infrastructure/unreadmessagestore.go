package infrastructure

import (
	"github.com/callicoder/go-docker/pkg/common/infrastructure/sql"
	"github.com/callicoder/go-docker/pkg/common/uuid"
	"github.com/callicoder/go-docker/pkg/counter/app"
	"github.com/pkg/errors"
)

type unreadMessageStore struct {
	client sql.Client
	dbName string
}

func (r unreadMessageStore) IncreaseUnreadMessages(conversationID, userID uuid.UUID) error {
	sqlQuery := `UPDATE ` + r.dbName + `user_unread_message SET count = count + 1 WHERE conversation_id = ? AND user_id = ?`
	_, err := r.client.Exec(sqlQuery, sql.BinaryUUID(conversationID), sql.BinaryUUID(userID))
	if err != nil {
		return errors.WithStack(err)
	}

	sqlQuery = `INSERT IGNORE INTO ` + r.dbName + `user_unread_message (conversation_id, user_id, count) VALUES (?, ?, 1)`
	_, err = r.client.Exec(sqlQuery, sql.BinaryUUID(conversationID), sql.BinaryUUID(userID))
	return errors.WithStack(err)
}

func (r unreadMessageStore) DecreaseUnreadMessages(conversationID, userID uuid.UUID, count int) error {
	if count == 0 {
		return nil
	}
	sqlQuery := `UPDATE ` + r.dbName + `user_unread_message SET count = count - ? WHERE conversation_id = ? AND user_id = ?`
	_, err := r.client.Exec(sqlQuery, count, sql.BinaryUUID(conversationID), sql.BinaryUUID(userID))
	return errors.WithStack(err)
}

func NewUnreadMessageStore(client sql.Client, dbName string) app.Store {
	if dbName != "" {
		dbName += "."
	}
	return &unreadMessageStore{
		client: client,
		dbName: dbName,
	}
}
