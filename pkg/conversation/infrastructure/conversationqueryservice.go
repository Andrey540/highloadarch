package infrastructure

import (
	"github.com/callicoder/go-docker/pkg/common/infrastructure/mysql"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/sql"
	"github.com/callicoder/go-docker/pkg/common/uuid"
	"github.com/callicoder/go-docker/pkg/conversation/app"
	"github.com/pkg/errors"
)

type conversationQueryService struct {
	client sql.Client
}

func (s conversationQueryService) ListMessages(conversationID uuid.UUID) ([]*app.Message, error) {
	const sqlQuery = `SELECT id, user_id, conversation_id, text FROM message WHERE conversation_id=?`
	rows, err := s.client.Query(sqlQuery, mysql.BinaryUUID(conversationID))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if rows.Err() != nil {
		return nil, errors.WithStack(rows.Err())
	}
	var result []*app.Message
	for rows.Next() {
		var message app.Message
		err1 := rows.Scan(&message.ID, &message.UserID, &message.ConversationID, &message.Text)
		if err1 != nil {
			return []*app.Message{}, errors.WithStack(err)
		}
		result = append(result, &message)
	}
	defer rows.Close()
	return result, nil
}

func NewConversationQueryService(client sql.Client) app.ConversationQueryService {
	return &conversationQueryService{
		client: client,
	}
}
