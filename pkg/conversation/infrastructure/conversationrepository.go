package infrastructure

import (
	"database/sql"
	"encoding/json"
	"strings"

	commonsql "github.com/callicoder/go-docker/pkg/common/infrastructure/sql"
	"github.com/callicoder/go-docker/pkg/common/uuid"
	"github.com/callicoder/go-docker/pkg/conversation/app"
	"github.com/pkg/errors"
	satoriuuid "github.com/satori/go.uuid"
)

type conversationRepository struct {
	client commonsql.Client
}

func (r conversationRepository) NewID() uuid.UUID {
	return uuid.UUID(satoriuuid.NewV1())
}

func (r conversationRepository) GetUserConversation(userID, target uuid.UUID) (*app.Conversation, error) {
	sqlQuery := `SELECT conversation_id FROM user_conversation WHERE user_id=? AND target=?`
	rows, err := r.client.Query(sqlQuery, commonsql.BinaryUUID(userID), commonsql.BinaryUUID(target))
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
	var conversationID uuid.UUID
	err = rows.Scan(&conversationID)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return r.Find(conversationID)
}

func (r conversationRepository) Find(conversationID uuid.UUID) (*app.Conversation, error) {
	sqlQuery := `SELECT id, data FROM conversation WHERE id=?`
	rows, err := r.client.Query(sqlQuery, commonsql.BinaryUUID(conversationID))
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
	result, err := scanConversation(rows)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return result, nil
}

func (r conversationRepository) Store(conversation *app.Conversation) error {
	if len(conversation.UserIDs) != 2 {
		return app.ErrInvalidUsers
	}
	sqlQuery := `INSERT INTO conversation (id, data) VALUES (?, ?)`
	users := make([]string, 0, len(conversation.UserIDs))
	for _, userID := range conversation.UserIDs {
		users = append(users, userID.String())
	}
	usersStr, err := json.Marshal(users)
	if err != nil {
		return errors.WithStack(err)
	}
	_, err = r.client.Exec(sqlQuery, commonsql.BinaryUUID(conversation.ID), usersStr)
	if err != nil {
		return errors.WithStack(err)
	}

	values := []string{`(?, ?, ?)`, `(?, ?, ?)`}
	firstUserID := conversation.UserIDs[0]
	secondUserID := conversation.UserIDs[1]
	params := []interface{}{commonsql.BinaryUUID(conversation.ID), commonsql.BinaryUUID(firstUserID), commonsql.BinaryUUID(secondUserID),
		commonsql.BinaryUUID(conversation.ID), commonsql.BinaryUUID(secondUserID), commonsql.BinaryUUID(firstUserID)}
	sqlQuery = `INSERT INTO user_conversation (conversation_id, user_id, target) VALUES` + strings.Join(values, ",")

	_, err = r.client.Exec(sqlQuery, params...)
	return errors.WithStack(err)
}

func scanConversation(rows *sql.Rows) (*app.Conversation, error) {
	var conversation app.Conversation
	var usersStr string
	err := rows.Scan(&conversation.ID, &usersStr)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var users []string
	err = json.Unmarshal([]byte(usersStr), &users)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	for _, id := range users {
		uid, err1 := uuid.FromString(id)
		if err1 != nil {
			return nil, err1
		}
		conversation.UserIDs = append(conversation.UserIDs, uid)
	}
	return &conversation, errors.WithStack(err)
}

func NewConversationRepository(client commonsql.Client) app.ConversationRepository {
	return &conversationRepository{
		client: client,
	}
}
