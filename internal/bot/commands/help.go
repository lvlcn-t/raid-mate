package commands

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/lvlcn-t/loggerhead/logger"
)

// green is the color green.
const green = 0xf44336

var (
	_ Command[*events.ApplicationCommandInteractionCreate] = (*Help)(nil)
	_ InteractionCommand                                   = (*Help)(nil)
)

// Help is a command to get help on how to use the bot.
// It is an interaction command.
type Help struct {
	// Base is the common base for all commands.
	*Base[*events.ApplicationCommandInteractionCreate]
	// Commands is the list of commands to get help for.
	Commands []InteractionCommand
	// log is the logger.
	log logger.Logger
}

// newHelp creates a new help command.
func newHelp(cmds []InteractionCommand) *Help {
	name := "help"
	cmd := &Help{
		Commands: cmds,
		log:      logger.NewNamedLogger(name),
	}
	cmd.Base = NewBase(name, cmd.handle)
	cmd.Commands = append(cmd.Commands, cmd)
	return cmd
}

// Execute is the handler for the command that is called when the event is triggered.
func (c *Help) handle(ctx context.Context, event *events.ApplicationCommandInteractionCreate) {
	data := event.SlashCommandInteractionData()
	command := data.String("name")
	if command == "" {
		c.sendDefaultHelp(ctx, event)
		return
	}

	cmd := c.lookupCommand(command)
	if cmd == nil {
		c.sendDefaultHelp(ctx, event)
		return
	}

	info := c.getCommandInfo(cmd)
	err := event.CreateMessage(discord.NewMessageCreateBuilder().
		AddEmbeds(info).
		SetEphemeral(true).
		Build(),
	)
	if err != nil {
		c.log.ErrorContext(ctx, "Error replying to interaction", "error", err)
	}
}

func (c *Help) Info() InfoBuilder {
	var choices []discord.ApplicationCommandOptionChoiceString
	for _, command := range c.Commands {
		choices = append(choices, NewStringOptionChoice(command.Name(), command.Name(), nil))
	}

	return NewInfoBuilder().
		Name(c.Name(), nil).
		Description("Get help on how to use the bot.", map[discord.Locale]string{
			discord.LocaleGerman: "Erhalte Hilfe, wie du den Bot benutzen kannst.",
		}).
		Option(NewStringOptionBuilder().
			Name("name", nil).
			Description("The name of the command to get help for.", map[discord.Locale]string{
				discord.LocaleGerman: "Der Name des Befehls, für den du Hilfe benötigst.",
			}).
			Required(false).
			Choices(choices...).
			Build(),
		)
}

// lookupCommand finds the interaction command with the given name.
func (c *Help) lookupCommand(name string) InteractionCommand {
	for _, command := range c.Commands {
		if command.Name() == name {
			return command
		}
	}
	return nil
}

// getCommandInfo returns the information for the given command.
func (c *Help) getCommandInfo(command InteractionCommand) discord.Embed {
	info := command.Info().Build()

	return discord.NewEmbedBuilder().
		SetTitle(command.Name()).
		SetDescription(info.Description).
		SetColor(green).
		Build()
}

// sendDefaultHelp sends the default help message.
// After calling this you should return from the command handler.
func (c *Help) sendDefaultHelp(ctx context.Context, event *events.ApplicationCommandInteractionCreate) {
	var fields []discord.EmbedField
	for _, cmd := range c.Commands {
		fields = append(fields, discord.EmbedField{
			Name:   fmt.Sprintf("Command: `/%s`", cmd.Name()),
			Value:  cmd.Info().Build().Description,
			Inline: toPtr(false),
		})
	}

	embed := discord.NewEmbedBuilder().
		SetTitle("Help").
		SetDescription("Here are the available commands:").
		SetColor(green).
		AddFields(fields...).
		Build()

	err := event.CreateMessage(discord.NewMessageCreateBuilder().
		AddEmbeds(embed).
		SetEphemeral(true).
		Build(),
	)
	if err != nil {
		c.log.ErrorContext(ctx, "Error replying to interaction", "error", err)
	}
}
