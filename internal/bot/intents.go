package bot

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

type Intent struct {
	Name  string           `yaml:"Name"`
	Value discordgo.Intent `yaml:"-"`
}

var (
	IntentGuildMembers = Intent{
		Name:  "Guild Members",
		Value: discordgo.IntentGuildMembers,
	}

	IntentGuildPresences = Intent{
		Name:  "Guild Presences",
		Value: discordgo.IntentGuildPresences,
	}

	IntentMessageContent = Intent{
		Name:  "Message Content",
		Value: discordgo.IntentMessageContent,
	}

	// IntentsPrivileged is the list of intents that require privileged intents.
	IntentsPrivileged = []Intent{IntentGuildMembers, IntentGuildPresences, IntentMessageContent}

	// IntentsAll is the list of all intents.
	IntentsAll = append(IntentsPrivileged, Intent{
		Name:  "Unprivileged",
		Value: discordgo.IntentsAllWithoutPrivileged,
	})
)

// IntentsConfig defines how intents are configured.
type IntentsConfig struct {
	Unprivileged bool     `yaml:"unprivileged"`
	Privileged   []string `yaml:"privileged"`
}

func (c IntentsConfig) List() []discordgo.Intent {
	var intents []discordgo.Intent
	if c.Unprivileged {
		intents = append(intents, discordgo.IntentsAllWithoutPrivileged)
	}

	for _, name := range c.Privileged {
		for _, i := range IntentsAll {
			if strings.EqualFold(i.Name, name) {
				intents = append(intents, i.Value)
				break
			}
		}
	}

	return intents
}
