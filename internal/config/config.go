package config

import (
	"context"
	"reflect"

	"github.com/lvlcn-t/go-kit/config"
	"github.com/lvlcn-t/raid-mate/internal/bot"
	"github.com/lvlcn-t/raid-mate/internal/services"
)

// Config is the configuration for the application.
type Config struct {
	// Bot is the configuration for the bot.
	Bot bot.Config `yaml:"bot" mapstructure:"bot"`
	// Services is the configuration for the services.
	Services services.Config `yaml:"services" mapstructure:"services"`
}

// IsEmpty returns whether the configuration is empty.
// It implements the config.Settings interface.
func (c Config) IsEmpty() bool { //nolint:gocritic // The viper cannot handle pointer receivers
	return reflect.DeepEqual(c, Config{})
}

// Validate validates the configuration.
func (c *Config) Validate(_ context.Context) error {
	// TODO: Add validation.
	return nil
}

// Load loads the configuration from the given path.
func Load(path string) (Config, error) {
	config.SetBinaryName("raidmate")
	return config.Load[Config](path)
}
