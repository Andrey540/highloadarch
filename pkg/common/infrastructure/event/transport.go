package event

import "github.com/callicoder/go-docker/pkg/common/app"

type Handler interface {
	Handle(msg string) error
}

type Transport interface {
	app.Transport
	SetHandler(handler Handler)
}
