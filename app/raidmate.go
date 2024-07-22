package app

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/lvlcn-t/go-kit/apimanager"
	"github.com/lvlcn-t/go-kit/apimanager/middleware"
	"github.com/lvlcn-t/loggerhead/logger"
	"github.com/lvlcn-t/raid-mate/app/bot"
	"github.com/lvlcn-t/raid-mate/app/config"
	"github.com/lvlcn-t/raid-mate/app/database"
	"github.com/lvlcn-t/raid-mate/app/services"
)

const shutdownTimeout = 60 * time.Second

type RaidMate struct {
	// config is the configuration for the application.
	config *config.Config
	// api is the API server.
	api apimanager.Server
	// bot is the bot for the application.
	bot bot.Bot
	// services is the collection of services.
	services *services.Collection
	// errCh is the channel for errors.
	errCh chan error
	// once is used to ensure that the application is only shutdown once.
	once sync.Once
}

// New creates a new application.
func New(cfg *config.Config) (*RaidMate, error) {
	db, err := database.New(&cfg.Database)
	if err != nil {
		return nil, err
	}

	r := &RaidMate{
		config:   cfg,
		bot:      nil,
		services: services.NewCollection(&cfg.Services, db),
		api:      apimanager.New(&cfg.API, middleware.Logger("/healthz"), middleware.Recover()),
		errCh:    make(chan error, 1),
		once:     sync.Once{},
	}

	r.bot = bot.New(cfg.Bot, r.services)
	err = r.api.MountGroup(apimanager.RouteGroup{
		Path: "/v1",
		App:  r.bot.Router(),
	})
	if err != nil {
		return nil, err
	}

	return r, nil
}

// Run starts the application and blocks until it is stopped.
func (r *RaidMate) Run(ctx context.Context) error {
	log := logger.FromContext(ctx)

	go func() {
		err := r.api.Run(ctx)
		if err != nil {
			log.ErrorContext(ctx, "Failed to start API server", "error", err)
			r.errCh <- err
		}
	}()

	go func() {
		err := r.bot.Run(ctx)
		if err != nil {
			log.ErrorContext(ctx, "Failed to start bot", "error", err)
			r.errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		return r.Shutdown(ctx)
	case err := <-r.errCh:
		return errors.Join(err, r.Shutdown(ctx))
	}
}

func (r *RaidMate) Shutdown(ctx context.Context) error {
	var errs *errShutdown
	r.once.Do(func() {
		defer func() {
			r.errCh <- errs
			close(r.errCh)
		}()

		errs = &errShutdown{}
		if !errors.Is(ctx.Err(), context.Canceled) {
			errs.ctxErr = ctx.Err()
		}

		c, cancel := context.WithTimeout(ctx, shutdownTimeout)
		defer cancel()

		errs.apiErr = r.api.Shutdown(c)
		errs.botErr = r.bot.Shutdown(c)
	})
	if errs != nil && errs.HasErrors() {
		return errs
	}

	return nil
}
