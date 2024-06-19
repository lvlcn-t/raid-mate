package guild

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/lvlcn-t/raid-mate/internal/database/repo"
)

const (
	logsBaseURL    = "https://www.warcraftlogs.com"
	profileBaseURL = "https://raider.io"
	msPerSec       = 1000
)

type client struct {
	client http.Client
	token  string
}

func NewClient(token string, timeout time.Duration) *client {
	return &client{
		client: http.Client{
			Timeout: timeout,
		},
		token: token,
	}
}

type report struct {
	Id    string `json:"id"`
	Title string `json:"title"`
	Owner string `json:"owner"`
	Zone  int    `json:"zone"`
	Start int    `json:"start"`
	End   int    `json:"end"`
}

func (c *client) FetchReports(ctx context.Context, guild repo.Guild, date time.Time) (reports []report, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/v1/reports/guild/%s/%s/%s", logsBaseURL, guild.Name, guild.ServerName, guild.ServerRegion), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	query := req.URL.Query()
	query.Add("start", fmt.Sprintf("%d", date.Unix()*msPerSec))
	query.Add("end", fmt.Sprintf("%d", date.AddDate(0, 0, 1).Unix()*msPerSec))
	req.URL.RawQuery = query.Encode()

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer func() {
		err = errors.Join(err, resp.Body.Close())
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	err = json.Unmarshal(b, &reports)
	if err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return reports, nil
}

type Profiles struct {
	UserProfile  *UserProfile  `json:"user_profile,omitempty"`
	GuildProfile *GuildProfile `json:"guild_profile,omitempty"`
}

func (p *Profiles) IsUser() bool {
	return p.UserProfile != nil
}

func (p *Profiles) IsGuild() bool {
	return p.GuildProfile != nil
}

type baseProfile struct {
	Name            string                     `json:"name"`
	Faction         string                     `json:"faction"`
	Region          string                     `json:"region"`
	Realm           string                     `json:"realm"`
	ProfileURL      string                     `json:"profile_url"`
	RaidProgression map[string]RaidProgression `json:"raid_progression"`
}

type GuildProfile struct {
	baseProfile
	RaidRankings map[string]RaidRanking `json:"raid_rankings"`
}

type UserProfile struct {
	baseProfile
	Race                     string                     `json:"race"`
	Class                    string                     `json:"class"`
	ActiveSpecName           string                     `json:"active_spec_name"`
	ActiveSpecRole           string                     `json:"active_spec_role"`
	Gender                   string                     `json:"gender"`
	Gear                     Gear                       `json:"gear"`
	MythicPlusScoresBySeason []MythicPlusScoresBySeason `json:"mythic_plus_scores_by_season"`
	MythicPlusRanks          MythicPlusRanks            `json:"mythic_plus_ranks"`
	PreviousMythicPlusRanks  MythicPlusRanks            `json:"previous_mythic_plus_ranks"`
	MythicPlusRecentRuns     []MythicPlusRun            `json:"mythic_plus_recent_runs"`
	MythicPlusBestRuns       []MythicPlusRun            `json:"mythic_plus_best_runs"`
	MythicPlusAlternateRuns  []MythicPlusRun            `json:"mythic_plus_alternate_runs"`
}

type Gear struct {
	ItemLevelEquipped int `json:"item_level_equipped"`
	ItemLevelTotal    int `json:"item_level_total"`
	ArtifactTraits    int `json:"artifact_traits"`
}

type MythicPlusRun struct {
	Dungeon             string    `json:"dungeon"`
	ShortName           string    `json:"short_name"`
	MythicLevel         int       `json:"mythic_level"`
	CompletedAt         time.Time `json:"completed_at"`
	ClearTimeMS         int       `json:"clear_time_ms"`
	NumKeystoneUpgrades int       `json:"num_keystone_upgrades"`
	Score               float64   `json:"score"`
	URL                 string    `json:"url"`
}

type MythicPlusRanks struct {
	Overall     Class `json:"overall"`
	Tank        Class `json:"tank"`
	Healer      Class `json:"healer"`
	Dps         Class `json:"dps"`
	Class       Class `json:"class"`
	ClassTank   Class `json:"class_tank"`
	ClassHealer Class `json:"class_healer"`
	ClassDps    Class `json:"class_dps"`
}

type Class struct {
	World  int `json:"world"`
	Region int `json:"region"`
	Realm  int `json:"realm"`
}

type MythicPlusScoresBySeason struct {
	Season   string   `json:"season"`
	Scores   Scores   `json:"scores"`
	Segments Segments `json:"segments"`
}

type Scores struct {
	All    int `json:"all"`
	Dps    int `json:"dps"`
	Healer int `json:"healer"`
	Tank   int `json:"tank"`
	Spec0  int `json:"spec_0"`
	Spec1  int `json:"spec_1"`
	Spec2  int `json:"spec_2"`
	Spec3  int `json:"spec_3"`
}

type Segments struct {
	All    All `json:"all"`
	Dps    All `json:"dps"`
	Healer All `json:"healer"`
	Tank   All `json:"tank"`
	Spec0  All `json:"spec_0"`
	Spec1  All `json:"spec_1"`
	Spec2  All `json:"spec_2"`
	Spec3  All `json:"spec_3"`
}

type All struct {
	Score int    `json:"score"`
	Color string `json:"color"`
}

type RaidProgression struct {
	Summary            string `json:"summary"`
	TotalBosses        int    `json:"total_bosses"`
	NormalBossesKilled int    `json:"normal_bosses_killed"`
	HeroicBossesKilled int    `json:"heroic_bosses_killed"`
	MythicBossesKilled int    `json:"mythic_bosses_killed"`
}

type RaidRanking struct {
	Normal Stats `json:"normal"`
	Heroic Stats `json:"heroic"`
	Mythic Stats `json:"mythic"`
}

type Stats struct {
	World  int `json:"world"`
	Region int `json:"region"`
	Realm  int `json:"realm"`
}

func (c *client) FetchProfile(ctx context.Context, req *RequestProfile) (*Profiles, error) {
	switch req.Type {
	case "guild":
		p, err := c.getGuildProfile(ctx, req)
		if err != nil {
			return nil, err
		}
		return &Profiles{GuildProfile: p}, nil
	case "user":
		p, err := c.getUserProfile(ctx, req)
		if err != nil {
			return nil, err
		}
		return &Profiles{UserProfile: p}, nil
	default:
		return nil, errors.New("invalid profile type")
	}
}

func (c *client) getGuildProfile(ctx context.Context, r *RequestProfile) (profile *GuildProfile, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/api/v1/characters/profile", profileBaseURL), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	query := req.URL.Query()
	query.Add("region", r.guild.ServerRegion)
	query.Add("realm", r.guild.ServerRealm)
	query.Add("name", r.guild.Name)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer func() {
		err = errors.Join(err, resp.Body.Close())
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	err = json.Unmarshal(b, &profile)
	if err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return profile, nil
}

func (c *client) getUserProfile(ctx context.Context, r *RequestProfile) (profile *UserProfile, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/api/v1/guilds/profile", profileBaseURL), http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))

	query := req.URL.Query()
	query.Add("region", r.guild.ServerRegion)
	query.Add("realm", r.guild.ServerRealm)
	query.Add("name", r.User)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer func() {
		err = errors.Join(err, resp.Body.Close())
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	err = json.Unmarshal(b, &profile)
	if err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return profile, nil
}
