package commands

import (
	"context"
	"errors"
	"strings"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/lvlcn-t/loggerhead/logger"
	"github.com/lvlcn-t/raid-mate/internal/services/guild"
)

var (
	_ Command[*events.ApplicationCommandInteractionCreate] = (*Credentials)(nil)
	_ InteractionCommand                                   = (*Credentials)(nil)
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
		Base:    NewBase("credentials"),
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

	credentials, err := c.service.GetCredentials(ctx, event.GuildID().String(), account)
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

// Info returns the interaction command information.
func (c *Credentials) Info() InfoBuilder {
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
		)
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
