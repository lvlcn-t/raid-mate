package commands

import (
	"context"
	"fmt"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/lvlcn-t/loggerhead/logger"
	"github.com/lvlcn-t/raid-mate/app/bot/colors"
)

var (
	_ Command[*events.ApplicationCommandInteractionCreate] = (*Help)(nil)
	_ ApplicationInteractionCommand                        = (*Help)(nil)
)

// Help is a command to get help on how to use the bot.
type Help struct {
	// Base is the common base for all commands.
	*Base[*events.ApplicationCommandInteractionCreate]
	// commands is the list of commands to get help for.
	commands []ApplicationInteractionCommand
}

// newHelp creates a new help command.
func newHelp(cmds []ApplicationInteractionCommand) *Help {
	cmd := &Help{
		Base:     NewBase[*events.ApplicationCommandInteractionCreate]("help"),
		commands: cmds,
	}
	cmd.commands = append(cmd.commands, cmd)
	return cmd
}

// Handle is the handler for the command that is called when the event is triggered.
func (c *Help) Handle(ctx context.Context, event *events.ApplicationCommandInteractionCreate) {
	log := logger.FromContext(ctx).With("command", c.Name())
	data := event.SlashCommandInteractionData()
	command := data.String("name")
	if command == "" {
		c.sendDefaultHelp(ctx, event)
		return
	}

	cmd := c.lookup(command)
	if cmd == nil {
		c.sendDefaultHelp(ctx, event)
		return
	}

	info := c.getInfo(cmd)
	err := event.CreateMessage(discord.NewMessageCreateBuilder().
		AddEmbeds(info).
		SetEphemeral(true).
		Build(),
	)
	if err != nil {
		log.ErrorContext(ctx, "Error replying to interaction", "error", err)
	}
}

func (c *Help) Info() discord.ApplicationCommandCreate {
	var choices []discord.ApplicationCommandOptionChoiceString
	for _, command := range c.commands {
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
		).Build()
}

// lookup finds the interaction command with the given name.
func (c *Help) lookup(name string) ApplicationInteractionCommand {
	for _, command := range c.commands {
		if command.Name() == name {
			return command
		}
	}
	return nil
}

// getInfo returns the information for the given command.
func (c *Help) getInfo(command ApplicationInteractionCommand) discord.Embed {
	info := command.Info().(discord.SlashCommandCreate)
	return discord.NewEmbedBuilder().
		SetTitle(command.Name()).
		SetDescription(info.Description).
		SetColor(colors.Red.Int()).
		Build()
}

// sendDefaultHelp sends the default help message.
// After calling this you should return from the command handler.
func (c *Help) sendDefaultHelp(ctx context.Context, event *events.ApplicationCommandInteractionCreate) {
	log := logger.FromContext(ctx).With("command", c.Name())
	var fields []discord.EmbedField
	for _, cmd := range c.commands {
		fields = append(fields, discord.EmbedField{
			Name:   fmt.Sprintf("Command: `/%s`", cmd.Name()),
			Value:  cmd.Info().(discord.SlashCommandCreate).Description,
			Inline: toPtr(false),
		})
	}

	embed := discord.NewEmbedBuilder().
		SetTitle("Help").
		SetDescription("Here are the available commands:").
		SetColor(colors.Red.Int()).
		AddFields(fields...).
		Build()

	err := event.CreateMessage(discord.NewMessageCreateBuilder().
		AddEmbeds(embed).
		SetEphemeral(true).
		Build(),
	)
	if err != nil {
		log.ErrorContext(ctx, "Error replying to interaction", "error", err)
	}
}
