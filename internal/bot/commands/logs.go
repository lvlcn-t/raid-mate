package commands

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/lvlcn-t/loggerhead/logger"
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
	// log is the logger.
	log logger.Logger
}

// NewLogs creates a new logs command.
func NewLogs(svc services.Guild) *Logs {
	name := "logs"
	return &Logs{
		Base:    NewBase[*discordgo.InteractionCreate](name),
		service: svc,
		log:     logger.NewNamedLogger(name),
	}
}

// Execute is the handler for the command that is called when the event is triggered.
func (c *Logs) Execute(s *discordgo.Session, i *discordgo.InteractionCreate) {
	ctx := context.TODO()

	choices := i.ApplicationCommandData().Options
	if len(choices) > 1 {
		err := c.ReplyToInteraction(ctx, s, i, &discordgo.InteractionResponseData{
			Content: "invalid number of options",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		if err != nil {
			c.log.ErrorContext(ctx, "Error replying to interaction", "error", err)
		}
		return
	}

	date := time.Now().Format(time.DateOnly)
	if len(choices) == 1 {
		date = choices[0].StringValue()
	}

	d, err := c.parseDate(date)
	if err != nil {
		rErr := c.ReplyToInteraction(ctx, s, i, &discordgo.InteractionResponseData{
			Content: err.Error(),
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		if rErr != nil {
			c.log.ErrorContext(ctx, "Error replying to interaction", "error", rErr, "parseDateError", err)
		}
		return
	}

	logs, err := c.service.GetLogs(ctx, i.GuildID, d)
	if err != nil {
		rErr := c.ReplyToInteraction(ctx, s, i, &discordgo.InteractionResponseData{
			Content: "error getting logs",
			Flags:   discordgo.MessageFlagsEphemeral,
		})
		if rErr != nil {
			c.log.ErrorContext(ctx, "Error replying to interaction", "error", rErr, "getLogsError", err)
		}
		return
	}

	err = c.ReplyToInteraction(ctx, s, i, &discordgo.InteractionResponseData{
		Content: strings.Join(logs, "\n"),
	})
	if err != nil {
		c.log.ErrorContext(ctx, "Error replying to interaction", "error", err)
	}
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
				NameLocalizations: map[discordgo.Locale]string{
					discordgo.German: "datum",
				},
				DescriptionLocalizations: map[discordgo.Locale]string{
					discordgo.German: "Das Datum der zu erhaltenden Logs (JJJJ-MM-TT oder JJJJ.MM.TT). Wenn nicht angegeben, wird das aktuelle Datum verwendet.",
				},
				Type:     discordgo.ApplicationCommandOptionString,
				Required: false,
			},
		},
	}
}

func (c *Logs) parseDate(date string) (time.Time, error) {
	d, err := time.Parse(time.DateOnly, date)
	if err == nil {
		return d, nil
	}

	d, err = time.Parse(strings.ReplaceAll(time.DateOnly, "-", "."), date)
	if err == nil {
		return d, nil
	}

	return time.Time{}, errors.New("invalid date format")
}
