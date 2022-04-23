package infrastructure

import (
	"github.com/callicoder/go-docker/pkg/common/infrastructure/sql"
	"github.com/callicoder/go-docker/pkg/common/uuid"
	"github.com/callicoder/go-docker/pkg/conversation/app"
	"github.com/pkg/errors"
	satoriuuid "github.com/satori/go.uuid"
)

type messageRepository struct {
	client sql.Client
}

func (r messageRepository) NewID() uuid.UUID {
	return uuid.UUID(satoriuuid.NewV1())
}

func (r messageRepository) Store(message *app.Message) error {
	const sqlQuery = `INSERT INTO message (id, conversation_id, user_id, text) VALUES(?, ?, ?, ?)`
	_, err := r.client.Exec(sqlQuery, sql.BinaryUUID(message.ID), sql.BinaryUUID(message.ConversationID), sql.BinaryUUID(message.UserID), message.Text)
	return errors.WithStack(err)
}

func NewMessageRepository(client sql.Client) app.MessageRepository {
	return &messageRepository{
		client: client,
	}
}
