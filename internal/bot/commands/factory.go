package commands

import (
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/lvlcn-t/raid-mate/internal/services"
)

// Collection is a collection of commands.
type Collection struct {
	// logs is the logs command.
	logs *Logs
	// credentials is the credentials command.
	credentials *Credentials
	// feedback is the feedback command.
	feedback *Feedback
	// profile is the profile command.
	profile *Profile
	// help is the help command.
	help *Help
}

// NewCollection creates a new collection of commands.
func NewCollection(svcs services.Collection) Collection {
	c := Collection{
		logs:        newLogs(svcs.Guild),
		credentials: newCredentials(svcs.Guild),
		feedback:    newFeedback(svcs.Feedback),
		profile:     newProfile(svcs.Guild),
	}
	c.help = newHelp(c.InteractionCommands())
	return c
}

// Get returns the command with the given name.
func (c *Collection) Get(name string) InteractionCommand {
	switch name {
	case c.logs.Name():
		return c.logs
	case c.credentials.Name():
		return c.credentials
	case c.feedback.Name():
		return c.feedback
	case c.profile.Name():
		return c.profile
	case c.help.Name():
		return c.help
	default:
		return nil
	}
}

// InteractionCommands returns the interaction commands in the collection.
func (c *Collection) InteractionCommands() []InteractionCommand {
	ic := []InteractionCommand{
		c.logs,
		c.credentials,
		c.feedback,
		c.profile,
	}
	if c.help != nil {
		ic = append(ic, c.help)
	}
	return ic
}

func (c *Collection) Infos() []discord.ApplicationCommandCreate {
	infos := make([]discord.ApplicationCommandCreate, len(c.InteractionCommands()))
	for i, cmd := range c.InteractionCommands() {
		infos[i] = cmd.Info()
	}
	return infos
}

// InteractionCommand is a command that is triggered by an interaction.
type InteractionCommand interface {
	Command[*events.ApplicationCommandInteractionCreate]
	// Info returns the interaction command information.
	Info() discord.ApplicationCommandCreate
}
