package feedback

import (
	"context"
	"strings"

	"github.com/disgoorg/disgo/bot"
)

// Service is the interface for the service.
type Service interface {
	// Submit submits the feedback.
	Submit(ctx context.Context, req Request, client bot.Client) error
}

// Request is the request for the feedback.
type Request struct {
	Feedback string `json:"feedback"`
	Server   string `json:"server"`
	User     string `json:"user"`
}

// feedback is the service for the feedback.
// It is a composite service that can use multiple services.
type feedback struct {
	// selected is the selected feedback service.
	selected string
	// registry is the registry of the services.
	registry map[string]Service
}

// Config is the configuration for the service.
type Config struct {
	// Service is the service to use.
	// It can be "github", "dm", or "all".
	// If it is not set, no service will be used.
	Service string `yaml:"service" mapstructure:"service"`
	// GitHub is the configuration for the GitHub service.
	GitHub githubConfig `yaml:"github" mapstructure:"github"`
	// DM is the configuration for the DM service.
	DM dmConfig `yaml:"dm" mapstructure:"dm"`
}

// NewService creates a new feedback service.
func NewService(c *Config) (Service, error) {
	return &feedback{
		selected: c.Service,
		registry: map[string]Service{
			"github": newGitHub(&c.GitHub),
			"dm":     newDM(&c.DM),
		},
	}, nil
}

// Submit submits the feedback.
func (s *feedback) Submit(ctx context.Context, req Request, client bot.Client) error {
	if s.selected == "" {
		return nil
	}

	if strings.EqualFold(s.selected, "all") {
		for _, svc := range s.registry {
			return svc.Submit(ctx, req, client)
		}
	}

	if svc, ok := s.registry[s.selected]; ok {
		return svc.Submit(ctx, req, client)
	}

	return &ErrUnknownService{s.selected}
}
