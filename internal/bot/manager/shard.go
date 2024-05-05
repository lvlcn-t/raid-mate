package manager

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
)

type EventHandler struct {
	Handler any
	remover func()
}

// Shard represents a single shard of the bot.
// Shard is safe for simultaneous use by multiple goroutines.
type Shard struct {
	// mu is a mutex to ensure concurrent safety.
	mu sync.RWMutex
	// ID is the shard ID.
	ID int
	// Session is the Discord session.
	Session *discordgo.Session
	// handlers is the list of event handlers.
	handlers []EventHandler
}

// shardInfo is the information required to connect a shard.
type shardInfo struct {
	Token      string
	ID         int
	ShardCount int
	Intent     discordgo.Intent
}

// Connect connects the shard to Discord.
// The shard must not already be connected. Otherwise this will return ErrAlreadyConnected.
func (s *Shard) Connect(ctx context.Context, conn shardInfo) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Session != nil {
		return newErrAlreadyConnected(s.ID)
	}

	if !strings.HasPrefix(conn.Token, "Bot ") {
		conn.Token = fmt.Sprintf("Bot %s", conn.Token)
	}

	s.ID = conn.ID

	sess, err := discordgo.New(conn.Token)
	if err != nil {
		return err
	}

	sess.ShardCount = conn.ShardCount
	sess.ShardID = conn.ID
	sess.Identify.Intents = conn.Intent
	s.Session = sess

	for _, h := range s.handlers {
		s.Session.AddHandler(h.Handler)
	}

	return s.Session.Open()
}

// Disconnect disconnects the shard from Discord.
// The shard must already be connected. Otherwise this will return ErrNotConnected.
func (s *Shard) Disconnect(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Session == nil {
		return newErrNotConnected(s.ID)
	}

	return s.Session.Close()
}

// AddHandler adds an event handler to the shard.
//
// This should only be called before the shard has connected.
// Safe to call concurrently.
func (s *Shard) AddHandlers(h ...EventHandler) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Session != nil {
		for _, h := range h {
			h.remover = s.Session.AddHandler(h.Handler)
		}
	}

	s.handlers = append(s.handlers, h...)
}

// RemoveHandlers removes an event handler from the shard.
//
// This should only be called before the shard has connected.
// Safe to call concurrently.
func (s *Shard) RemoveHandlers() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Session != nil {
		for _, h := range s.handlers {
			h.remover()
		}
	}

	s.handlers = nil
}

// RegisterCommand registers a command for the shard.
//
// This should only be called after the shard has connected.
// Safe to call concurrently.
func (s *Shard) RegisterCommand(guildID string, cmd *discordgo.ApplicationCommand) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.Session == nil {
		return nil
	}

	_, err := s.Session.ApplicationCommandCreate(s.Session.State.User.ID, guildID, cmd)
	return err
}

// RegisterCommandsOverwrite registers multiple commands for the shard, possibly overwriting existing commands.
//
// This should only be called after the shard has connected.
// Safe to call concurrently.
func (s *Shard) RegisterCommandsOverwrite(guildID string, cmds []*discordgo.ApplicationCommand) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.Session == nil {
		return nil
	}

	_, err := s.Session.ApplicationCommandBulkOverwrite(s.Session.State.User.ID, guildID, cmds)
	return err
}

// DeleteCommand deletes a command from the shard.
//
// This should only be called after the shard has connected.
// Safe to call concurrently.
func (s *Shard) DeleteCommand(guildID, cmdID string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.Session == nil {
		return nil
	}

	return s.Session.ApplicationCommandDelete(s.Session.State.User.ID, guildID, cmdID)
}

// GuildCount returns the number of guilds the shard is connected to.
//
// Safe to call concurrently.
func (s *Shard) GuildCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.Session == nil {
		return 0
	}

	s.Session.State.RLock()
	defer s.Session.State.RUnlock()
	return len(s.Session.State.Guilds)
}
