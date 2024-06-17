package services

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lvlcn-t/raid-mate/internal/services/feedback"
	"github.com/lvlcn-t/raid-mate/internal/services/guild"
)

// Collection is the collection of services.
type Collection struct {
	Feedback feedback.Service
	Guild    guild.Service
}

// Config is the configuration for the services.
type Config struct {
	// Feedback is the configuration for the feedback service.
	Feedback feedback.Config `yaml:"feedback" mapstructure:"feedback"`
	// Guild is the configuration for the guild service.
	Guild guild.Config `yaml:"guild" mapstructure:"guild"`
}

func (c *Config) Validate() error {
	err := errors.Join(c.Feedback.Validate())
	return errors.Join(err, c.Guild.Validate())
}

// NewCollection creates a new collection of services.
func NewCollection(c *Config, db *sql.DB) (Collection, error) {
	fb, err := feedback.NewService(&c.Feedback)
	if err != nil {
		return Collection{}, err
	}

	g, err := guild.NewService(&c.Guild, db)
	if err != nil {
		return Collection{}, err
	}

	return Collection{
		Feedback: fb,
		Guild:    g,
	}, nil
}

func (c *Collection) Connect() error {
	// TODO: connect all services
	return nil
}

func (c *Collection) Close(_ context.Context) error {
	// TODO: close all services
	return nil
}
