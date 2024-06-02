package feedback

import "fmt"

// ErrUnknownService is the error for an unknown service.
type ErrUnknownService struct {
	service string
}

// Error returns the error message.
func (e *ErrUnknownService) Error() string {
	return fmt.Sprintf("unknown service: %s", e.service)
}

// Is checks if the target is an ErrUnknownService.
func (e *ErrUnknownService) Is(target error) bool {
	_, ok := target.(*ErrUnknownService)
	return ok
}
