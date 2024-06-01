package config

import (
	"context"
	"reflect"

	"github.com/lvlcn-t/go-kit/config"
	"github.com/lvlcn-t/raid-mate/internal/bot"
)

type Config struct {
	Bot bot.Config `yaml:"bot" mapstructure:"bot"`
}

func (c Config) IsEmpty() bool {
	return reflect.DeepEqual(c, Config{})
}

func (c *Config) Validate(_ context.Context) error {
	return nil
}

func Load(path string) (Config, error) {
	config.SetBinaryName("raidmate")
	return config.Load[Config](path)
}
