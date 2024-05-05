package commands

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/lvlcn-t/raid-mate/internal/services"
)

var (
	_ Command[*discordgo.InteractionCreate] = (*Credentials)(nil)
	_ InteractionCommand                    = (*Credentials)(nil)
)

// Credentials is a command to get the login credentials for an account.
// It is an interaction command.
type Credentials struct {
	// Base is the common base for all commands.
	*Base[*discordgo.InteractionCreate]
	// service is the guild service.
	service services.Guild
}

// NewCredentials creates a new credentials command.
func NewCredentials(svc services.Guild) *Credentials {
	return &Credentials{
		Base:    NewBase[*discordgo.InteractionCreate]("credentials"),
		service: svc,
	}
}

// Execute is the handler for the command that is called when the event is triggered.
func (c *Credentials) Execute(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	ctx := context.TODO()

	choices := i.ApplicationCommandData().Options
	if len(choices) != 1 {
		return errors.New("invalid number of options")
	}

	account := choices[0].StringValue()
	if err := c.validateRequest(account); err != nil {
		return err
	}

	credentials, err := c.service.GetCredentials(ctx, i.GuildID, account)
	if err != nil {
		return err
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: fmt.Sprintf("The login credentials for %q are:\n%s\n%s", account, credentials.Username, credentials.Password),
		},
	}, discordgo.WithContext(ctx))
	if err != nil {
		return err
	}

	return nil
}

// Info returns the interaction command information.
func (c *Credentials) Info() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        c.Name(),
		Description: "Get the login credentials for an account",
		NameLocalizations: &map[discordgo.Locale]string{
			discordgo.German: "logindaten",
		},
		DescriptionLocalizations: &map[discordgo.Locale]string{
			discordgo.German: "Erhalte die Login-Daten für einen Account",
		},
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "account",
				Description: "The account to get the login credentials for",
				DescriptionLocalizations: map[discordgo.Locale]string{
					discordgo.German: "Der Account, für den die Login-Daten abgerufen werden sollen",
				},
				Required: true,
				Type:     discordgo.ApplicationCommandOptionString,
				Choices: []*discordgo.ApplicationCommandOptionChoice{
					{
						Name:  "raidbots",
						Value: "raidbots",
					},
				},
			},
		},
	}
}

// validateRequest validates the credentials request.
func (c *Credentials) validateRequest(account string) error {
	if account == "" {
		return errors.New("invalid account")
	}

	if strings.Contains(strings.ToLower(account), "raidbots") {
		return errors.New("unknown account")
	}

	return nil
}
