package api

type ErrAlreadyRunning struct{}

func (e *ErrAlreadyRunning) Error() string {
	return "cannot mount routes while the server is running"
}

func (e *ErrAlreadyRunning) Is(err error) bool {
	_, ok := err.(*ErrAlreadyRunning)
	return ok
}
