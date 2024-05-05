package bot

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/lvlcn-t/raid-mate/internal/bot/commands"
	"github.com/lvlcn-t/raid-mate/internal/bot/manager"
	"github.com/lvlcn-t/raid-mate/internal/services"
)

const shutdownTimeout = 40 * time.Second

var _ Bot = (*bot)(nil)

// Bot is the interface for the bot.
type Bot interface {
	// Run starts the bot and blocks until it is stopped.
	Run(ctx context.Context) error
	// Shutdown stops the bot.
	Shutdown(ctx context.Context) error
}

// Config is the configuration for the bot.
type Config struct {
	// Token is the Discord bot token.
	Token string `yaml:"token" mapstructure:"token"`
	// Intents is the list of intents the bot should use.
	Intents IntentsConfig `yaml:"intents" mapstructure:"intents"`
}

// bot is the implementation of the Bot interface.
type bot struct {
	// cfg is the bot configuration.
	cfg Config
	// manager is the shard manager that holds all the shards.
	manager *manager.Manager
	// services is the collection of services.
	services services.Collection
	// done is the channel to signal the bot is done.
	done chan struct{}
}

// New creates a new bot instance.
func New(cfg Config) (Bot, error) {
	svcs, err := services.NewCollection()
	if err != nil {
		return nil, err
	}

	sess, err := discordgo.New(fmt.Sprintf("Bot %s", cfg.Token))
	if err != nil {
		return nil, err
	}

	mgr, err := manager.New(sess)
	if err != nil {
		return nil, err
	}
	mgr.SetIntent(cfg.Intents.List()[0]) // TODO: set the intents properly

	err = registerCommandsAndHandlers(svcs, mgr)
	if err != nil {
		return nil, err
	}

	return &bot{
		cfg:      cfg,
		manager:  mgr,
		services: svcs,
		done:     make(chan struct{}, 1),
	}, nil
}

// Run starts the bot and blocks until it is stopped.
func (b *bot) Run(ctx context.Context) error {
	err := b.services.Connect()
	if err != nil {
		return err
	}

	err = b.manager.Start(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return b.Shutdown(ctx)
		case <-b.done:
			return nil
		}
	}
}

// Shutdown stops the bot and all its components.
func (b *bot) Shutdown(ctx context.Context) error {
	defer close(b.done)
	errs := &ErrShutdown{
		ctxErr: ctx.Err(),
	}

	ctx, cancel := context.WithTimeout(ctx, shutdownTimeout)
	defer cancel()

	errs.mgrErr = b.manager.Shutdown(ctx)
	errs.svcErr = b.services.Close()

	if errs.HasErrors() {
		return errs
	}

	// Send the done signal to shutdown the bot if the shutdown wasn't caused by an error.
	b.done <- struct{}{}
	return nil
}

// registerCommandsAndHandlers registers all commands and their respective handlers,
// and ensures that all handlers are added.
func registerCommandsAndHandlers(svcs services.Collection, mgr *manager.Manager) (err error) {
	cmdFactory := commands.NewFactory()
	allCommands := cmdFactory.Commands(svcs)

	for _, cmd := range allCommands {
		// Always add handlers for all commands
		mgr.AddHandlers(manager.EventHandler{Handler: cmd})

		icmds, ok := cmd.([]commands.InteractionCommand)
		if !ok {
			continue
		}

		if sErr := syncApplicationCommands(icmds, mgr); sErr != nil {
			err = errors.Join(err, fmt.Errorf("failed to sync application commands: %w", sErr))
		}
	}

	return err
}

// syncApplicationCommands syncs the application commands to all guilds.
func syncApplicationCommands(icmds []commands.InteractionCommand, mgr *manager.Manager) (err error) {
	infos := make([]*discordgo.ApplicationCommand, 0, len(icmds))
	for i, icmd := range icmds {
		infos[i] = icmd.Info()
	}

	for _, guild := range mgr.ListGuilds() {
		if rErr := mgr.RegisterCommandsOverwrite(guild.ID, infos); rErr != nil {
			err = errors.Join(err, fmt.Errorf("failed to register commands for guild %s: %w", guild.ID, rErr))
		}
	}

	return err
}
