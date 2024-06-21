package commands

import (
	"context"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/gofiber/fiber/v3"
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
	// guild is the guild component command.
	guild *Guild
}

// NewCollection creates a new collection of commands.
func NewCollection(svcs services.Collection) Collection {
	c := Collection{
		logs:        newLogs(svcs.Guild),
		credentials: newCredentials(svcs.Guild),
		feedback:    newFeedback(svcs.Feedback),
		profile:     newProfile(svcs.Guild),
		guild:       newGuild(svcs.Guild),
	}
	c.help = newHelp(c.ApplicationInteractionCommands())
	return c
}

// Get returns the command with the given name.
func (c *Collection) GetAppCommand(name string) ApplicationInteractionCommand {
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

func (c *Collection) GetComponentCommand(name string) ComponentInteractionCommand {
	if name == c.guild.Name() {
		return c.guild
	}
	return nil
}

// InteractionCommands returns the interaction commands in the collection.
func (c *Collection) ApplicationInteractionCommands() []ApplicationInteractionCommand {
	ic := []ApplicationInteractionCommand{
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
	infos := make([]discord.ApplicationCommandCreate, len(c.ApplicationInteractionCommands()))
	for i, cmd := range c.ApplicationInteractionCommands() {
		infos[i] = cmd.Info()
	}
	return infos
}

// Router returns a router for the collection.
func (c *Collection) Router() fiber.Router {
	app := fiber.New()
	for _, cmd := range c.ApplicationInteractionCommands() {
		methods, path := cmd.Route()
		if methods == nil {
			app.All(path, cmd.HandleHTTP)
			continue
		}
		app.Add(methods, path, cmd.HandleHTTP)
	}
	return app
}

// ApplicationInteractionCommand is a command that is triggered by an interaction.
type ApplicationInteractionCommand interface {
	Command[*events.ApplicationCommandInteractionCreate]
	// Info returns the interaction command information.
	Info() discord.ApplicationCommandCreate
}

type ComponentInteractionCommand interface {
	Command[*events.ComponentInteractionCreate]
	HandleSubmission(ctx context.Context, event *events.ModalSubmitInteractionCreate)
}
