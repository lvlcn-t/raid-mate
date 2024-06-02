package feedback

import "fmt"

// ErrUnrecognizedServices is the error for unrecognized services.
type ErrUnrecognizedServices struct {
	services []string
}

// Error returns the error message.
func (e *ErrUnrecognizedServices) Error() string {
	return fmt.Sprintf("no service was recognized: %v", e.services)
}

// Is checks if the target is an ErrUnknownService.
func (e *ErrUnrecognizedServices) Is(target error) bool {
	_, ok := target.(*ErrUnrecognizedServices)
	return ok
}
