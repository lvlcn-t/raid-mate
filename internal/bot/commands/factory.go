package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/lvlcn-t/raid-mate/internal/services"
)

// Factory represents a command factory.
type Factory interface {
	// Commands returns all the commands available.
	Commands(services.Collection) []any // TODO: is it possible to type this better?
}

// factory is the command factory.
type factory struct{}

// NewFactory creates a new command factory.
func NewFactory() Factory {
	return &factory{}
}

// registry is the list of command factories.
var registry = []func(services.Collection) any{
	newInteractionCommands,
}

// Commands returns all the commands available.
func (f *factory) Commands(svcs services.Collection) (cmds []any) {
	for _, r := range registry {
		cmds = append(cmds, r(svcs))
	}
	return cmds
}

// InteractionCommand is a command that is triggered by an interaction.
type InteractionCommand interface {
	Command[*discordgo.InteractionCreate]
	// Info returns the interaction command information.
	Info() *discordgo.ApplicationCommand
}

// newInteractionCommands returns all the interaction commands.
func newInteractionCommands(svcs services.Collection) any {
	ic := []InteractionCommand{
		NewCredentials(svcs.Guild),
		NewFeedback(svcs.GitHub),
		NewMove(),
		NewLogs(svcs.Guild),
	}

	ic = append(ic, NewHelp(ic))
	return ic
}