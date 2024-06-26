package commands

import (
	"context"
	"fmt"
	"net/http"

	"github.com/disgoorg/disgo/events"
	"github.com/gofiber/fiber/v3"
)

// Event is an constaint interface for all Discord events.
// This should only be used for type constraints.
type Event interface {
	*events.ApplicationCommandInteractionCreate | *events.ComponentInteractionCreate
}

// Command is an interface for a command.
type Command[T Event] interface {
	// Name returns the name of the command.
	Name() string
	// Handle is the handler for the command that is called when the event is triggered.
	Handle(ctx context.Context, event T)
	// HandleHTTP is the handler for the command that is called when the HTTP request is triggered.
	HandleHTTP(ctx fiber.Ctx) error
	// Route returns the route for the command.
	Route() (methods []string, path string)
}

// Base is a common base for all commands.
// It should not be executed but rather embedded in other commands.
// The embedder may only implement command specific methods.
//
// Example:
//
//	type MyCommand struct {
//	    *Base[*events.ApplicationCommandInteractionCreate]
//	}
//
//	func NewMyCommand() *MyCommand {
//	    return &MyCommand{Base: NewBase("my-command")}
//	}
//
//	func (c *MyCommand) Handle(ctx context.Context, event *events.ApplicationCommandInteractionCreate) {
//	    // handle the command
//	}
type Base[T Event] struct {
	// name is the name of the command.
	name string
}

// Name returns the name of the command.
func (c *Base[T]) Name() string {
	return c.name
}

// Handle is the handler for the command that is called when the event is triggered.
// This is a default implementation that panics if not overridden.
func (c *Base[T]) Handle(_ context.Context, _ T) {
	panic(fmt.Sprintf("command %q does not have a handler", c.Name()))
}

// HandleHTTP is the handler for the command that is called when the HTTP request is triggered.
// This is a default implementation that returns a 404 status.
func (c *Base[T]) HandleHTTP(ctx fiber.Ctx) error {
	return ctx.SendStatus(http.StatusNotFound)
}

// Route returns the route for the command.
func (c *Base[T]) Route() (methods []string, path string) {
	return nil, fmt.Sprintf("/%s", c.Name())
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

// toPtr returns a pointer to the given value.
func toPtr[T any](v T) *T {
	return &v
}
