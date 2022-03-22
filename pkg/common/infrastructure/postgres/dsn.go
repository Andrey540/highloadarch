package postgres

import "fmt"

type DSN struct {
	User     string
	Password string
	Host     string
	Port     string
	Database string
}

func (dsn *DSN) String() string {
	return fmt.Sprintf("host=%s port=%s user=%s "+"password=%s dbname=%s sslmode=disable",
		dsn.Host, dsn.Port, dsn.User, dsn.Password, dsn.Database)
}
