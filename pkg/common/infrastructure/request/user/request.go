package user

const (
	AppID = "user"

	urlPrefix = "/" + AppID

	SignInURL          = urlPrefix + "/api/v1/signin"
	RegisterURL        = urlPrefix + "/api/v1/register"
	ProfileURL         = urlPrefix + "/api/v1/profile/{id}"
	FindUserURL        = urlPrefix + "/api/v1/profile/find/{username}"
	UpdateUserURL      = urlPrefix + "/api/v1/update/{id}"
	DeleteURL          = urlPrefix + "/api/v1/delete/{id}"
	AddFriendURL       = urlPrefix + "/api/v1/friend/add/{id}"
	RemoveFriendURL    = urlPrefix + "/api/v1/friend/remove/{id}"
	ListUserFriendsURL = urlPrefix + "/api/v1/friend/list/{id}"
	ListUsersURL       = urlPrefix + "/api/v1/list"
)

type RegisterUser struct {
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Age       int    `json:"age"`
	Sex       int    `json:"sex"`
	Interests string `json:"interests"`
	City      string `json:"city"`
	Password  string `json:"password"`
}

type UpdateUser struct {
	Username  string `json:"username"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Age       int    `json:"age"`
	Sex       int    `json:"sex"`
	Interests string `json:"interests"`
	City      string `json:"city"`
	Password  string `json:"password"`
}

type Auth struct {
	UserName string `json:"username"`
	Password string `json:"password"`
}

type ListUsers struct {
	UserIds []string `json:"user_ids"`
}
