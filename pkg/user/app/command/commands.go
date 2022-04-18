package command

const (
	RegisterUserCommand     = "user.register"
	UpdateUserCommand       = "user.update"
	RemoveUserCommand       = "user.remove"
	AddUserFriendCommand    = "user.add_friend"
	RemoveUserFriendCommand = "user.remove_friend"
)

type RegisterUser struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Age       int    `json:"age"`
	Sex       int    `json:"sex"`
	Interests string `json:"interests"`
	City      string `json:"city"`
	Password  string `json:"password"`
}

func (command RegisterUser) CommandType() string {
	return RegisterUserCommand
}

func (command RegisterUser) CommandID() string {
	return command.ID
}

type UpdateUser struct {
	ID        string `json:"id"`
	UserID    string `json:"userId"`
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Age       int    `json:"age"`
	Sex       int    `json:"sex"`
	Interests string `json:"interests"`
	City      string `json:"city"`
	Password  string `json:"password"`
}

func (command UpdateUser) CommandType() string {
	return UpdateUserCommand
}

func (command UpdateUser) CommandID() string {
	return command.ID
}

type RemoveUser struct {
	ID     string `json:"id"`
	UserID string `json:"userId"`
}

func (command RemoveUser) CommandType() string {
	return RemoveUserCommand
}

func (command RemoveUser) CommandID() string {
	return command.ID
}

type AddUserFriend struct {
	ID       string `json:"id"`
	UserID   string `json:"userId"`
	FriendID string `json:"friendId"`
}

func (command AddUserFriend) CommandType() string {
	return AddUserFriendCommand
}

func (command AddUserFriend) CommandID() string {
	return command.ID
}

type RemoveUserFriend struct {
	ID       string `json:"id"`
	UserID   string `json:"userId"`
	FriendID string `json:"friendId"`
}

func (command RemoveUserFriend) CommandType() string {
	return RemoveUserFriendCommand
}

func (command RemoveUserFriend) CommandID() string {
	return command.ID
}
