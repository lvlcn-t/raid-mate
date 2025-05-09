package commands

import (
	"context"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/disgo/rest"
	"github.com/lvlcn-t/loggerhead/logger"
	"github.com/lvlcn-t/raid-mate/app/database/repo"
	"github.com/lvlcn-t/raid-mate/app/services/guild"
)

type Guild struct {
	// Base is the common base for all commands.
	*Base[*events.ComponentInteractionCreate]
	// service is the guild service.
	service guild.Service
}

// newGuild creates a new guild command.
func newGuild(svc guild.Service) *Guild {
	return &Guild{
		Base:    NewBase[*events.ComponentInteractionCreate]("guild"),
		service: svc,
	}
}

const (
	// minRealmLength is the minimum length of the realm name.
	minRealmLength = 2
	// maxRealmLength is the maximum length of the realm name.
	maxRealmLength = 30
	// minRegionLength is the minimum length of the region name.
	minRegionLength = 2
	// maxRegionLength is the maximum length of the region name.
	maxRegionLength = 2
	// minFactionLength is the minimum length of the faction name.
	minFactionLength = 5
	// maxFactionLength is the maximum length of the faction name.
	maxFactionLength = 8
)

// Handle is the handler for the command that is called when the event is triggered.
func (c *Guild) Handle(ctx context.Context, event *events.ComponentInteractionCreate) {
	log := logger.FromContext(ctx).With("command", c.Name())

	err := event.Modal(discord.NewModalCreateBuilder().SetTitle("Setup your Guild").
		AddContainerComponents(
			discord.NewActionRow(
				discord.NewTextInput("guild_name", discord.TextInputStyleShort, "Name of the Guild").
					WithRequired(true).WithPlaceholder("My Guild"),
			),
			discord.NewActionRow(
				discord.NewTextInput("guild_realm", discord.TextInputStyleShort, "Realm of the Guild").
					WithRequired(true).WithPlaceholder("Draenor").WithMinLength(minRealmLength).WithMaxLength(maxRealmLength),
			),
			discord.NewActionRow(
				discord.NewTextInput("guild_region", discord.TextInputStyleShort, "Region of the Guild (EU, US, etc.)").
					WithRequired(true).WithPlaceholder("EU").WithMinLength(minRegionLength).WithMaxLength(maxRegionLength),
			),
			discord.NewActionRow(
				discord.NewTextInput("guild_faction", discord.TextInputStyleShort, "Faction of the Guild (Alliance / Horde)").
					WithRequired(true).WithPlaceholder("Horde").WithMinLength(minFactionLength).WithMaxLength(maxFactionLength),
			),
		).
		SetCustomID("guild").
		Build(), rest.WithCtx(ctx))
	if err != nil {
		log.ErrorContext(ctx, "Error replying to interaction", "error", err)
	}
}

func (c *Guild) HandleSubmission(ctx context.Context, event *events.ModalSubmitInteractionCreate) {
	log := logger.FromContext(ctx).With("command", c.Name())
	name := event.Data.Text("guild_name")
	realm := event.Data.Text("guild_realm")
	region := event.Data.Text("guild_region")
	_ = event.Data.Text("guild_faction")

	err := c.service.Create(ctx, repo.NewGuildParams{
		ID:           int64(*event.GuildID()), //nolint:gosec // Snowflake cannot overflow AFAIK
		Name:         name,
		ServerName:   realm,
		ServerRegion: region,
	})
	if err != nil {
		log.ErrorContext(ctx, "Error creating guild", "error", err)
		err = event.CreateMessage(discord.NewMessageCreateBuilder().
			SetContent("Error while creating guild").
			SetEphemeral(true).
			Build(),
		)
		if err != nil {
			log.ErrorContext(ctx, "Error replying to interaction", "error", err)
		}
		return
	}

	err = event.CreateMessage(discord.NewMessageCreateBuilder().
		SetContent("Guild created").
		SetEphemeral(true).
		Build(),
	)
	if err != nil {
		log.ErrorContext(ctx, "Error replying to interaction", "error", err)
	}
}
