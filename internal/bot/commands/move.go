package commands

import (
	"context"
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/lvlcn-t/loggerhead/logger"
)

var (
	_ Command[*discordgo.InteractionCreate] = (*Move)(nil)
	_ InteractionCommand                    = (*Move)(nil)
)

// Move is a command to move members from one voice channel to another.
// It is an interaction command.
type Move struct {
	// Base is the common base for all commands.
	*Base[*discordgo.InteractionCreate]
	// log is the logger.
	log logger.Logger
}

// NewMove creates a new move command.
func NewMove() *Move {
	name := "move"
	return &Move{
		Base: NewBase[*discordgo.InteractionCreate](name),
		log:  logger.NewNamedLogger(name),
	}
}

// Execute is the handler for the command that is called when the event is triggered.
func (c *Move) Execute(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := context.TODO()

	choices := i.ApplicationCommandData().Options
	if len(choices) != 2 {
		err := c.ReplyToInteraction(ctx, s, i, &discordgo.InteractionResponseData{
			Content: "invalid number of options",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		if err != nil {
			c.log.ErrorContext(ctx, "Error replying to interaction", "error", err)
		}
		return
	}

	from := choices[0].ChannelValue(s)
	to := choices[1].ChannelValue(s)

	if err := c.validateRequest(from, to); err != nil {
		rErr := c.ReplyToInteraction(ctx, s, i, &discordgo.InteractionResponseData{
			Content: err.Error(),
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		if rErr != nil {
			c.log.ErrorContext(ctx, "Error replying to interaction", "error", rErr, "validationError", err)
		}
		return
	}

	for _, m := range from.Members {
		err := s.GuildMemberMove(from.GuildID, m.ID, &to.ID)
		if err != nil {
			c.log.ErrorContext(ctx, "Error moving member", "error", err)
		}
	}

	err := c.ReplyToInteraction(ctx, s, i, &discordgo.InteractionResponseData{
		Content: fmt.Sprintf("Moved %d members from %q to %q", len(from.Members), from.Name, to.Name),
		Flags:   discordgo.MessageFlagsEphemeral,
	})
	if err != nil {
		c.log.ErrorContext(ctx, "Error replying to interaction", "error", err)
	}
}

// Info returns the interaction command information.
func (c *Move) Info() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        c.Name(),
		Description: "Move members from one voice channel to another",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "from",
				Description: "The voice channel to move members from",
				Required:    true,
			},
			{
				Type:        discordgo.ApplicationCommandOptionChannel,
				Name:        "to",
				Description: "The voice channel to move members to",
				Required:    true,
			},
		},
	}
}

// validateRequest validates the move request.
func (c *Move) validateRequest(from, to *discordgo.Channel) error {
	if from == nil || to == nil {
		return errors.New("invalid channel")
	}

	if from.GuildID != to.GuildID {
		return errors.New("channels must be in the same guild")
	}

	if from.Type != discordgo.ChannelTypeGuildVoice || to.Type != discordgo.ChannelTypeGuildStageVoice {
		return errors.New("channels must be voice or stage channels")
	}

	if from.ID == to.ID {
		return errors.New("channels must be different")
	}

	return nil
}
