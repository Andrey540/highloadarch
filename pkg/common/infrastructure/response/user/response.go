package user

type User struct {
	UserID string `json:"user_id"`
}

type Data struct {
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

type Friend struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

type ListItemDTO struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	IsFriend bool   `json:"is_friend"`
}
