package commands

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/lvlcn-t/go-kit/apimanager/fiberutils"
	"github.com/lvlcn-t/loggerhead/logger"
	"github.com/lvlcn-t/raid-mate/internal/services/guild"
)

var (
	_ Command[*events.ApplicationCommandInteractionCreate] = (*Logs)(nil)
	_ ApplicationInteractionCommand                        = (*Logs)(nil)
)

// Logs is a command to get the logs for a guild.
type Logs struct {
	// Base is the common base for all commands.
	*Base[*events.ApplicationCommandInteractionCreate]
	// service is the guild service.
	service guild.Service
}

// newLogs creates a new logs command.
func newLogs(svc guild.Service) *Logs {
	return &Logs{
		Base:    NewBase[*events.ApplicationCommandInteractionCreate]("logs"),
		service: svc,
	}
}

// Handle is the handler for the command that is called when the event is triggered.
func (c *Logs) Handle(ctx context.Context, event *events.ApplicationCommandInteractionCreate) {
	log := logger.FromContext(ctx).With("command", c.Name())
	data := event.SlashCommandInteractionData()
	date := data.String("date")
	if date == "" {
		date = time.Now().Format(time.DateOnly)
	}

	d, err := c.parseDate(date)
	if err != nil {
		cErr := event.CreateMessage(discord.NewMessageCreateBuilder().
			SetContent(err.Error()).
			SetEphemeral(true).
			Build(),
		)
		if cErr != nil {
			log.ErrorContext(ctx, "Error replying to interaction", "error", cErr, "parseDateError", err)
		}
		return
	}

	logs, err := c.service.GetReports(ctx, *event.GuildID(), d)
	if err != nil {
		cErr := event.CreateMessage(discord.NewMessageCreateBuilder().
			SetContent("Error while getting logs").
			SetEphemeral(true).
			Build(),
		)
		if cErr != nil {
			log.ErrorContext(ctx, "Error replying to interaction", "error", cErr, "getLogsError", err)
		}
		return
	}

	err = event.CreateMessage(discord.NewMessageCreateBuilder().
		SetContent(strings.Join(logs, "\n")).
		Build(),
	)
	if err != nil {
		log.ErrorContext(ctx, "Error replying to interaction", "error", err)
	}
}

// HandleHTTP is the handler for the command that is called when the HTTP request is triggered.
func (c *Logs) HandleHTTP(ctx fiber.Ctx) error {
	log := logger.FromContext(ctx.UserContext()).With("command", c.Name())
	gid, err := fiberutils.Params(ctx, "guildID", snowflake.Parse)
	if err != nil {
		log.DebugContext(ctx.Context(), "Error parsing guild ID", "error", err)
		return fiberutils.BadRequestResponse(ctx, "missing or invalid guild ID")
	}

	date, err := c.parseDate(ctx.Query("date", time.Now().Format(time.DateOnly)))
	if err != nil {
		log.DebugContext(ctx.Context(), "Error parsing date", "error", err)
		return fiberutils.BadRequestResponse(ctx, "invalid date")
	}

	logs, err := c.service.GetReports(ctx.Context(), gid, date)
	if err != nil {
		return fiberutils.InternalServerErrorResponse(ctx, "Error while getting logs")
	}

	return ctx.Status(http.StatusOK).JSON(fiber.Map{"logs": logs})
}

// Route returns the route for the command.
func (c *Logs) Route() (methods []string, path string) {
	return []string{http.MethodGet}, "/guilds/:guildID/logs"
}

func (c *Logs) Info() discord.ApplicationCommandCreate {
	return NewInfoBuilder().
		Name(c.Name(), nil).
		Description("Fetch guild logs.", map[discord.Locale]string{
			discord.LocaleGerman: "Hole Gilde-Logs.",
		}).
		Option(NewStringOptionBuilder().
			Name("date", map[discord.Locale]string{
				discord.LocaleGerman: "datum",
			}).
			Description("Date of logs (YYYY-MM-DD or YYYY.MM.DD). Defaults to today.", map[discord.Locale]string{
				discord.LocaleGerman: "Datum der Logs (JJJJ-MM-TT oder JJJJ.MM.TT). Standard ist heute.",
			}).
			Required(false).
			Build(),
		).Build()
}

func (c *Logs) parseDate(date string) (time.Time, error) {
	layouts := []string{
		time.DateOnly,
		strings.ReplaceAll(time.DateOnly, "-", "."),
		"02.01.2006",
		"02-01-2006",
	}

	for _, layout := range layouts {
		d, err := time.Parse(layout, date)
		if err == nil {
			return d, nil
		}
	}

	return time.Time{}, errors.New("Invalid date format")
}
