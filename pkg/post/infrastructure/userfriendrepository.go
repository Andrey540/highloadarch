package infrastructure

import (
	"github.com/callicoder/go-docker/pkg/common/infrastructure/mysql"
	"github.com/callicoder/go-docker/pkg/post/app"
	"github.com/pkg/errors"
)

type userFriendRepository struct {
	client mysql.Client
}

func (r userFriendRepository) AddFriend(userFriend *app.UserFriend) error {
	const sqlQuery = `INSERT INTO user_friend (user_id, friend_id) VALUES(?, ?)`
	_, err := r.client.Exec(sqlQuery, mysql.BinaryUUID(userFriend.UserID), mysql.BinaryUUID(userFriend.FriendID))
	return errors.WithStack(err)
}

func (r userFriendRepository) RemoveFriend(userFriend *app.UserFriend) error {
	const sqlQuery = `DELETE FROM user_friend WHERE user_id = ? AND friend_id = ?`
	_, err := r.client.Exec(sqlQuery, mysql.BinaryUUID(userFriend.UserID), mysql.BinaryUUID(userFriend.FriendID))
	return errors.WithStack(err)
}

func NewUserFriendRepository(client mysql.Client) app.UserFriendRepository {
	return &userFriendRepository{
		client: client,
	}
}
