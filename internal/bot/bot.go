package bot

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/disgoorg/disgo"
	disbot "github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/disgo/sharding"
	"github.com/lvlcn-t/loggerhead/logger"
	"github.com/lvlcn-t/raid-mate/internal/api"
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

func (c *Config) Validate() error {
	var err error
	if c.Token == "" {
		err = errors.New("token is required")
	}
	return errors.Join(err, c.Intents.Validate())
}

// bot is the implementation of the Bot interface.
type bot struct {
	// cfg is the bot configuration.
	cfg Config
	// api is the API server.
	api api.Server
	// commands is the collection of commands.
	commands commands.Collection
	// services is the collection of services.
	services services.Collection
	// conn is the Discord connection.
	conn disbot.Client
	// errCh is the channel to signal errors.
	errCh chan error
	// done is the channel to signal the bot is done.
	done chan struct{}
	// once is the sync.Once to ensure the bot is only shutdown once.
	once sync.Once
}

// New creates a new bot instance.
func New(cfg Config, svcs services.Collection) (Bot, error) {
	b := &bot{
		cfg:      cfg,
		api:      api.NewServer(&api.Config{Address: ":8080"}), // TODO: move api server to dedicated layer to avoid circular dependency
		commands: commands.NewCollection(svcs),
		services: svcs,
		conn:     nil,
		errCh:    make(chan error, 1),
		done:     make(chan struct{}, 1),
		once:     sync.Once{},
	}

	err := b.api.Mount(api.RouteGroup{
		Path: "/v1",
		App:  b.commands.Router(),
	})
	if err != nil {
		return nil, err
	}

	return b, nil
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

	err = b.startBot(ctx)
	defer func() {
		if b.conn != nil {
			b.conn.Close(ctx)
		}
	}()
	if err != nil {
		log.ErrorContext(ctx, "Failed to start bot", "error", err)
		return err
	}

	go func() {
		err = b.api.Run(ctx)
		if err != nil {
			log.ErrorContext(ctx, "Failed to start API server", "error", err)
			b.errCh <- err
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return b.Shutdown(ctx)
		case <-b.done:
			return nil
		case err := <-b.errCh:
			if err != nil {
				return errors.Join(err, b.Shutdown(ctx))
			}
		}
	}
}

// Shutdown stops the bot and all its components.
func (b *bot) Shutdown(ctx context.Context) error {
	var errs *ErrShutdown

	b.once.Do(func() {
		defer close(b.done)
		defer close(b.errCh)
		errs = &ErrShutdown{
			ctxErr: ctx.Err(),
		}

		c, cancel := context.WithTimeout(ctx, shutdownTimeout)
		defer cancel()

		errs.svcErr = b.services.Close(c)
		errs.apiErr = b.api.Shutdown(c)

		if !errs.HasErrors() {
			// Send the done signal to shutdown the bot if the shutdown wasn't caused by an error.
			b.done <- struct{}{}
		}
	})

	if errs != nil && errs.HasErrors() {
		return errs
	}

	return nil
}

func (b *bot) startBot(ctx context.Context) error {
	log := logger.FromContext(ctx)
	err := b.newConnection(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to create connection", "error", err)
		return err
	}

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

	return nil
}

// newConnection creates a new Discord connection.
func (b *bot) newConnection(ctx context.Context) (err error) {
	log := logger.FromContext(ctx)
	b.conn, err = disgo.New(b.cfg.Token,
		disbot.WithShardManagerConfigOpts(
			sharding.WithShardIDs(0, 1),
			sharding.WithShardCount(2),
			sharding.WithAutoScaling(true),
			sharding.WithGatewayConfigOpts(
				gateway.WithIntents(b.cfg.Intents.List()...),
				gateway.WithCompress(true),
				gateway.WithPresenceOpts(
					gateway.WithOnlineStatus(discord.OnlineStatusOnline),
					gateway.WithWatchingActivity("you"),
				),
			),
		),
		disbot.WithEventListeners(&events.ListenerAdapter{
			OnGuildReady: func(event *events.GuildReady) {
				log.InfoContext(ctx, "Guild ready", "guild", event.Guild.ID.String())
				log.InfoContext(ctx, "Joined new guild", "guild", event.Guild.ID.String())
				err = b.sendRegistrationForm(ctx, event)
				if err != nil {
					log.ErrorContext(ctx, "Failed to send registration form", "error", err)
				}
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
			OnGuildJoin: func(_ *events.GuildJoin) {
			},
			OnComponentInteraction: func(event *events.ComponentInteractionCreate) {
				if err = b.handleFormSubmission(ctx, event); err != nil {
					log.ErrorContext(ctx, "Failed to handle form submission", "error", err)
				}
			},
		}),
		disbot.WithLogger(log.ToSlog()),
	)
	return err
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

// sendRegistrationForm sends a registration form to the specified guild.
func (b *bot) sendRegistrationForm(ctx context.Context, event *events.GuildReady) error {
	log := logger.FromContext(ctx)

	_, err := b.conn.Rest().CreateMessage(*event.Guild.SystemChannelID, discord.NewMessageCreateBuilder().
		AddActionRow(discord.NewTextInput("guild_name", discord.TextInputStyleShort, "Name of the Guild").WithRequired(true)).
		AddActionRow(discord.NewTextInput("guild_name", discord.TextInputStyleShort, "Name of the Guild").WithRequired(true)).
		AddActionRow(discord.NewTextInput("guild_realm", discord.TextInputStyleShort, "Server of the Guild").WithRequired(true)).
		AddActionRow(discord.NewTextInput("guild_region", discord.TextInputStyleShort, "Region of the Guild (EU, US, etc.)").WithRequired(true)).
		AddActionRow(discord.NewTextInput("guild_faction", discord.TextInputStyleShort, "Faction of the Guild (Alliance, Horde, etc.)").WithRequired(true)).
		Build(), rest.WithCtx(ctx))
	if err != nil {
		log.ErrorContext(ctx, "Failed to send registration form", "error", err)
		return err
	}
	return nil
}

// handleFormSubmission processes the form data when a form is submitted.
func (b *bot) handleFormSubmission(ctx context.Context, event *events.ComponentInteractionCreate) error { //nolint:unparam // TODO: implement this
	log := logger.FromContext(ctx)
	data := event.Data.CustomID()
	log.InfoContext(ctx, "Form submitted", "data", data)
	return nil
}
