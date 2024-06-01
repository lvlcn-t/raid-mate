package services

import (
	"context"

	"github.com/lvlcn-t/raid-mate/internal/services/github"
	"github.com/lvlcn-t/raid-mate/internal/services/guild"
)

// Collection is the collection of services.
type Collection struct {
	GitHub github.Service
	Guild  guild.Service
}

// NewCollection creates a new collection of services.
func NewCollection() (Collection, error) {
	gh, err := github.NewService()
	if err != nil {
		return Collection{}, err
	}

	g, err := guild.NewService()
	if err != nil {
		return Collection{}, err
	}

	return Collection{
		GitHub: gh,
		Guild:  g,
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
