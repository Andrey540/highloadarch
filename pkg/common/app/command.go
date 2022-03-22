package app

type Command interface {
	CommandID() string
	CommandType() string
}
