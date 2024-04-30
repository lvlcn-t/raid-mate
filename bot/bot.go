package bot

import (
	"context"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

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
	Token string
}

// bot is the implementation of the Bot interface.
type bot struct {
	// session is the Discord session.
	session *discordgo.Session
	// cfg is the bot configuration.
	cfg Config
}

// New creates a new bot instance.
func New(cfg Config) (Bot, error) {
	sess, err := discordgo.New(fmt.Sprintf("Bot %s", cfg.Token))
	if err != nil {
		return nil, err
	}

	return &bot{
		session: sess,
		cfg:     cfg,
	}, nil
}

// Run starts the bot and blocks until it is stopped.
func (b *bot) Run(ctx context.Context) error {
	return nil
}

// Shutdown stops the bot.
func (b *bot) Shutdown(ctx context.Context) error {
	return nil
}
