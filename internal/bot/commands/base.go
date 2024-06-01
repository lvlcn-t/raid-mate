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

// Base is a common base for all commands.
// It should not be executed but rather embedded in other commands.
// The embedder may only implement command specific methods.
type Base[T Event] struct {
	// name is the name of the command.
	name string
}

// Name returns the name of the command.
func (c *Base[T]) Name() string {
	return c.name
}

// NewBase creates the common base for all commands.
// The name is the name of the command.
// The name should be unique and should not contain spaces.
func NewBase[T Event](name string) *Base[T] {
	if name == "" {
		panic("name cannot be empty")
	}

	return &Base[T]{name: name}
}

func toPtr[T any](v T) *T {
	return &v
}
