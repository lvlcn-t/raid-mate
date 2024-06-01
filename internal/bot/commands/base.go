package commands

import (
	"context"

	"github.com/disgoorg/disgo/events"
)

// Event is an constaint interface for all Discord events.
// This should only be used for type constraints.
type Event interface {
	*events.ApplicationCommandInteractionCreate
}

// Command is an interface for a command.
type Command[T Event] interface {
	// Name returns the name of the command.
	Name() string
	// Handle is the handler for the command that is called when the event is triggered.
	Handle(ctx context.Context, event T)
}

type eventHandler[T Event] func(ctx context.Context, event T)

// Base is a common base for all commands.
// It should not be executed but rather embedded in other commands.
// The embedder may only implement command specific methods.
//
// Example:

type Base[T Event] struct {
	// name is the name of the command.
	name string
	// handleEvent is the handler for the command that is called when the event is triggered.
	handleEvent eventHandler[T]
}

// Name returns the name of the command.
func (c *Base[T]) Name() string {
	return c.name
}

// Execute is the handler for the command that is called when the event is triggered.
func (c *Base[T]) Handle(ctx context.Context, event T) {
	c.handleEvent(ctx, event)
}

// NewBase creates the common base for all commands.
// The name is the name of the command.
// The name should be unique and should not contain spaces.
func NewBase[T Event](name string, handler eventHandler[T]) *Base[T] {
	if name == "" {
		panic("name cannot be empty")
	}

	if handler == nil {
		panic("handler cannot be nil")
	}

	return &Base[T]{
		name:        name,
		handleEvent: handler,
	}
}

func toPtr[T any](v T) *T {
	return &v
}
