package services

import (
	"database/sql"

	"github.com/lvlcn-t/raid-mate/app/services/feedback"
	"github.com/lvlcn-t/raid-mate/app/services/guild"
)

// Collection is the collection of services.
type Collection struct {
	Feedback feedback.Service
	Guild    guild.Service
}

// Config is the configuration for the services.
type Config struct {
	// Feedback is the configuration for the feedback service.
	Feedback feedback.Config `yaml:"feedback" mapstructure:"feedback" validate:"required"`
	// Guild is the configuration for the guild service.
	Guild guild.Config `yaml:"guild" mapstructure:"guild" validate:"required"`
}

// NewCollection creates a new collection of services.
func NewCollection(c *Config, db *sql.DB) *Collection {
	return &Collection{
		Feedback: feedback.NewService(&c.Feedback),
		Guild:    guild.NewService(&c.Guild, db),
	}
}
