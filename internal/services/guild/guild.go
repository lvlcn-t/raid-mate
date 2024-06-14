package guild

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type Service interface {
	List(ctx context.Context) ([]Guild, error)
	Get(ctx context.Context, guildID string) (*Guild, error)
	Create(ctx context.Context, guildID string, guild *Guild) error
	Update(ctx context.Context, guildID string, guild *Guild) error
	Delete(ctx context.Context, guildID string) error
	GetCredentials(ctx context.Context, guildID, account string) (Credentials, error)
	SetCredentials(ctx context.Context, guildID string, credentials Credentials) error
	GetReports(ctx context.Context, guildID string, date time.Time) ([]string, error)
	GetProfile(ctx context.Context, profile *RequestProfile) (*Profiles, error)
}

type Guild struct {
	Name         string
	ServerName   string
	ServerRegion string
	ServerRealm  string
}

type RequestProfile struct {
	Type    string
	GuildID string
	User    string
	guild   Guild
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

func (s *guild) List(_ context.Context) ([]Guild, error) {
	var guilds []Guild
	s.database.Range(func(_, value any) bool {
		guilds = append(guilds, value.(Guild))
		return true
	})
	return guilds, nil
}

func (s *guild) Get(_ context.Context, guildID string) (*Guild, error) {
	guild, ok := s.database.Load(guildID) // TODO: implement this with the postgreSQL database
	if !ok {
		return nil, errors.New("not found")
	}
	g := guild.(Guild)
	return &g, nil
}

func (s *guild) Create(_ context.Context, guildID string, guild *Guild) error {
	// TODO: implement this with the postgreSQL database
	s.database.Store(guildID, *guild)
	return nil
}

func (s *guild) Update(_ context.Context, guildID string, guild *Guild) error {
	_, ok := s.database.Swap(guildID, *guild) // TODO: implement this with the postgreSQL database
	if !ok {
		return errors.New("not found")
	}
	return nil
}

func (s *guild) Delete(_ context.Context, guildID string) error {
	s.database.Delete(guildID) // TODO: implement this with the postgreSQL database
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

func (s *guild) GetReports(ctx context.Context, guildID string, date time.Time) ([]string, error) {
	guild, ok := s.database.Load(guildID) // TODO: implement this with the postgreSQL database
	if !ok {
		return nil, errors.New("not found")
	}

	reports, err := s.client.FetchReports(ctx, guild.(Guild), date)
	if err != nil {
		return nil, fmt.Errorf("error fetching reports: %w", err)
	}

	var reportUrls []string
	for _, r := range reports {
		reportUrls = append(reportUrls, fmt.Sprintf("%s/reports/%s", logsBaseURL, r.Id))
	}

	return reportUrls, nil
}

func (s *guild) GetProfile(ctx context.Context, req *RequestProfile) (*Profiles, error) {
	if req.Type == "guild" {
		guild, ok := s.database.Load(req.GuildID) // TODO: implement this with the postgreSQL database
		if !ok {
			return nil, errors.New("not found")
		}
		req.guild = guild.(Guild)
	}

	return s.client.FetchProfile(ctx, req)
}
