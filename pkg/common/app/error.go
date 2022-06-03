package app

import "github.com/pkg/errors"

var ErrCommandAlreadyProcessed = errors.New("Command already processed")
var ErrCommandHandlerNotFound = errors.New("Command handler not found")
var ErrNotAuthenticated = errors.New("Not authenticated")
var ErrPermissionDenied = errors.New("Permission denied")
var ErrNotFound = errors.New("Not found")
