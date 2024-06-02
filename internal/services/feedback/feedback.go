package feedback

import (
	"context"
	"strings"
)

// Service is the interface for the service.
type Service interface {
	// Submit submits the feedback.
	Submit(ctx context.Context, feedback string) error
}

type feedback struct {
	selected string
	registry map[string]Service
}

// Config is the configuration for the service.
type Config struct {
	// Service is the service to use.
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

func (s *feedback) Submit(ctx context.Context, feedback string) error {
	if s.selected == "" {
		return nil
	}

	if strings.EqualFold(s.selected, "all") {
		for _, svc := range s.registry {
			return svc.Submit(ctx, feedback)
		}
	}

	if svc, ok := s.registry[s.selected]; ok {
		return svc.Submit(ctx, feedback)
	}

	return &ErrUnknownService{s.selected}
}
