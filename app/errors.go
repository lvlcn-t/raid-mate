package app

import (
	"errors"
	"fmt"
)

// errShutdown is an error that occurs when the application is shutting down.
type errShutdown struct {
	ctxErr error
	apiErr error
	botErr error
	svcErr error
}

// Error returns the error message.
func (e errShutdown) Error() string {
	return fmt.Sprintf("%v", errors.Join(e.ctxErr, e.botErr, e.apiErr))
}

// Is checks if the target error is an [errShutdown].
func (e errShutdown) Is(target error) bool {
	_, ok := target.(*errShutdown)
	return ok
}

// HasErrors checks if there are any errors.
func (e errShutdown) HasErrors() bool {
	return e.ctxErr != nil || e.botErr != nil || e.apiErr != nil
}
