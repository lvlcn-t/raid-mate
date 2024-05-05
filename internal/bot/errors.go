package bot

import (
	"errors"
	"fmt"
)

var _ error = (*ErrShutdown)(nil)

// ErrShutdown is an error that occurs when the bot is shutting down.
type ErrShutdown struct {
	ctxErr error
	mgrErr error
	svcErr error
}

// Error returns the error message.
func (e *ErrShutdown) Error() string {
	return fmt.Sprintf("%v", errors.Join(e.ctxErr, e.mgrErr, e.svcErr))
}

// Is checks if the target error is an [ErrShutdown].
func (e *ErrShutdown) Is(target error) bool {
	_, ok := target.(*ErrShutdown)
	return ok
}

// HasErrors checks if there are any errors.
func (e *ErrShutdown) HasErrors() bool {
	return e.ctxErr != nil || e.mgrErr != nil || e.svcErr != nil
}
