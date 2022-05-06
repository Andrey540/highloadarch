package infrastructure

import (
	"github.com/callicoder/go-docker/pkg/common/infrastructure/mysql"
	"github.com/callicoder/go-docker/pkg/common/uuid"
	"github.com/callicoder/go-docker/pkg/post/app"
	"github.com/pkg/errors"
)

type userProvider struct {
	client mysql.Client
}

func (s userProvider) ListUserSubscribers(userID uuid.UUID) ([]uuid.UUID, error) {
	const sqlQuery = `SELECT user_id FROM user_friend WHERE friend_id=?`
	rows, err := s.client.Query(sqlQuery, mysql.BinaryUUID(userID))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if rows.Err() != nil {
		return nil, errors.WithStack(rows.Err())
	}
	var result []uuid.UUID
	for rows.Next() {
		var subscriberID uuid.UUID
		err1 := rows.Scan(&subscriberID)
		if err1 != nil {
			return nil, errors.WithStack(err)
		}
		result = append(result, subscriberID)
	}
	defer rows.Close()
	return result, nil
}

func NewUserProvider(client mysql.Client) app.UserProvider {
	return &userProvider{
		client: client,
	}
}
