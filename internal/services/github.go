package services

import "context"

type GitHub interface {
	CreateIssue(ctx context.Context, feedback string) error
}

type github struct{}

func NewGitHubService() (GitHub, error) {
	return &github{}, nil
}

func (s *github) CreateIssue(_ context.Context, _ string) error {
	// TODO: do the grpc call to the github service
	return nil
}
