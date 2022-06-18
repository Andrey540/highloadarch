package infrastructure

import (
	"github.com/callicoder/go-docker/pkg/common/infrastructure/mysql"
	"github.com/callicoder/go-docker/pkg/common/infrastructure/sql"
	"github.com/callicoder/go-docker/pkg/common/uuid"
	"github.com/callicoder/go-docker/pkg/counter/app"
	"github.com/pkg/errors"
)

type conversationQueryService struct {
	client sql.Client
	dbName string
}

func (s conversationQueryService) ListConversations(userID uuid.UUID) ([]*app.UserConversation, error) {
	sqlQuery := `SELECT conversation_id, count FROM ` + s.dbName + `user_unread_message WHERE user_id=?`
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
		err1 := rows.Scan(&conversation.ConversationID, &conversation.Count)
		if err1 != nil {
			return []*app.UserConversation{}, errors.WithStack(err)
		}
		result = append(result, &conversation)
	}
	defer rows.Close()
	return result, nil
}

func NewUnreadMessagesQueryService(client sql.Client, dbName string) app.UnreadMessagesQueryService {
	if dbName != "" {
		dbName += "."
	}
	return &conversationQueryService{
		client: client,
		dbName: dbName,
	}
}
