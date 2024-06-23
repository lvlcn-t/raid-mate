package bot

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/disgoorg/disgo"
	disbot "github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/gateway"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/disgo/sharding"
	"github.com/gofiber/fiber/v3"
	"github.com/lvlcn-t/loggerhead/logger"
	"github.com/lvlcn-t/raid-mate/app/bot/colors"
	"github.com/lvlcn-t/raid-mate/app/bot/commands"
	"github.com/lvlcn-t/raid-mate/app/services"
)

var _ Bot = (*bot)(nil)

// Bot is the interface for the bot.
type Bot interface {
	// Run starts the bot and blocks until it is stopped.
	Run(ctx context.Context) error
	// Shutdown stops the bot.
	Shutdown(ctx context.Context) error
	// Router returns the router for the bot's API.
	Router() fiber.Router
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
		err = errors.New("bot.token is required")
	}
	return errors.Join(err, c.Intents.Validate())
}

// bot is the implementation of the Bot interface.
type bot struct {
	// cfg is the bot configuration.
	cfg Config
	// commands is the collection of commands.
	commands *commands.Collection
	// services is the collection of services.
	services *services.Collection
	// conn is the Discord connection.
	conn disbot.Client
	// app is the bot's application info.
	app *discord.Application
	// done is the channel for when the bot is done.
	done chan struct{}
}

// New creates a new bot instance.
func New(cfg Config, svcs *services.Collection) Bot {
	return &bot{
		cfg:      cfg,
		commands: commands.NewCollection(svcs),
		services: svcs,
		conn:     nil,
		app:      nil,
		done:     make(chan struct{}, 1),
	}
}

// Run starts the bot and blocks until it is stopped.
func (b *bot) Run(ctx context.Context) error {
	ctx, cancel := logger.NewContextWithLogger(ctx)
	defer cancel()
	log := logger.FromContext(ctx)

	err := b.launchBot(ctx)
	defer func() {
		if b.conn != nil {
			b.conn.Close(ctx)
		}
	}()
	if err != nil {
		log.ErrorContext(ctx, "Failed to start bot", "error", err)
		return err
	}

	select {
	case <-ctx.Done():
		return b.Shutdown(ctx)
	case <-b.done:
		return nil
	}
}

// Shutdown shuts down the bot.
func (b *bot) Shutdown(ctx context.Context) error {
	// This defer is necessary to ensure that the bot is shut down properly
	// if the shutdown was triggered by calling Shutdown instead of per context cancellation.
	defer func() {
		b.done <- struct{}{}
		close(b.done)
	}()
	if errors.Is(ctx.Err(), context.Canceled) {
		return nil
	}

	return ctx.Err()
}

// Router returns the router for the bot's API.
func (b *bot) Router() fiber.Router {
	return b.commands.Router()
}

// launchBot starts the bot and registers its commands.
func (b *bot) launchBot(ctx context.Context) error {
	log := logger.FromContext(ctx)
	err := b.newConnection(ctx)
	if err != nil {
		log.ErrorContext(ctx, "Failed to create connection", "error", err)
		return err
	}

	b.app, err = b.conn.Rest().GetBotApplicationInfo(rest.WithCtx(ctx))
	if err != nil {
		log.ErrorContext(ctx, "Failed to get bot application info", "error", err)
		return err
	}

	err = b.registerCommands(ctx)
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
		disbot.WithEventListeners(b.newEventListeners(ctx)),
		disbot.WithLogger(log.ToSlog()),
	)
	return err
}

// registerCommands registers the bot's commands with Discord.
func (b *bot) registerCommands(ctx context.Context) error {
	infos := b.commands.Infos()
	_, err := b.conn.Rest().SetGlobalCommands(b.conn.ApplicationID(), infos, rest.WithCtx(ctx))
	if err != nil {
		for _, info := range infos {
			_, cErr := b.conn.Rest().CreateGlobalCommand(b.conn.ApplicationID(), info, rest.WithCtx(ctx))
			if cErr != nil {
				err = errors.Join(err, cErr)
			}
		}
	}
	return err
}

// newEventListeners creates the event listeners for the bot.
func (b *bot) newEventListeners(ctx context.Context) *events.ListenerAdapter {
	log := logger.FromContext(ctx)
	return &events.ListenerAdapter{
		OnGuildReady: func(event *events.GuildReady) {
			log.InfoContext(ctx, "Guild ready", "guild", event.Guild.ID.String())
		},
		OnGuildsReady: func(event *events.GuildsReady) {
			log.InfoContext(ctx, "Guilds on shard ready", "shard", event.ShardID())
		},
		OnApplicationCommandInteraction: func(event *events.ApplicationCommandInteractionCreate) {
			log.DebugContext(ctx, "Command interaction", "command", event.Data.CommandName())
			cmd := b.commands.GetAppCommand(event.Data.CommandName())
			if cmd != nil {
				cmd.Handle(ctx, event)
			}
		},
		OnGuildJoin: func(event *events.GuildJoin) {
			log.DebugContext(ctx, "Guild join", "guild", event.Guild.ID.String())
			b.handleGuildJoin(ctx, event)
		},
		OnComponentInteraction: func(event *events.ComponentInteractionCreate) {
			log.DebugContext(ctx, "Component interaction", "custom_id", event.Data.CustomID())
			cmd := b.commands.GetComponentCommand(event.Data.CustomID())
			if cmd != nil {
				cmd.Handle(ctx, event)
			}
		},
		OnModalSubmit: func(event *events.ModalSubmitInteractionCreate) {
			log.DebugContext(ctx, "Modal submit", "custom_id", event.Data.CustomID)
			cmd := b.commands.GetComponentCommand(event.Data.CustomID)
			if cmd != nil {
				cmd.HandleSubmission(ctx, event)
			}
		},
	}
}

// handleGuildJoin sends a welcome message to a guild when the bot joins it.
func (b *bot) handleGuildJoin(ctx context.Context, event *events.GuildJoin) {
	log := logger.FromContext(ctx)
	_, err := b.services.Guild.Get(ctx, event.GuildID)
	if !errors.Is(err, sql.ErrNoRows) {
		if err != nil {
			log.ErrorContext(ctx, "Failed to get guild", "error", err)
		}
		return
	}

	appIcon := b.app.Bot.EffectiveAvatarURL()
	guildIcon := event.Guild.IconURL()
	if guildIcon == nil {
		guildIcon = &appIcon
	}

	log.InfoContext(ctx, "Joined new guild", "guild", event.Guild.ID.String(), "app_icon", appIcon, "guild_icon", *guildIcon)
	embed := discord.NewEmbedBuilder().
		SetTitle("Welcome to Raid Mate!").
		SetDescription("Hello! I'm Raid Mate, your friendly raid bot. Let's get your guild set up. Click the button below to get started.").
		SetColor(colors.Blue.Int()).
		SetThumbnail(*guildIcon).
		AddField("Getting Started", "Click the button below to configure your guild.", false).
		SetFooter(b.app.Name, b.app.Bot.EffectiveAvatarURL()).
		SetTimestamp(time.Now()).
		Build()

	_, err = event.Client().Rest().CreateMessage(*event.Guild.SystemChannelID, discord.NewMessageCreateBuilder().
		AddEmbeds(embed).
		AddActionRow(discord.NewPrimaryButton("Set me up!", "guild")).
		Build(), rest.WithCtx(ctx))
	if err != nil {
		log.ErrorContext(ctx, "Failed to send message", "error", err)
	}
}
