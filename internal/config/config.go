package config

import (
	"context"

	"github.com/lvlcn-t/go-kit/config"
	"github.com/lvlcn-t/raid-mate/internal/bot"
)

type Config struct {
	Bot bot.Config `yaml:"bot" mapstructure:"bot"`
}

func (c *Config) IsEmpty() bool {
	return c == (&Config{})
}

func (c Config) Validate(_ context.Context) error {
	return nil
}

func Load(path string) (*Config, error) {
	return config.Load[*Config](path)
}
