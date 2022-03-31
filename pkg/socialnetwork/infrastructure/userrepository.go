package infrastructure

import (
	"database/sql"

	"github.com/callicoder/go-docker/pkg/common/infrastructure/mysql"
	"github.com/callicoder/go-docker/pkg/common/uuid"
	"github.com/callicoder/go-docker/pkg/socialnetwork/app"
	"github.com/pkg/errors"
	satoriuuid "github.com/satori/go.uuid"
)

const selectUserSQL = `
		SELECT u.* FROM user u`

type userRepository struct {
	client          mysql.Client
	passwordEncoder app.PasswordEncoder
	identityMap     map[uuid.UUID]*app.User
}

func (r userRepository) NewID() uuid.UUID {
	return uuid.UUID(satoriuuid.NewV1())
}

func (r userRepository) GetUserByNameAndPassword(userName, password string) (*app.User, error) {
	const sqlQuery = selectUserSQL + ` WHERE u.username=? AND u.password=?`
	rows, err := r.client.Query(sqlQuery, userName, r.passwordEncoder.Encode(password))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !rows.Next() {
		return nil, nil
	}
	result, err := r.hydrateUser(rows)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()
	return result, nil
}

func (r userRepository) GetUserByName(userName string) (*app.User, error) {
	const sqlQuery = selectUserSQL + ` WHERE u.username=?`
	rows, err := r.client.Query(sqlQuery, userName)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !rows.Next() {
		return nil, nil
	}
	result, err := r.hydrateUser(rows)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()
	return result, nil
}

func (r userRepository) Store(user app.User) error {
	if _, found := r.identityMap[user.ID]; found {
		sqlQuery := `UPDATE user
			SET username = ?, first_name = ?, last_name = ?, age = ?, sex = ?, interests = ?, city = ?, password = ?
			WHERE id = ?;`
		_, err := r.client.Exec(sqlQuery, user.Username, user.FirstName, user.LastName, user.Age, user.Sex, user.Interests, user.City, user.Password, mysql.BinaryUUID(user.ID))
		return errors.WithStack(err)
	}
	sqlQuery := `INSERT INTO user
			(id, username, first_name, last_name, age, sex, interests, city, password)
			VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?);`
	_, err := r.client.Exec(sqlQuery, mysql.BinaryUUID(user.ID), user.Username, user.FirstName, user.LastName, user.Age, user.Sex, user.Interests, user.City, user.Password)
	if err == nil {
		r.identityMap[user.ID] = &user
	}
	return errors.WithStack(err)
}

func (r userRepository) Delete(id uuid.UUID) error {
	const sqlQuery = `DELETE FROM user WHERE id=?`
	_, err := r.client.Exec(sqlQuery, mysql.BinaryUUID(id))
	return errors.WithStack(err)
}

func (r userRepository) GetUser(id uuid.UUID) (*app.User, error) {
	const sqlQuery = selectUserSQL + ` WHERE u.id=?`
	rows, err := r.client.Query(sqlQuery, mysql.BinaryUUID(id))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !rows.Next() {
		return nil, nil
	}
	result, err := r.hydrateUser(rows)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer rows.Close()
	return result, nil
}

func (r userRepository) hydrateUser(rows *sql.Rows) (*app.User, error) {
	user, err := scanUser(rows)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	r.identityMap[user.ID] = user
	return user, nil
}

func scanUser(rows *sql.Rows) (*app.User, error) {
	var user app.User
	err := rows.Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.Age, &user.Sex, &user.Interests, &user.City, &user.Password)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &user, nil
}

func NewUserRepository(client mysql.Client) app.UserRepository {
	return &userRepository{
		client:          client,
		passwordEncoder: app.NewPasswordEncoder(),
		identityMap:     make(map[uuid.UUID]*app.User),
	}
}
