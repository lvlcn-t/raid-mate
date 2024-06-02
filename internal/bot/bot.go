package bot

import (
	"context"
	"errors"
	"time"

	"github.com/disgoorg/disgo"
	disbot "github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/sharding"
	"github.com/lvlcn-t/loggerhead/logger"
	"github.com/lvlcn-t/raid-mate/internal/bot/commands"
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
	// commands is the collection of commands.
	commands commands.Collection
	// services is the collection of services.
	services services.Collection
	// conn is the Discord connection.
	conn disbot.Client
	// done is the channel to signal the bot is done.
	done chan struct{}
}

// New creates a new bot instance.
func New(cfg Config, svcs services.Collection) (Bot, error) {
	return &bot{
		cfg:      cfg,
		commands: commands.NewCollection(svcs),
		services: svcs,
		conn:     nil,
		done:     make(chan struct{}, 1),
	}, nil
}

// Run starts the bot and blocks until it is stopped.
func (b *bot) Run(ctx context.Context) error {
	ctx, cancel := logger.NewContextWithLogger(ctx)
	defer cancel()
	log := logger.FromContext(ctx)

	err := b.services.Connect()
	if err != nil {
		log.ErrorContext(ctx, "Failed to connect services", "error", err)
		return err
	}

	b.conn, err = b.newConnection(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to create connection", "error", err)
		return err
	}
	defer b.conn.Close(ctx)

	err = b.registerCommands()
	if err != nil {
		log.ErrorContext(ctx, "Failed to register commands", "error", err)
		return err
	}

	err = b.conn.OpenShardManager(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to open gateway", "error", err)
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

	errs.svcErr = b.services.Close(ctx)

	if errs.HasErrors() {
		return errs
	}

	// Send the done signal to shutdown the bot if the shutdown wasn't caused by an error.
	b.done <- struct{}{}
	return nil
}

// newConnection creates a new Discord connection.
func (b *bot) newConnection(ctx context.Context) (disbot.Client, error) {
	log := logger.FromContext(ctx)
	return disgo.New(b.cfg.Token,
		disbot.WithShardManagerConfigOpts(
			sharding.WithShardIDs(0, 1),
			sharding.WithShardCount(2),
			sharding.WithAutoScaling(true),
			sharding.WithGatewayConfigOpts(
				gateway.WithIntents(b.cfg.Intents.List()...),
				gateway.WithCompress(true),
			),
		),
		disbot.WithEventListeners(&events.ListenerAdapter{
			OnGuildReady: func(event *events.GuildReady) {
				log.InfoContext(ctx, "Guild ready", "guild", event.Guild.ID.String())
			},
			OnGuildsReady: func(event *events.GuildsReady) {
				log.InfoContext(ctx, "Guilds on shard ready", "shard", event.ShardID())
			},
			OnApplicationCommandInteraction: func(event *events.ApplicationCommandInteractionCreate) {
				cmd := b.commands.Get(event.Data.CommandName())
				if cmd != nil {
					cmd.Handle(ctx, event)
				}
			},
		}),
		disbot.WithLogger(log.ToSlog()),
	)
}

// registerCommands registers the bot's commands with Discord.
func (b *bot) registerCommands() error {
	infos := b.commands.Infos()
	_, err := b.conn.Rest().SetGlobalCommands(b.conn.ApplicationID(), infos)
	if err != nil {
		for _, info := range infos {
			_, cErr := b.conn.Rest().CreateGlobalCommand(b.conn.ApplicationID(), info)
			if cErr != nil {
				err = errors.Join(err, cErr)
			}
		}
	}
	return err
}
