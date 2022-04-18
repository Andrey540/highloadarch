package infrastructure

import (
	commonapp "github.com/callicoder/go-docker/pkg/common/app"
	"github.com/callicoder/go-docker/pkg/user/app"
)

type UnitOfWorkCompleteNotifier func()

func NewNotifyingUnitOfWorkFactory(unitOfWorkFactory commonapp.UnitOfWorkFactory, completeNotifier UnitOfWorkCompleteNotifier) commonapp.UnitOfWorkFactory {
	return &notifyingUnitOfWorkFactoryDecorator{factory: unitOfWorkFactory, completeNotifier: completeNotifier}
}

type notifyingUnitOfWorkFactoryDecorator struct {
	factory          commonapp.UnitOfWorkFactory
	completeNotifier UnitOfWorkCompleteNotifier
}

func (decorator notifyingUnitOfWorkFactoryDecorator) NewUnitOfWork(lockNames []string) (commonapp.UnitOfWork, error) {
	unitOfWork, err := decorator.factory.NewUnitOfWork(lockNames)
	if err != nil {
		return nil, err
	}

	if decorator.completeNotifier != nil {
		return &notifyingUnitOfWorkDecorator{UnitOfWork: unitOfWork.(app.UnitOfWork), completeNotifier: decorator.completeNotifier}, nil
	}
	return unitOfWork.(app.UnitOfWork), nil
}

type notifyingUnitOfWorkDecorator struct {
	app.UnitOfWork
	completeNotifier UnitOfWorkCompleteNotifier
}

func (decorator notifyingUnitOfWorkDecorator) Complete(err error) error {
	err = decorator.UnitOfWork.Complete(err)
	if err == nil {
		decorator.completeNotifier()
	}
	return err
}
