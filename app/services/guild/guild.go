package guild

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/disgoorg/snowflake/v2"
	"github.com/lvlcn-t/raid-mate/app/database/repo"
)

// Service is the interface for the guild service.
type Service interface {
	guildService
	credentialService
	reportService
	profileService
}

type guildService interface {
	// List returns a list of guilds.
	List(ctx context.Context) ([]repo.Guild, error)
	// Get returns the guild with the given ID.
	Get(ctx context.Context, id snowflake.ID) (repo.Guild, error)
	// Create creates a new guild.
	Create(ctx context.Context, ngp repo.NewGuildParams) error
	// Update updates the guild with the given parameters.
	Update(ctx context.Context, ugp repo.UpdateGuildParams) error
	// Delete deletes the guild with the given ID.
	Delete(ctx context.Context, id snowflake.ID) error
}

type credentialService interface {
	// GetCredentials returns the credentials for the given parameters.
	GetCredentials(ctx context.Context, gcp repo.GetCredentialsParams) (repo.Credential, error)
	// SetCredentials sets the credentials for the given parameters.
	SetCredentials(ctx context.Context, scp repo.SetCredentialsParams) error
}

type reportService interface {
	// GetReports returns the reports for the given guild and date.
	GetReports(ctx context.Context, guildID snowflake.ID, date time.Time) ([]string, error)
}

type profileService interface {
	// GetProfile returns the profile for the given parameters.
	GetProfile(ctx context.Context, req *RequestProfile) (*Profiles, error)
}

// RequestProfile is the request for the profile.
type RequestProfile struct {
	Type    string
	GuildID snowflake.ID
	User    string
	guild   repo.Guild
}

// guild implements [Service] for the guild service.
type guild struct {
	// database is the database repository.
	database repo.DBTX
	// client is the http client.
	client *client
}

// Config is the configuration for the guild service.
type Config struct {
	// Client is the configuration for the client.
	Client struct {
		// Token is the token for the client.
		Token string `yaml:"token" mapstructure:"token" validate:"required"`
		// Timeout is the timeout for the client.
		Timeout time.Duration `yaml:"timeout" mapstructure:"timeout" validate:"gte=0"`
	} `yaml:"client" mapstructure:"client"`
}

// NewService creates a new guild service.
func NewService(c *Config, db *sql.DB) Service {
	return &guild{
		database: db,
		client:   NewClient(c.Client.Token, c.Client.Timeout),
	}
}

func (s *guild) List(ctx context.Context) ([]repo.Guild, error) {
	return repo.New(s.database).ListGuilds(ctx)
}

func (s *guild) Get(ctx context.Context, id snowflake.ID) (repo.Guild, error) {
	return repo.New(s.database).GetGuild(ctx, int64(id)) //nolint:gosec // Snowflake cannot overflow AFAIK
}

func (s *guild) Create(ctx context.Context, ngp repo.NewGuildParams) error {
	return repo.New(s.database).NewGuild(ctx, ngp)
}

func (s *guild) Update(ctx context.Context, ugp repo.UpdateGuildParams) error {
	return repo.New(s.database).UpdateGuild(ctx, ugp)
}

func (s *guild) Delete(ctx context.Context, id snowflake.ID) error {
	return repo.New(s.database).DeleteGuild(ctx, int64(id)) //nolint:gosec // Snowflake cannot overflow AFAIK
}

func (s *guild) GetCredentials(ctx context.Context, gcp repo.GetCredentialsParams) (repo.Credential, error) {
	return repo.New(s.database).GetCredentials(ctx, gcp)
}

func (s *guild) SetCredentials(ctx context.Context, scp repo.SetCredentialsParams) error {
	return repo.New(s.database).SetCredentials(ctx, scp)
}

func (s *guild) GetReports(ctx context.Context, guildID snowflake.ID, date time.Time) ([]string, error) {
	guild, err := repo.New(s.database).GetGuild(ctx, int64(guildID)) //nolint:gosec // Snowflake cannot overflow AFAIK
	if err != nil {
		return nil, fmt.Errorf("error getting guild: %w", err)
	}

	reports, err := s.client.FetchReports(ctx, guild, date)
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
	guild, err := s.Get(ctx, req.GuildID)
	if err != nil {
		return nil, fmt.Errorf("error getting guild: %w", err)
	}
	req.guild = guild

	return s.client.FetchProfile(ctx, req)
}
