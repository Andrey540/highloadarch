package infrastructure

import (
	"database/sql"

	"github.com/callicoder/go-docker/pkg/common/infrastructure/mysql"
	"github.com/callicoder/go-docker/pkg/common/uuid"
	"github.com/callicoder/go-docker/pkg/socialnetwork/app"
	"github.com/pkg/errors"
)

type userQueryService struct {
	userRepository  app.UserRepository
	passwordEncoder app.PasswordEncoder
	client          mysql.Client
}

func (s userQueryService) GetUserByNameAndPassword(userName, password string) (*app.UserProfileDTO, error) {
	const sqlQuery = selectUserSQL + ` WHERE u.username=? AND u.password=?`
	rows, err := s.client.Query(sqlQuery, userName, s.passwordEncoder.Encode(password))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !rows.Next() {
		return nil, nil
	}
	result, err := scanUserDTO(rows)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return result, rows.Close()
}

func (s userQueryService) GetUserProfile(id uuid.UUID) (*app.UserProfileDTO, error) {
	const sqlQuery = selectUserSQL + ` WHERE u.id=?`
	rows, err := s.client.Query(sqlQuery, mysql.BinaryUUID(id))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	if !rows.Next() {
		return nil, nil
	}
	result, err := scanUserDTO(rows)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return result, rows.Close()
}

func (s userQueryService) ListUserProfiles(userName string) ([]*app.UserProfileDTO, error) {
	const sqlQuery = selectUserSQL + ` WHERE u.firstName LIKE ? AND lastName LIKE ? ORDER BY id`
	userNameParameter := userName + "%"
	rows, err := s.client.Query(sqlQuery, userNameParameter, userNameParameter)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var users []*app.UserProfileDTO
	for rows.Next() {
		user, err1 := scanUserDTO(rows)
		if err1 != nil {
			return nil, errors.WithStack(err)
		}
		if err1 != nil {
			return nil, err1
		}
		users = append(users, user)
	}
	return users, nil
}

func (s userQueryService) ListUsers() ([]*app.UserListItemDTO, error) {
	const sqlQuery = `SELECT u.id, u.username FROM user u`
	rows, err := s.client.Query(sqlQuery)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var users []*app.UserListItemDTO
	for rows.Next() {
		var user app.UserListItemDTO
		err1 := rows.Scan(&user.ID, &user.Username)
		if err1 != nil {
			return nil, errors.WithStack(err)
		}
		if err1 != nil {
			return nil, err1
		}
		users = append(users, &user)
	}
	return users, nil
}

func (s userQueryService) ListUserFriends(userID uuid.UUID) ([]*app.UserFriendDTO, error) {
	const sqlQuery = `SELECT uf.friend_id, u.username FROM user_friend uf INNER JOIN user u ON u.id = uf.friend_id WHERE uf.user_id = ?`
	rows, err := s.client.Query(sqlQuery, mysql.BinaryUUID(userID))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	var result []*app.UserFriendDTO
	for rows.Next() {
		user, err1 := scanUserFriend(rows)
		if err1 != nil {
			return nil, err1
		}
		result = append(result, user)
	}
	return result, nil
}

func scanUserFriend(rows *sql.Rows) (*app.UserFriendDTO, error) {
	var user app.UserFriendDTO
	err := rows.Scan(&user.ID, &user.Username)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &user, nil
}

func scanUserDTO(rows *sql.Rows) (*app.UserProfileDTO, error) {
	var user app.UserProfileDTO
	err := rows.Scan(&user.ID, &user.Username, &user.FirstName, &user.LastName, &user.Age, &user.Sex, &user.Interests, &user.City, &user.Password)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &user, nil
}

func NewUserQueryService(client mysql.Client) app.UserQueryService {
	return &userQueryService{
		client:          client,
		passwordEncoder: app.NewPasswordEncoder(),
		userRepository:  NewUserRepository(client),
	}
}
