package commands

import (
	"context"
	"errors"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/lvlcn-t/loggerhead/logger"
	"github.com/lvlcn-t/raid-mate/internal/services"
)

var (
	_ Command[*events.ApplicationCommandInteractionCreate] = (*Feedback)(nil)
	_ InteractionCommand                                   = (*Feedback)(nil)
)

// Feedback is a command to submit feedback.
type Feedback struct {
	// Base is the common base for all commands.
	*Base[*events.ApplicationCommandInteractionCreate]
	// service is the GitHub service.
	service services.GitHub
}

// newFeedback creates a new feedback command.
func newFeedback(svc services.GitHub) *Feedback {
	name := "feedback"
	return &Feedback{
		Base:    NewBase(name),
		service: svc,
	}
}

// Handle is the handler for the command that is called when the event is triggered.
func (c *Feedback) Handle(ctx context.Context, event *events.ApplicationCommandInteractionCreate) {
	log := logger.FromContext(ctx).With("command", c.Name())
	data := event.SlashCommandInteractionData()
	feedback := data.String("feedback")

	err := c.validateRequest(feedback)
	if err != nil {
		err = event.CreateMessage(discord.NewMessageCreateBuilder().
			SetContent(err.Error()).
			SetEphemeral(true).
			Build(),
		)
		if err != nil {
			log.ErrorContext(ctx, "Error replying to interaction", "error", err)
		}
		return
	}

	err = c.service.CreateIssue(ctx, feedback)
	if err != nil {
		cErr := event.CreateMessage(discord.NewMessageCreateBuilder().
			SetContent("Error while submitting feedback").
			SetEphemeral(true).
			Build(),
		)
		if cErr != nil {
			log.ErrorContext(ctx, "Error replying to interaction", "error", cErr, "createIssueError", err)
		}
		return
	}

	err = event.CreateMessage(discord.NewMessageCreateBuilder().
		SetContentf("Feedback submitted: %q", feedback).
		SetEphemeral(true).
		Build(),
	)
	if err != nil {
		log.ErrorContext(ctx, "Error replying to interaction", "error", err)
	}
}

// Info returns the interaction command information.
func (c *Feedback) Info() InfoBuilder {
	return NewInfoBuilder().
		Name(c.Name(), nil).
		Description("Submit feedback", map[discord.Locale]string{
			discord.LocaleGerman: "Feedback einreichen",
		}).
		Option(NewStringOptionBuilder().
			Name("feedback", nil).
			Description("The feedback to submit", map[discord.Locale]string{
				discord.LocaleGerman: "Das Feedback, das eingereicht werden soll",
			}).
			Required(true).
			Build(),
		)
}

// validateRequest validates the feedback request.
func (c *Feedback) validateRequest(feedback string) error {
	if feedback == "" {
		return errors.New("invalid feedback")
	}

	return nil
}
