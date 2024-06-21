package commands

import (
	"context"
	"fmt"
	"net/http"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/disgoorg/snowflake/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/lvlcn-t/loggerhead/logger"
	"github.com/lvlcn-t/raid-mate/internal/api"
	"github.com/lvlcn-t/raid-mate/internal/services/guild"
)

var (
	_ Command[*events.ApplicationCommandInteractionCreate] = (*Profile)(nil)
	_ ApplicationInteractionCommand                        = (*Profile)(nil)
)

// Profile is a command to get profiles.
type Profile struct {
	// Base is the common base for all commands.
	*Base[*events.ApplicationCommandInteractionCreate]
	// service is the guild service.
	service guild.Service
}

// newProfile creates a new profile command.
func newProfile(svc guild.Service) *Profile {
	return &Profile{
		Base:    NewBase[*events.ApplicationCommandInteractionCreate]("profile"),
		service: svc,
	}
}

// Handle is the handler for the command that is called when the event is triggered.
func (c *Profile) Handle(ctx context.Context, event *events.ApplicationCommandInteractionCreate) {
	log := logger.FromContext(ctx).With("command", c.Name())
	data := event.SlashCommandInteractionData()
	typ := data.String("name")
	username := data.String("username")

	member := event.Member()
	if member == nil {
		log.ErrorContext(ctx, "No member found in interaction")
		return
	}

	profile, err := c.service.GetProfile(ctx, &guild.RequestProfile{
		Type:    typ,
		GuildID: *event.GuildID(),
		User:    username,
	})
	if err != nil {
		log.ErrorContext(ctx, "Error getting profile", "error", err)
		err = event.CreateMessage(discord.NewMessageCreateBuilder().
			SetContent("Error while getting profile").
			SetEphemeral(true).
			Build(),
		)
		if err != nil {
			log.ErrorContext(ctx, "Error replying to interaction", "error", err)
		}
		return
	}

	embed := c.createEmbed(profile)
	err = event.CreateMessage(discord.NewMessageCreateBuilder().
		AddEmbeds(embed).
		Build(),
	)
	if err != nil {
		log.ErrorContext(ctx, "Error replying to interaction", "error", err)
	}
}

// HandleHTTP is the handler for the command that is called when the HTTP request is triggered.
func (c *Profile) HandleHTTP(ctx fiber.Ctx) error {
	log := logger.FromContext(ctx.UserContext()).With("command", c.Name())

	gid, err := api.Params(ctx, "guildID", snowflake.Parse)
	if err != nil {
		log.DebugContext(ctx.Context(), "Error parsing guild ID", "error", err)
		return api.BadRequestResponse(ctx, "missing or invalid guild ID")
	}

	typ, err := api.Params(ctx, "name", func(s string) (string, error) {
		switch s {
		case "user", "guild":
			return s, nil
		default:
			return "", fmt.Errorf("invalid profile type. Options: %q, %q", "user", "guild")
		}
	})
	if err != nil {
		log.DebugContext(ctx.Context(), "Error getting name", "error", err)
		return api.BadRequestResponse(ctx, err.Error())
	}

	username := ctx.Query("username")
	if typ == "user" && username == "" {
		return api.BadRequestResponse(ctx, "missing username")
	}

	profile, err := c.service.GetProfile(ctx.Context(), &guild.RequestProfile{
		Type:    typ,
		GuildID: gid,
		User:    username,
	})
	if err != nil {
		log.ErrorContext(ctx.Context(), "Error getting profile", "error", err)
		return api.InternalServerErrorResponse(ctx, "Error while getting profile")
	}

	return ctx.Status(http.StatusOK).JSON(fiber.Map{"profile": profile})
}

// Route returns the route of the command.
func (c *Profile) Route() (methods []string, path string) {
	return []string{http.MethodGet}, "/guilds/:guildID/profile/:name"
}

// Info returns the interaction command information.
func (c *Profile) Info() discord.ApplicationCommandCreate {
	return NewInfoBuilder().
		Name(c.Name(), map[discord.Locale]string{
			discord.LocaleGerman: "profil",
		}).
		Description("Gets the profile of the given chosen option.", map[discord.Locale]string{
			discord.LocaleGerman: "Gibt das Profil der gewählten Option zurück.",
		}).
		Option(NewStringOptionBuilder().
			Name("name", nil).
			Description("The name of the profile. Note that on user profile requests, you need to provide the user's name.", map[discord.Locale]string{
				discord.LocaleGerman: "Der Name des Profils. Bei User-Anfragen muss der Name des Benutzers angegeben werden.",
			}).
			Required(true).
			Choices([]discord.ApplicationCommandOptionChoiceString{
				{
					Name:              "user",
					NameLocalizations: nil,
					Value:             "user",
				},
				{
					Name: "guild",
					NameLocalizations: map[discord.Locale]string{
						discord.LocaleGerman: "Gilde",
					},
					Value: "guild",
				},
			}...).
			Build(),
		).
		Option(NewStringOptionBuilder().
			Name("username", nil).
			Description("The username to get the profile from.", map[discord.Locale]string{
				discord.LocaleGerman: "Der Benutzername, von dem das Profil abgerufen werden soll.",
			}).
			Required(false).
			Build(),
		).Build()
}

func (c *Profile) createEmbed(profile *guild.Profiles) discord.Embed {
	embed := discord.NewEmbedBuilder().
		SetTitle("Profile").
		SetDescription("Here is the profile you requested.").
		SetColor(red)

	if profile.IsGuild() {
		embed.AddFields(
			discord.EmbedField{
				Name:   profile.GuildProfile.Name,
				Value:  fmt.Sprintf("Region: %s\nRealm: %s\nFaction: %s", profile.GuildProfile.Region, profile.GuildProfile.Realm, profile.GuildProfile.Faction),
				Inline: toPtr(true),
			},
		)
	}

	if profile.IsUser() {
		embed.AddFields(
			discord.EmbedField{
				Name:   profile.UserProfile.Name,
				Value:  fmt.Sprintf("Region: %s\nRealm: %s\nFaction: %s", profile.UserProfile.Region, profile.UserProfile.Realm, profile.UserProfile.Faction),
				Inline: toPtr(true),
			},
		)
	}

	return embed.Build()
}
