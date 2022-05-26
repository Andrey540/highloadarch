package infrastructure

import (
	"database/sql"

	"github.com/callicoder/go-docker/pkg/common/infrastructure/mysql"
	"github.com/callicoder/go-docker/pkg/common/uuid"
	"github.com/callicoder/go-docker/pkg/post/app"
	"github.com/pkg/errors"
)

type userRepository struct {
	client mysql.Client
}

func (r userRepository) AddUser(user *app.User) error {
	const sqlQuery = `INSERT INTO user (id, username) VALUES(?, ?)`
	_, err := r.client.Exec(sqlQuery, mysql.BinaryUUID(user.UserID), user.Username)
	return errors.WithStack(err)
}

func (r userRepository) RemoveUser(userID uuid.UUID) error {
	const sqlQuery = `DELETE FROM user WHERE id = ?`
	_, err := r.client.Exec(sqlQuery, mysql.BinaryUUID(userID))
	return errors.WithStack(err)
}

func (r userRepository) GetUser(userID uuid.UUID) (*app.User, error) {
	const sqlQuery = `SELECT * FROM user WHERE id=?`
	rows, err := r.client.Query(sqlQuery, mysql.BinaryUUID(userID))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !rows.Next() {
		return nil, nil
	}
	result, err := scanUser(rows)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()
	return result, nil
}

func scanUser(rows *sql.Rows) (*app.User, error) {
	var user app.User
	err := rows.Scan(&user.UserID, &user.Username)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &user, nil
}

func NewUserRepository(client mysql.Client) app.UserRepository {
	return &userRepository{
		client: client,
	}
}
