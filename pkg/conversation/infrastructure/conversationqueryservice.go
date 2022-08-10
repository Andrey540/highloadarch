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
	dbName string
}

func (s conversationQueryService) GetConversation(conversationID, userID uuid.UUID) (*app.UserConversation, error) {
	sqlQuery := `SELECT conversation_id, target FROM ` + s.dbName + `.user_conversation WHERE conversation_id=? AND user_id=?`
	rows, err := s.client.Query(sqlQuery, mysql.BinaryUUID(conversationID), mysql.BinaryUUID(userID))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if rows.Err() != nil {
		return nil, errors.WithStack(rows.Err())
	}

	defer rows.Close()
	if !rows.Next() {
		return nil, nil
	}
	var conversation app.UserConversation
	err1 := rows.Scan(&conversation.ID, &conversation.UserID)
	return &conversation, errors.WithStack(err1)
}

func (s conversationQueryService) ListMessages(userID, conversationID uuid.UUID) ([]*app.UserMessage, error) {
	sqlQuery := `SELECT
				   m.id, m.user_id, m.conversation_id, m.text, MAX(urm.id IS NOT NULL)
				 FROM ` + s.dbName + `.message m
				 LEFT JOIN ` + s.dbName + `.user_unread_message urm ON urm.conversation_id = m.conversation_id AND urm.user_id = ?
				 WHERE m.conversation_id = ?
				 GROUP BY m.id ORDER BY m.created_at`
	rows, err := s.client.Query(sqlQuery, mysql.BinaryUUID(userID), mysql.BinaryUUID(conversationID))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if rows.Err() != nil {
		return nil, errors.WithStack(rows.Err())
	}
	var result []*app.UserMessage
	for rows.Next() {
		var message app.UserMessage
		err1 := rows.Scan(&message.ID, &message.UserID, &message.ConversationID, &message.Text, &message.Unread)
		if err1 != nil {
			return []*app.UserMessage{}, errors.WithStack(err)
		}
		result = append(result, &message)
	}
	defer rows.Close()
	return result, nil
}

func (s conversationQueryService) ListConversations(userID uuid.UUID) ([]*app.UserConversation, error) {
	sqlQuery := `SELECT conversation_id, target FROM ` + s.dbName + `.user_conversation WHERE user_id=?`
	rows, err := s.client.Query(sqlQuery, mysql.BinaryUUID(userID))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if rows.Err() != nil {
		return nil, errors.WithStack(rows.Err())
	}
	var result []*app.UserConversation
	for rows.Next() {
		var conversation app.UserConversation
		err1 := rows.Scan(&conversation.ID, &conversation.UserID)
		if err1 != nil {
			return []*app.UserConversation{}, errors.WithStack(err)
		}
		result = append(result, &conversation)
	}
	defer rows.Close()
	return result, nil
}

func NewConversationQueryService(client sql.Client, dbName string) app.ConversationQueryService {
	return &conversationQueryService{
		client: client,
		dbName: dbName,
	}
}
