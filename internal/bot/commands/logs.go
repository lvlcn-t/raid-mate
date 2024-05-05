package commands

import (
	"github.com/bwmarrin/discordgo"
	"github.com/lvlcn-t/raid-mate/internal/services"
)

var (
	_ Command[*discordgo.InteractionCreate] = (*Logs)(nil)
	_ InteractionCommand                    = (*Logs)(nil)
)

// Logs is a command to get the logs for a guild.
// It is an interaction command.
type Logs struct {
	// Base is the common base for all commands.
	*Base[*discordgo.InteractionCreate]
	// service is the guild service.
	service services.Guild
}

// NewLogs creates a new logs command.
func NewLogs(svc services.Guild) *Logs {
	return &Logs{
		Base:    NewBase[*discordgo.InteractionCreate]("logs"),
		service: svc,
	}
}

// Execute is the handler for the command that is called when the event is triggered.
func (c *Logs) Execute(_ *discordgo.Session, _ *discordgo.InteractionCreate) error {
	// TODO: implement the logs command
	err := c.validateRequest()
	if err != nil {
		return err
	}
	return nil
}

func (c *Logs) Info() *discordgo.ApplicationCommand {
	return &discordgo.ApplicationCommand{
		Name:        c.Name(),
		Description: "Get the logs for the guild.",
		DescriptionLocalizations: &map[discordgo.Locale]string{
			discordgo.German: "Erhalte die Logs für die Gilde.",
		},
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "date",
				Description: "The date of the logs to get (YYYY-MM-DD or YYYY.MM.DD). If not provided, the current date is used.",
				DescriptionLocalizations: map[discordgo.Locale]string{
					discordgo.German: "Das Datum der zu erhaltenden Logs (JJJJ-MM-TT oder JJJJ.MM.TT). Wenn nicht angegeben, wird das aktuelle Datum verwendet.",
				},
				Type:     discordgo.ApplicationCommandOptionString,
				Required: false,
			},
		},
	}
}

// validateRequest validates the request.
func (c *Logs) validateRequest() error {
	// TODO: implement the request validation
	return nil
}
