package github

import "context"

// Service is the interface for the service.
type Service interface {
	CreateIssue(ctx context.Context, feedback string) error
}

// github is the implementation of the Service interface for GitHub.
type github struct{}

// NewService creates a new GitHub service.
func NewService() (*github, error) {
	return &github{}, nil
}

// CreateIssue creates a new issue on GitHub.
func (s *github) CreateIssue(_ context.Context, _ string) error {
	// TODO: do the grpc call to the github service
	return nil
}
