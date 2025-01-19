package config

import (
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
	Bot bot.Config `yaml:"bot" mapstructure:"bot" validate:"required"`
	// Services is the configuration for the services.
	Services services.Config `yaml:"services" mapstructure:"services" validate:"required"`
	// API is the configuration for the API server.
	API apimanager.Config `yaml:"api" mapstructure:"api" validate:"required"`
	// Database is the configuration for the database.
	Database database.Config `yaml:"database" mapstructure:"database" validate:"required"`
	// version is the version of the application.
	Version string `yaml:"-" mapstructure:"-" validate:"-"`
}

// IsEmpty returns whether the configuration is empty.
// It implements the config.Settings interface.
func (c *Config) IsEmpty() bool {
	return reflect.DeepEqual(c, &Config{})
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
