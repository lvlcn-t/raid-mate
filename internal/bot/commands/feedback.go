package commands

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/gofiber/fiber/v3"
	"github.com/lvlcn-t/loggerhead/logger"
	"github.com/lvlcn-t/raid-mate/internal/api"
	"github.com/lvlcn-t/raid-mate/internal/services/feedback"
)

var (
	_ Command[*events.ApplicationCommandInteractionCreate] = (*Feedback)(nil)
	_ ApplicationInteractionCommand                        = (*Feedback)(nil)
)

// Feedback is a command to submit feedback.
type Feedback struct {
	// Base is the common base for all commands.
	*Base[*events.ApplicationCommandInteractionCreate]
	// service is the GitHub service.
	service feedback.Service
}

// newFeedback creates a new feedback command.
func newFeedback(svc feedback.Service) *Feedback {
	name := "feedback"
	return &Feedback{
		Base:    NewBase[*events.ApplicationCommandInteractionCreate](name),
		service: svc,
	}
}

// Handle is the handler for the command that is called when the event is triggered.
func (c *Feedback) Handle(ctx context.Context, event *events.ApplicationCommandInteractionCreate) {
	log := logger.FromContext(ctx).With("command", c.Name())
	data := event.SlashCommandInteractionData()
	fb := data.String("feedback")

	err := c.validateRequest(fb)
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

	guild, ok := event.Guild()
	if !ok {
		guild.Name = "DM"
	}

	err = c.service.Submit(ctx, feedback.Request{
		Feedback: fb,
		Server:   guild.Name,
		Username: event.User().Username,
		UserID:   event.User().ID,
	}, event.Client())
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
		SetContentf("Feedback submitted: %q", fb).
		SetEphemeral(true).
		Build(),
	)
	if err != nil {
		log.ErrorContext(ctx, "Error replying to interaction", "error", err)
	}
}

// HandleHTTP is the handler for the command that is called when the HTTP request is triggered.
func (c *Feedback) HandleHTTP(ctx fiber.Ctx) error {
	log := logger.FromContext(ctx.UserContext()).With("command", c.Name())

	var req feedback.Request
	err := json.Unmarshal(ctx.Body(), &req)
	if err != nil {
		log.DebugContext(ctx.Context(), "Error unmarshalling request", "error", err)
		return api.BadRequestResponse(ctx, "invalid request")
	}

	err = c.validateRequest(req.Feedback)
	if err != nil {
		log.DebugContext(ctx.Context(), "Error validating request", "error", err)
		return api.BadRequestResponse(ctx, "invalid feedback")
	}

	err = c.service.Submit(ctx.Context(), req, nil)
	if err != nil {
		log.ErrorContext(ctx.Context(), "Error submitting feedback", "error", err)
		return api.InternalServerErrorResponse(ctx, "error submitting feedback")
	}

	return ctx.Status(http.StatusOK).JSON(fiber.Map{"status": http.StatusText(http.StatusOK)})
}

// Info returns the interaction command information.
func (c *Feedback) Info() discord.ApplicationCommandCreate {
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
		).Build()
}

// validateRequest validates the feedback request.
func (c *Feedback) validateRequest(fb string) error {
	if fb == "" {
		return errors.New("invalid feedback")
	}

	return nil
}
