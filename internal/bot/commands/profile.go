package commands

import (
	"context"
	"strconv"

	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/events"
	"github.com/lvlcn-t/loggerhead/logger"
	"github.com/lvlcn-t/raid-mate/internal/services/guild"
)

var (
	_ Command[*events.ApplicationCommandInteractionCreate] = (*Profile)(nil)
	_ InteractionCommand                                   = (*Profile)(nil)
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
		Base:    NewBase("profile"),
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
		log.Error("Member is nil")
		return
	}

	profile, err := c.service.GetProfile(ctx, &guild.RequestProfile{
		Type:    typ,
		GuildID: event.GuildID().String(),
		User:    username,
	})
	if err != nil {
		cErr := event.CreateMessage(discord.NewMessageCreateBuilder().
			SetContent("Error while getting profile").
			SetEphemeral(true).
			Build(),
		)
		if cErr != nil {
			log.ErrorContext(ctx, "Error replying to interaction", "error", cErr, "getProfileError", err)
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

func (c *Profile) createEmbed(profile *guild.Profiles) discord.Embed { //nolint:funlen // Can't be shortened
	embed := discord.NewEmbedBuilder().
		SetTitle("Profile").
		SetDescription("Here is the profile you requested.").
		SetColor(red)

	if profile.IsGuild() {
		embed.AddField("Name", profile.GuildProfile.Name, true)
		embed.AddField("Faction", profile.GuildProfile.Faction, true)
		embed.AddField("Region", profile.GuildProfile.Region, true)
		embed.AddField("Realm", profile.GuildProfile.Realm, true)
		embed.AddField("Profile URL", profile.GuildProfile.ProfileURL, true)
		for raid, progress := range profile.GuildProfile.RaidProgression {
			embed.AddField("Raid", raid, true)
			embed.AddField("Summary", progress.Summary, true)
			embed.AddField("Total Bosses", strconv.Itoa(progress.TotalBosses), true)
			embed.AddField("Normal Bosses Killed", strconv.Itoa(progress.NormalBossesKilled), true)
			embed.AddField("Heroic Bosses Killed", strconv.Itoa(progress.HeroicBossesKilled), true)
			embed.AddField("Mythic Bosses Killed", strconv.Itoa(progress.MythicBossesKilled), true)
		}

		for raid, ranking := range profile.GuildProfile.RaidRankings {
			embed.AddField("Raid", raid, true)
			embed.AddField("Normal World", strconv.Itoa(ranking.Normal.World), true)
			embed.AddField("Normal Region", strconv.Itoa(ranking.Normal.Region), true)
			embed.AddField("Normal Realm", strconv.Itoa(ranking.Normal.Realm), true)
			embed.AddField("Heroic World", strconv.Itoa(ranking.Heroic.World), true)
			embed.AddField("Heroic Region", strconv.Itoa(ranking.Heroic.Region), true)
			embed.AddField("Heroic Realm", strconv.Itoa(ranking.Heroic.Realm), true)
			embed.AddField("Mythic World", strconv.Itoa(ranking.Mythic.World), true)
			embed.AddField("Mythic Region", strconv.Itoa(ranking.Mythic.Region), true)
			embed.AddField("Mythic Realm", strconv.Itoa(ranking.Mythic.Realm), true)
		}
	}

	if profile.IsUser() {
		embed.AddField("Name", profile.UserProfile.Name, true)
		embed.AddField("Faction", profile.UserProfile.Faction, true)
		embed.AddField("Region", profile.UserProfile.Region, true)
		embed.AddField("Realm", profile.UserProfile.Realm, true)
		embed.AddField("Profile URL", profile.UserProfile.ProfileURL, true)
		for raid, progress := range profile.UserProfile.RaidProgression {
			embed.AddField("Raid", raid, true)
			embed.AddField("Summary", progress.Summary, true)
			embed.AddField("Total Bosses", strconv.Itoa(progress.TotalBosses), true)
			embed.AddField("Normal Bosses Killed", strconv.Itoa(progress.NormalBossesKilled), true)
			embed.AddField("Heroic Bosses Killed", strconv.Itoa(progress.HeroicBossesKilled), true)
			embed.AddField("Mythic Bosses Killed", strconv.Itoa(progress.MythicBossesKilled), true)
		}

		embed.AddField("Race", profile.UserProfile.Race, true)
		embed.AddField("Class", profile.UserProfile.Class, true)
		embed.AddField("Active Spec Name", profile.UserProfile.ActiveSpecName, true)
		embed.AddField("Active Spec Role", profile.UserProfile.ActiveSpecRole, true)
		embed.AddField("Gender", profile.UserProfile.Gender, true)

		for i := range profile.UserProfile.MythicPlusScoresBySeason {
			score := &profile.UserProfile.MythicPlusScoresBySeason[i]
			embed.AddField("Season", score.Season, true)
			embed.AddField("All", strconv.Itoa(score.Scores.All), true)
			embed.AddField("DPS", strconv.Itoa(score.Scores.Dps), true)
			embed.AddField("Healer", strconv.Itoa(score.Scores.Healer), true)
			embed.AddField("Tank", strconv.Itoa(score.Scores.Tank), true)
			embed.AddField("Spec 0", strconv.Itoa(score.Scores.Spec0), true)
			embed.AddField("Spec 1", strconv.Itoa(score.Scores.Spec1), true)
			embed.AddField("Spec 2", strconv.Itoa(score.Scores.Spec2), true)
			embed.AddField("Spec 3", strconv.Itoa(score.Scores.Spec3), true)
			embed.AddField("All Segment Score", strconv.Itoa(score.Segments.All.Score), true)
			embed.AddField("All Segment Color", score.Segments.All.Color, true)
			embed.AddField("DPS Segment Score", strconv.Itoa(score.Segments.Dps.Score), true)
			embed.AddField("DPS Segment Color", score.Segments.Dps.Color, true)
			embed.AddField("Healer Segment Score", strconv.Itoa(score.Segments.Healer.Score), true)
			embed.AddField("Healer Segment Color", score.Segments.Healer.Color, true)
			embed.AddField("Tank Segment Score", strconv.Itoa(score.Segments.Tank.Score), true)
			embed.AddField("Tank Segment Color", score.Segments.Tank.Color, true)
			embed.AddField("Spec 0 Segment Score", strconv.Itoa(score.Segments.Spec0.Score), true)
			embed.AddField("Spec 0 Segment Color", score.Segments.Spec0.Color, true)
			embed.AddField("Spec 1 Segment Score", strconv.Itoa(score.Segments.Spec1.Score), true)
			embed.AddField("Spec 1 Segment Color", score.Segments.Spec1.Color, true)
			embed.AddField("Spec 2 Segment Score", strconv.Itoa(score.Segments.Spec2.Score), true)
			embed.AddField("Spec 2 Segment Color", score.Segments.Spec2.Color, true)
			embed.AddField("Spec 3 Segment Score", strconv.Itoa(score.Segments.Spec3.Score), true)
			embed.AddField("Spec 3 Segment Color", score.Segments.Spec3.Color, true)
		}
	}

	return embed.Build()
}
