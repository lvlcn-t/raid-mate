package guild

import (
	"context"
	"time"
)

type Service interface {
	GetCredentials(ctx context.Context, guildID, account string) (Credentials, error)
	SetCredentials(ctx context.Context, guildID string, credentials Credentials) error
	GetLogs(ctx context.Context, guildID string, date time.Time) ([]string, error)
}

type guild struct{}

func NewService() (*guild, error) {
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

func (s *guild) GetLogs(_ context.Context, _ string, _ time.Time) ([]string, error) {
	// TODO: do the grpc call to the guild service
	return nil, nil
}
