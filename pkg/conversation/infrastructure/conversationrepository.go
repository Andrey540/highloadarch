package infrastructure

import (
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/callicoder/go-docker/pkg/common/infrastructure/mysql"
	"github.com/callicoder/go-docker/pkg/common/uuid"
	"github.com/callicoder/go-docker/pkg/conversation/app"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	satoriuuid "github.com/satori/go.uuid"
)

type conversationRepository struct {
	client mysql.Client
}

func (r conversationRepository) NewID() uuid.UUID {
	return uuid.UUID(satoriuuid.NewV1())
}

func (r conversationRepository) GetUsersConversation(userIDs []uuid.UUID) (*app.Conversation, error) {
	if len(userIDs) < 2 {
		return nil, nil
	}
	filterQuery, filterQueryParams, err := sqlx.In(" AND uc2.user_id IN (?)", mysql.ConvertToUuids(userIDs))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	sqlQuery := `SELECT uc2.conversation_id, JSON_ARRAYAGG(BIN_TO_UUID(uc2.user_id)) AS members, count(uc2.user_id) AS members_count FROM user_conversation uc1 
                      INNER JOIN user_conversation uc2 ON uc2.conversation_id = uc1.conversation_id
                      WHERE uc1.user_id=? ` + filterQuery + ` GROUP BY uc2.conversation_id HAVING members_count = ?`
	params := []interface{}{mysql.BinaryUUID(userIDs[0])}
	params = append(params, filterQueryParams...)
	params = append(params, len(userIDs))
	rows, err := r.client.Query(sqlQuery, params...)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !rows.Next() {
		return nil, nil
	}
	result, err := scanConversation(rows)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()
	return result, nil
}

func (r conversationRepository) Store(conversation *app.Conversation) error {
	sqlQuery := `INSERT INTO conversation (id) VALUES (?)`
	_, err := r.client.Exec(sqlQuery, mysql.BinaryUUID(conversation.ID))
	if err != nil {
		return errors.WithStack(err)
	}

	values := make([]string, 0, len(conversation.UserIDs))
	params := []interface{}{}
	for _, userID := range conversation.UserIDs {
		values = append(values, `(?, ?)`)
		params = append(params, mysql.BinaryUUID(conversation.ID), mysql.BinaryUUID(userID))
	}
	sqlQuery = `INSERT INTO user_conversation (conversation_id, user_id) VALUES` + strings.Join(values, ",")

	_, err = r.client.Exec(sqlQuery, params...)
	return errors.WithStack(err)
}

func scanConversation(rows *sql.Rows) (*app.Conversation, error) {
	var conversation app.Conversation
	var membersCount int
	var usersStr string
	err := rows.Scan(&conversation.ID, &usersStr, &membersCount)
	if err != nil {
		return nil, err
	}
	var users []string
	err = json.Unmarshal([]byte(usersStr), &users)
	if err != nil {
		return nil, err
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

func NewConversationRepository(client mysql.Client) app.ConversationRepository {
	return &conversationRepository{
		client: client,
	}
}
