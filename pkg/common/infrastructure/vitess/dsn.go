package vitess

import "fmt"

type DSN struct {
	Host string
	Port string
}

func (dsn *DSN) String() string {
	return fmt.Sprintf("%s:%s", dsn.Host, dsn.Port)
}
