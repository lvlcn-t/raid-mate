package services

import (
	"context"
)

type Guild interface {
	GetCredentials(ctx context.Context, guildID, account string) (Credentials, error)
	SetCredentials(ctx context.Context, guildID string, credentials Credentials) error
}

type guild struct{}

func NewGuildService() (Guild, error) {
	return &guild{}, nil
}

type Credentials struct {
	Url      string
	Username string
	Password string
}

func (s *guild) GetCredentials(_ context.Context, _, _ string) (Credentials, error) {
	// TODO: do the grpc call to the guild service
	return Credentials{}, nil
}

func (s *guild) SetCredentials(_ context.Context, _ string, _ Credentials) error {
	// TODO: do the grpc call to the guild service
	return nil
}
