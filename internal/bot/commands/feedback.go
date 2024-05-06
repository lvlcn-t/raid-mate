package commands

import (
	"context"
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/lvlcn-t/loggerhead/logger"
	"github.com/lvlcn-t/raid-mate/internal/services"
)

var (
	_ Command[*discordgo.InteractionCreate] = (*Feedback)(nil)
	_ InteractionCommand                    = (*Feedback)(nil)
)

// Feedback is a command to submit feedback.
// It is an interaction command.
type Feedback struct {
	// Base is the common base for all commands.
	*Base[*discordgo.InteractionCreate]
	// service is the GitHub service.
	service services.GitHub
	// log is the logger.
	log logger.Logger
}

// NewFeedback creates a new feedback command.
func NewFeedback(svc services.GitHub) *Feedback {
	name := "feedback"
	return &Feedback{
		Base:    NewBase[*discordgo.InteractionCreate](name),
		service: svc,
		log:     logger.NewNamedLogger(name),
	}
}

// Execute is the handler for the command that is called when the event is triggered.
func (c *Feedback) Execute(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := context.TODO()

	choices := i.ApplicationCommandData().Options
	if len(choices) != 1 {
		err := c.ReplyToInteraction(ctx, s, i, &discordgo.InteractionResponseData{
			Content: "invalid number of options",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		if err != nil {
			c.log.ErrorContext(ctx, "Error replying to interaction", "error", err)
		}
		return
	}

	feedback := choices[0].StringValue()
	if err := c.validateRequest(feedback); err != nil {
		rErr := c.ReplyToInteraction(ctx, s, i, &discordgo.InteractionResponseData{
			Content: err.Error(),
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		if rErr != nil {
			c.log.ErrorContext(ctx, "Error replying to interaction", "error", rErr, "validationError", err)
		}
		return
	}

	err := c.service.CreateIssue(ctx, feedback)
	if err != nil {
		rErr := c.ReplyToInteraction(ctx, s, i, &discordgo.InteractionResponseData{
			Content: "failed to submit feedback",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		if rErr != nil {
			c.log.ErrorContext(ctx, "Error replying to interaction", "error", rErr, "createIssueError", err)
		}
		return
	}

	err = c.ReplyToInteraction(ctx, s, i, &discordgo.InteractionResponseData{
		Content: fmt.Sprintf("Feedback submitted: %q", feedback),
		Flags:   discordgo.MessageFlagsEphemeral,
	})
	if err != nil {
		c.log.ErrorContext(ctx, "Error replying to interaction", "error", err)
	}
}

// Info returns the interaction command information.
func (c *Feedback) Info() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        c.Name(),
		Description: "Submit feedback",
		DescriptionLocalizations: &map[discordgo.Locale]string{
			discordgo.German: "Feedback einreichen",
		},
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "feedback",
				Description: "The feedback to submit",
				DescriptionLocalizations: map[discordgo.Locale]string{
					discordgo.German: "Das Feedback, das eingereicht werden soll",
				},
				Required: true,
			},
		},
	}
}

// validateRequest validates the feedback request.
func (c *Feedback) validateRequest(feedback string) error {
	if feedback == "" {
		return errors.New("invalid feedback")
	}

	return nil
}
