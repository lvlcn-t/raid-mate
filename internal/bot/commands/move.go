package commands

import (
	"context"
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
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
}

// NewMove creates a new move command.
func NewMove() *Move {
	return &Move{
		Base: NewBase[*discordgo.InteractionCreate]("move"),
	}
}

// Execute is the handler for the command that is called when the event is triggered.
func (c *Move) Execute(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	choices := i.ApplicationCommandData().Options
	if len(choices) != 2 {
		return errors.New("invalid number of options")
	}

	from := choices[0].ChannelValue(s)
	to := choices[1].ChannelValue(s)

	if err := c.validateRequest(from, to); err != nil {
		return err
	}

	for _, m := range from.Members {
		err := s.GuildMemberMove(from.GuildID, m.ID, &to.ID)
		if err != nil {
			return err
		}
	}

	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("Moved %d members from %q to %q", len(from.Members), from.Name, to.Name),
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	}, discordgo.WithContext(context.TODO()))
	if err != nil {
		return err
	}

	return nil
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
