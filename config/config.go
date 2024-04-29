package config

import (
	"context"

	"github.com/lvlcn-t/go-kit/config"
)

type Config struct {
	// TODO: add actual configuration fields
	Host string `mapstructure:"host"`
}

func (c Config) IsEmpty() bool {
	return c == (Config{})
}

func (c Config) Validate(ctx context.Context) error {
	return nil
}

func Load(path string) (Config, error) {
	return config.Load[Config](path)
}
