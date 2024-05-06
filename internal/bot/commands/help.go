package commands

import (
	"context"

	"github.com/bwmarrin/discordgo"
	"github.com/lvlcn-t/loggerhead/logger"
)

var (
	_ Command[*discordgo.InteractionCreate] = (*Help)(nil)
	_ InteractionCommand                    = (*Help)(nil)
)

// Help is a command to get help on how to use the bot.
// It is an interaction command.
type Help struct {
	// Base is the common base for all commands.
	*Base[*discordgo.InteractionCreate]
	// Commands is the list of commands to get help for.
	Commands []InteractionCommand
	// log is the logger.
	log logger.Logger
}

// NewHelp creates a new help command.
func NewHelp(cmds []InteractionCommand) *Help {
	name := "help"
	return &Help{
		Base:     NewBase[*discordgo.InteractionCreate](name),
		Commands: cmds,
		log:      logger.NewNamedLogger(name),
	}
}

// Execute is the handler for the command that is called when the event is triggered.
func (c *Help) Execute(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := context.TODO()

	choices := i.ApplicationCommandData().Options
	if len(choices) == 0 {
		c.sendDefaultHelp(ctx, s, i)
		return
	}

	command := choices[0].StringValue()
	cmd := c.lookupCommand(command)
	if cmd == nil {
		c.sendDefaultHelp(ctx, s, i)
		return
	}

	info := c.getCommandInfo(cmd)
	err := c.ReplyToInteraction(ctx, s, i, &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{info},
		Flags:  discordgo.MessageFlagsEphemeral,
	})
	if err != nil {
		c.log.ErrorContext(ctx, "Error replying to interaction", "error", err)
	}
}

// Info returns the interaction command information.
func (c *Help) Info() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        c.Name(),
		Description: "Get help on how to use the bot.",
		DescriptionLocalizations: &map[discordgo.Locale]string{
			discordgo.German: "Erhalte Hilfe, wie du den Bot benutzen kannst.",
		},
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "command",
				Description: "The command to get help for.",
				DescriptionLocalizations: map[discordgo.Locale]string{
					discordgo.German: "Der Befehl, für den du Hilfe benötigst.",
				},
				Required: false,
				Type:     discordgo.ApplicationCommandOptionString,
			},
		},
	}
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
func (c *Help) getCommandInfo(command InteractionCommand) *discordgo.MessageEmbed {
	info := command.Info()

	embed := &discordgo.MessageEmbed{
		Title:       info.Name,
		Description: info.Description,
		Color:       0x00ff00,
		Fields:      []*discordgo.MessageEmbedField{},
	}

	for _, option := range info.Options {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   option.Name,
			Value:  option.Description,
			Inline: true,
		})
	}

	return embed
}

// sendDefaultHelp sends the default help message.
// After calling this you should return from the command handler.
func (c *Help) sendDefaultHelp(ctx context.Context, s *discordgo.Session, i *discordgo.InteractionCreate) {
	embed := &discordgo.MessageEmbed{
		Title:       "Help",
		Description: "Here are the available commands:",
		Color:       0x00ff00,
		Fields:      []*discordgo.MessageEmbedField{},
	}

	for _, command := range c.Commands {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   command.Name(),
			Value:  command.Info().Description,
			Inline: true,
		})
	}

	if err := c.ReplyToInteraction(ctx, s, i, &discordgo.InteractionResponseData{
		Embeds: []*discordgo.MessageEmbed{embed},
		Flags:  discordgo.MessageFlagsEphemeral,
	}); err != nil {
		c.log.ErrorContext(ctx, "Error replying to interaction", "error", err)
	}
}
