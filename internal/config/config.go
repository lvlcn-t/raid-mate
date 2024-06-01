package config

import (
	"context"
	"reflect"

	"github.com/lvlcn-t/go-kit/config"
	"github.com/lvlcn-t/raid-mate/internal/bot"
)

// Config is the configuration for the application.
type Config struct {
	// Bot is the configuration for the bot.
	Bot bot.Config `yaml:"bot" mapstructure:"bot"`
}

// IsEmpty returns whether the configuration is empty.
// It implements the config.Settings interface.
func (c Config) IsEmpty() bool {
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
