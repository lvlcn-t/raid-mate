package feedback

import (
	"context"
	"errors"
	"slices"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/snowflake/v2"
)

// Service is the interface for the service.
type Service interface {
	// Submit submits the feedback.
	Submit(ctx context.Context, req Request, client bot.Client) error
}

// Request is the request for the feedback.
type Request struct {
	Feedback string       `json:"feedback"`
	Server   string       `json:"server"`
	Username string       `json:"username"`
	UserID   snowflake.ID `json:"user_id"`
}

// feedback is the service for the feedback.
// It is a composite service that can use multiple services.
type feedback struct {
	// selected is the selected feedback service.
	selected []string
	// registry is the registry of the services.
	registry map[string]Service
}

// Config is the configuration for the service.
type Config struct {
	// Service is the service to use.
	// It can be "github", "dm", or "all".
	// If it is not set, no service will be used.
	Service []string `yaml:"service" mapstructure:"service"`
	// GitHub is the configuration for the GitHub service.
	GitHub githubConfig `yaml:"github" mapstructure:"github"`
	// DM is the configuration for the DM service.
	DM dmConfig `yaml:"dm" mapstructure:"dm"`
}

func (c *Config) Validate() error {
	return errors.Join(c.GitHub.Validate(), c.DM.Validate())
}

// NewService creates a new feedback service.
func NewService(c *Config) Service {
	return &feedback{
		selected: c.Service,
		registry: map[string]Service{
			"github": newGitHub(&c.GitHub),
			"dm":     newDM(&c.DM),
		},
	}
}

// Submit submits the feedback.
func (s *feedback) Submit(ctx context.Context, req Request, client bot.Client) error {
	if len(s.selected) == 0 {
		return nil
	}

	if slices.Contains(s.selected, "all") {
		for _, svc := range s.registry {
			if err := svc.Submit(ctx, req, client); err != nil {
				return err
			}
		}
		return nil
	}

	var unrecognized []string
	for _, svc := range s.selected {
		if fsvc, ok := s.registry[svc]; ok {
			if err := fsvc.Submit(ctx, req, client); err != nil {
				return err
			}
			continue
		}
		unrecognized = append(unrecognized, svc)
	}

	if len(unrecognized) > 0 {
		return &ErrUnrecognizedServices{services: unrecognized}
	}

	return nil
}
