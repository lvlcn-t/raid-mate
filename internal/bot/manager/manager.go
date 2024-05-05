package manager

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/lvlcn-t/go-kit/executors"
)

// shutdownTimeout is the timeout for shutting down all shards.
const shutdownTimeout = 20 * time.Second

// Manager is responsible for creating and managing all shards.
// It is safe for simultaneous use by multiple goroutines.
type Manager struct {
	// mu is a mutex to ensure concurrent safety.
	mu sync.RWMutex
	// Session is the Discord session.
	Session *discordgo.Session
	// intent is the intent for all sessions.
	intent discordgo.Intent
	// shards is the list of shards.
	shards []*Shard
	// shardCount is the number of shards.
	shardCount int
	// handlers is the list of event handlers.
	handlers []EventHandler
}

// New creates a new shard manager with the recommended number of shards for the bot.
// The session must be connected before creating a manager.
func New(s *discordgo.Session) (*Manager, error) {
	if s == nil {
		return nil, newErrNilSession()
	}

	mgr := &Manager{
		mu:         sync.RWMutex{},
		Session:    s,
		intent:     discordgo.IntentsNone,
		shards:     make([]*Shard, s.ShardCount),
		shardCount: s.ShardCount,
		handlers:   []EventHandler{},
	}

	gwInfo, err := mgr.Session.GatewayBot()
	if err != nil {
		return nil, err
	}

	if gwInfo.Shards < 1 {
		mgr.shardCount = 1
		return mgr, nil
	}

	mgr.shardCount = gwInfo.Shards
	return mgr, nil
}

// Start starts all shards. If an error occurs, the error will be returned.
// If the manager is already started, the error will be wrapped in an ErrAlreadyConnected.
func (m *Manager) Start(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.shardCount < 1 {
		m.shardCount = 1
	}

	m.shards = make([]*Shard, m.shardCount)
	for i := 0; i < m.shardCount; i++ {
		m.shards[i] = &Shard{}
	}

	for id, shard := range m.shards {
		connect := executors.Retry(func(ctx context.Context) error {
			for _, h := range m.handlers {
				shard.AddHandlers(h)
			}

			err := shard.Connect(ctx, shardInfo{
				Token:      m.Session.Identify.Token,
				ID:         id,
				ShardCount: m.shardCount,
				Intent:     m.intent,
			})
			if err != nil {
				return err
			}

			return nil
		})

		if id != len(m.shards)-1 {
			connect = connect.WithRateLimit(5)
		}

		if err := connect.Do(ctx); err != nil {
			return err
		}
	}

	return nil
}

// Shutdown disconnects all shards. If an error occurs, the error will be wrapped in an ErrShutdown.
func (m *Manager) Shutdown(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, shard := range m.shards {
		disconnect := executors.Retry(func(ctx context.Context) error {
			if err := shard.Disconnect(ctx); err != nil {
				if !errors.Is(err, &ErrNotConnected{}) {
					return err
				}
			}
			return nil
		}).WithTimeout(shutdownTimeout / time.Duration(m.shardCount))

		if err := disconnect.Do(ctx); err != nil {
			return &ErrShutdown{Err: err}
		}
	}

	return nil
}

// Restart restarts the manager with the same shard information.
// If an error occurs, either the old manager or the new manager will be returned.
// If the new manager is returned but the old manager fails to shut down,
// the error will be wrapped in an ErrShutdown.
func (m *Manager) Restart(ctx context.Context) (*Manager, error) {
	m.mu.RLock()

	mgr, err := New(m.Session)
	if err != nil {
		m.mu.RUnlock()
		return m, err
	}

	for _, h := range m.handlers {
		mgr.AddHandlers(h)
	}

	mgr.SetIntent(m.intent)

	m.mu.RUnlock()

	if err := mgr.Start(ctx); err != nil {
		return m, err
	}

	if err := m.Shutdown(ctx); err != nil {
		return mgr, err
	}

	return mgr, nil
}

// SetIntent sets the intent for all sessions.
func (m *Manager) SetIntent(intent discordgo.Intent) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.intent = intent
}

// AddHandlers adds event handlers to all shards.
// Safe to call concurrently.
func (m *Manager) AddHandlers(h ...EventHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.handlers = append(m.handlers, h...)

	for _, shard := range m.shards {
		shard.AddHandlers(h...)
	}
}

// RemoveHandlers removes all event handlers from all shards.
// Safe to call concurrently.
func (m *Manager) RemoveHandlers() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.handlers = nil

	for _, shard := range m.shards {
		shard.RemoveHandlers()
	}
}

// RegisterCommand registers a command for all shards.
// Safe to call concurrently.
func (m *Manager) RegisterCommand(guildID string, cmd *discordgo.ApplicationCommand) (err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, shard := range m.shards {
		if rErr := shard.RegisterCommand(guildID, cmd); err != nil {
			err = errors.Join(err, rErr)
		}
	}

	return
}

// RegisterCommandsOverwrite registers multiple commands for all shards, possibly overwriting existing commands.
// Safe to call concurrently.
func (m *Manager) RegisterCommandsOverwrite(guildID string, cmds []*discordgo.ApplicationCommand) (err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, shard := range m.shards {
		if rErr := shard.RegisterCommandsOverwrite(guildID, cmds); err != nil {
			err = errors.Join(err, rErr)
		}
	}

	return
}

// DeleteCommand deletes a command for all shards.
// Safe to call concurrently.
func (m *Manager) DeleteCommand(guildID, cmdID string) (err error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, shard := range m.shards {
		if rErr := shard.DeleteCommand(guildID, cmdID); err != nil {
			err = errors.Join(err, rErr)
		}
	}

	return
}

// ListGuilds returns a list of all guilds the manager's shards are connected to.
// Returns a copy of the list, so it is safe to modify the list.
func (m *Manager) ListGuilds() []*discordgo.Guild {
	m.Session.RLock()
	defer m.Session.RUnlock()
	guilds := make([]*discordgo.Guild, 0, len(m.Session.State.Guilds))
	_ = copy(guilds, m.Session.State.Guilds)
	return guilds
}

// GuildCount returns the number of guilds the manager's shards are connected to.
// Safe to call concurrently.
func (m *Manager) GuildCount() (count int) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, shard := range m.shards {
		count += shard.GuildCount()
	}

	return
}

// ShardCount returns the number of shards the manager has.
// Safe to call concurrently.
func (m *Manager) ShardCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.shardCount
}

// Shard returns the shard with the specified ID.
// Safe to call concurrently.
func (m *Manager) Shard(id int) *Shard {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if id < 0 || id >= m.shardCount {
		return nil
	}

	return m.shards[id]
}

// SetShardCount sets the number of shards the manager has.
// This won't take effect until the manager is restarted.
// Safe to call concurrently.
func (m *Manager) SetShardCount(count int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if count < 1 {
		count = 1
	}

	m.shardCount = count
}

// SetShard sets the shard with the specified ID.
// Safe to call concurrently.
func (m *Manager) SetShard(id int, shard *Shard) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if id < 0 || id >= m.shardCount {
		return
	}

	m.shards[id] = shard
}

// SessionForDM returns the session for DMs.
// Safe to call concurrently.
func (m *Manager) SessionForDM() *discordgo.Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// See https://discord.com/developers/docs/topics/gateway#sharding
	return m.shards[0].Session
}

// SessionForGuild returns the session for the specified guild.
// Safe to call concurrently.
func (m *Manager) SessionForGuild(guildID int64) *discordgo.Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// See https://discord.com/developers/docs/topics/gateway#sharding
	return m.shards[int(guildID>>22)%m.shardCount].Session
}
