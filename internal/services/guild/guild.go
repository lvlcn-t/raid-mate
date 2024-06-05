package guild

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type Service interface {
	Create(ctx context.Context, guildID string, guild Request) error
	GetCredentials(ctx context.Context, guildID, account string) (Credentials, error)
	SetCredentials(ctx context.Context, guildID string, credentials Credentials) error
	GetReport(ctx context.Context, guildID string, date time.Time) ([]string, error)
}

type Request struct {
	Name         string
	ServerName   string
	ServerRegion string
}

type guild struct {
	database sync.Map // TODO: use a postgreSQL database
	client   *client
}

type Config struct {
	// Client is the configuration for the client.
	Client struct {
		// Token is the token for the client.
		Token string
		// Timeout is the timeout for the client.
		Timeout time.Duration
	}
}

func NewService(c *Config) (*guild, error) {
	return &guild{
		database: sync.Map{},
		client:   NewClient(c.Client.Token, c.Client.Timeout),
	}, nil
}

type Credentials struct {
	Name     string
	Url      string
	Username string
	Password string
}

func (s *guild) Create(_ context.Context, guildID string, guild Request) error {
	// TODO: implement this with the postgreSQL database
	s.database.Store(guildID, guild)
	return nil
}

func (s *guild) GetCredentials(_ context.Context, guildID, account string) (Credentials, error) {
	// TODO: implement this with the postgreSQL database
	val, ok := s.database.Load(fmt.Sprintf("%s:%s", guildID, account))
	if !ok {
		return Credentials{}, errors.New("not found")
	}
	return val.(Credentials), nil
}

func (s *guild) SetCredentials(_ context.Context, guildID string, credentials Credentials) error {
	// TODO: implement this with the postgreSQL database
	key := fmt.Sprintf("%s:%s", guildID, credentials.Name)
	s.database.Store(key, credentials)
	return nil
}

func (s *guild) GetReport(ctx context.Context, guildID string, date time.Time) ([]string, error) {
	guild, ok := s.database.Load(guildID) // TODO: implement this with the postgreSQL database
	if !ok {
		return nil, errors.New("not found")
	}

	reports, err := s.client.FetchReport(ctx, guild.(Request), date)
	if err != nil {
		return nil, fmt.Errorf("error fetching reports: %w", err)
	}

	var reportUrls []string
	for _, r := range reports {
		reportUrls = append(reportUrls, fmt.Sprintf("%s/reports/%s", baseURL, r.Id))
	}

	return reportUrls, nil
}
