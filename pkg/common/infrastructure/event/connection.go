package event

type Connection interface {
	Start() error
	Stop() error
}
