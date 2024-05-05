package manager

import "fmt"

// Assign the error types to variables to ensure compile-time assertions.
var (
	_ error = (*errBase)(nil)
	_ error = (*ErrNilSession)(nil)
	_ error = (*ErrAlreadyConnected)(nil)
	_ error = (*ErrNotConnected)(nil)
)

// errBase is the base error type for all shard errors.
type errBase struct {
	// ID is the shard ID.
	ID int
}

// Error implements the error interface.
// It returns the error message for the shard error.
func (e errBase) Error() string {
	if e.ID == 0 {
		return "shards: unknown error"
	}
	return fmt.Sprintf("shard %d: unknown error", e.ID)
}

// Is implements the errors.Is interface.
// It returns true if the target error is of the same type as the receiver.
func (e errBase) Is(target error) bool {
	switch t := target.(type) {
	case errBase:
		return true
	case *ErrNilSession:
		return t.errBase.Is(e)
	case *ErrAlreadyConnected:
		return t.errBase.Is(e)
	case *ErrNotConnected:
		return t.errBase.Is(e)
	default:
		return false
	}
}

// ErrNilSession is the error for when the session is nil.
// This error is returned when the session is nil.
type ErrNilSession struct{ errBase }

// ErrAlreadyConnected is the error for when the shard is already connected.
// This error is returned when the shard is already connected.
type ErrAlreadyConnected struct{ errBase }

// ErrNotConnected is the error for when the shard is not connected.
// This error is returned when the shard is not connected.
type ErrNotConnected struct{ errBase }

// ErrShutdown is the error for when the shard manager fails to shut down all shards.
// This error is returned when the shard manager fails to shut down all shards.
type ErrShutdown struct{ Err error }

// newErrNilSession creates a new ErrNilSession error.
func newErrNilSession() *ErrNilSession {
	return &ErrNilSession{errBase: errBase{ID: 0}}
}

// newErrAlreadyConnected creates a new ErrAlreadyConnected error.
func newErrAlreadyConnected(id int) *ErrAlreadyConnected {
	return &ErrAlreadyConnected{errBase: errBase{ID: id}}
}

// newErrNotConnected creates a new ErrNotConnected error.
func newErrNotConnected(id int) *ErrNotConnected {
	return &ErrNotConnected{errBase: errBase{ID: id}}
}

// Error implements the error interface.
// It returns the error message for the ErrNilSession error.
func (e ErrNilSession) Error() string {
	return "session is nil"
}

// Error implements the error interface.
// It returns the error message for the ErrNotConnected error.
func (e ErrAlreadyConnected) Error() string {
	return fmt.Sprintf("shard %d is already connected", e.ID)
}

// Error implements the error interface.
// It returns the error message for the ErrNotConnected error.
func (e ErrNotConnected) Error() string {
	return fmt.Sprintf("shard %d is not connected", e.ID)
}

// Error implements the error interface.
// It returns the error message for the ErrShutdown error.
func (e ErrShutdown) Error() string {
	return fmt.Sprintf("failed to shut down all shards: %v", e.Err)
}

// Is implements the errors.Is interface.
// It returns true if the target error is of the same type as the receiver.
func (e ErrShutdown) Is(target error) bool {
	_, ok := target.(*ErrShutdown)
	return ok
}
