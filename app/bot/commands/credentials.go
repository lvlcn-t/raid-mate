package commands

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/lvlcn-t/go-kit/apimanager/fiberutils"
	"github.com/lvlcn-t/loggerhead/logger"
	"github.com/lvlcn-t/raid-mate/app/database/repo"
	"github.com/lvlcn-t/raid-mate/app/services/guild"
)

var (
	_ Command[*events.ApplicationCommandInteractionCreate] = (*Credentials)(nil)
	_ ApplicationInteractionCommand                        = (*Credentials)(nil)
)

// Credentials is a command to get the login credentials for an account.
type Credentials struct {
	// Base is the common base for all commands.
	*Base[*events.ApplicationCommandInteractionCreate]
	// service is the guild service.
	service guild.Service
}

// newCredentials creates a new credentials command.
func newCredentials(svc guild.Service) *Credentials {
	return &Credentials{
		Base:    NewBase[*events.ApplicationCommandInteractionCreate]("credentials"),
		service: svc,
	}
}

// Handle is the handler for the command that is called when the event is triggered.
func (c *Credentials) Handle(ctx context.Context, event *events.ApplicationCommandInteractionCreate) {
	log := logger.FromContext(ctx).With("command", c.Name())
	data := event.SlashCommandInteractionData()
	account := data.String("account")

	err := c.validateRequest(account)
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

	err = event.DeferCreateMessage(true)
	if err != nil {
		log.ErrorContext(ctx, "Error deferring interaction", "error", err)
		return
	}

	credentials, err := c.service.GetCredentials(ctx, repo.GetCredentialsParams{
		GuildID: int64(*event.GuildID()), //nolint:gosec // Snowflake cannot overflow AFAIK
		Name:    account,
	})
	if err != nil {
		cErr := event.CreateMessage(discord.NewMessageCreateBuilder().
			SetContent("Error while getting credentials").
			SetEphemeral(true).
			Build(),
		)
		if cErr != nil {
			log.ErrorContext(ctx, "Error replying to interaction", "error", cErr, "getCredentialsError", err)
		}
		return
	}

	err = event.CreateMessage(discord.NewMessageCreateBuilder().
		SetContentf("The login credentials for %q are:\nUsername: %s\nPassword: %s", account, credentials.Username, credentials.Password).
		SetEphemeral(true).
		Build(),
	)
	if err != nil {
		log.ErrorContext(ctx, "Error replying to interaction", "error", err)
	}
}

// HandleHTTP is the handler for the command that is called when the HTTP request is triggered.
func (c *Credentials) HandleHTTP(ctx fiber.Ctx) error {
	log := logger.FromContext(ctx.Context()).With("command", c.Name())
	type request struct {
		Account string `json:"account"`
	}

	req, err := fiberutils.Body[request](ctx)
	if err != nil {
		log.DebugContext(ctx.Context(), "Error unmarshalling request", "error", err)
		return fiberutils.BadRequestResponse(ctx, "malformed request")
	}

	gid, err := fiberutils.Params(ctx, "guildID", snowflake.Parse)
	if err != nil {
		log.DebugContext(ctx.Context(), "Error parsing guild ID", "error", err)
		return fiberutils.BadRequestResponse(ctx, "missing or invalid guild ID")
	}

	err = c.validateRequest(req.Account)
	if err != nil {
		log.DebugContext(ctx.Context(), "Error validating request", "error", err)
		return fiberutils.BadRequestResponse(ctx, err.Error())
	}

	credentials, err := c.service.GetCredentials(ctx.Context(), repo.GetCredentialsParams{
		GuildID: int64(gid), //nolint:gosec // Snowflake cannot overflow AFAIK
		Name:    req.Account,
	})
	if err != nil {
		log.ErrorContext(ctx.Context(), "Error getting credentials", "error", err)
		return fiberutils.InternalServerErrorResponse(ctx, "error getting credentials")
	}

	return ctx.Status(http.StatusOK).JSON(fiber.Map{
		"username": credentials.Username,
		"password": credentials.Password,
	})
}

// Route returns the route for the command.
func (c *Credentials) Route() (methods []string, route string) {
	return []string{http.MethodPost}, "/guilds/:guildID/credentials"
}

// Info returns the interaction command information.
func (c *Credentials) Info() discord.ApplicationCommandCreate {
	return NewInfoBuilder().
		Name(c.Name(), map[discord.Locale]string{
			discord.LocaleGerman: "logindaten",
		}).
		Description("Get the login credentials for an account", map[discord.Locale]string{
			discord.LocaleGerman: "Erhalte die Login-Daten für einen Account",
		}).
		Option(NewStringOptionBuilder().
			Name("account", nil).
			Description("The account to get the login credentials for", map[discord.Locale]string{
				discord.LocaleGerman: "Der Account, für den die Login-Daten abgerufen werden sollen",
			}).
			Required(true).
			Choices(NewStringOptionChoice("raidbots", "raidbots", nil)).
			Build(),
		).Build()
}

// validateRequest validates the credentials request.
func (c *Credentials) validateRequest(account string) error {
	if account == "" {
		return errors.New("invalid account")
	}

	if !strings.EqualFold(account, "raidbots") {
		return errors.New("unknown account")
	}

	return nil
}
