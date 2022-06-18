package app

import (
	"github.com/callicoder/go-docker/pkg/common/app"
)

type RepositoryFactory interface {
	UnreadMessagesStore() Store
}

type UnitOfWork interface {
	RepositoryFactory
	app.UnitOfWork
}
