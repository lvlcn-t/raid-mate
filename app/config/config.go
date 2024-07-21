package config

import (
	"errors"
	"reflect"

	"github.com/lvlcn-t/go-kit/apimanager"
	"github.com/lvlcn-t/go-kit/config"
	"github.com/lvlcn-t/raid-mate/app/bot"
	"github.com/lvlcn-t/raid-mate/app/database"
	"github.com/lvlcn-t/raid-mate/app/services"
)

var _ config.Loadable = (*Config)(nil)

// Config is the configuration for the application.
type Config struct {
	// Bot is the configuration for the bot.
	Bot bot.Config `yaml:"bot" mapstructure:"bot"`
	// Services is the configuration for the services.
	Services services.Config `yaml:"services" mapstructure:"services"`
	// API is the configuration for the API server.
	API apimanager.Config `yaml:"api" mapstructure:"api"`
	// Database is the configuration for the database.
	Database database.Config `yaml:"database" mapstructure:"database"`
	// version is the version of the application.
	Version string `yaml:"-" mapstructure:"-"`
}

// IsEmpty returns whether the configuration is empty.
// It implements the config.Settings interface.
func (c *Config) IsEmpty() bool {
	return reflect.DeepEqual(c, &Config{})
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	return errors.Join(c.Bot.Validate(), c.Services.Validate(), c.API.Validate(), c.Database.Validate())
}

// Load loads the configuration from the given path and validates it.
func Load(path string) (*Config, error) {
	config.SetName("raidmate")
	cfg, err := config.Load[*Config](path)
	if err != nil {
		return nil, err
	}

	err = config.Validate(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
