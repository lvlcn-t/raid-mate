package bot

import (
	"strings"

	"github.com/bwmarrin/discordgo"
)

// IntentsConfig defines how intents are configured.
type IntentsConfig struct {
	// Unprivileged is whether to use unprivileged intents.
	Unprivileged bool `yaml:"unprivileged"`
	// Privileged is the list of privileged intents.
	Privileged []string `yaml:"privileged"`
}

// List returns the list of intents based on the configuration.
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

// Intent defines a Discord intent.
type Intent struct {
	// Name is the name of the intent.
	Name string `yaml:"-"`
	// Value is the value of the intent.
	Value discordgo.Intent `yaml:"-"`
}

var (
	// IntentGuildMembers is the intent for guild members.
	IntentGuildMembers = Intent{
		Name:  "Guild Members",
		Value: discordgo.IntentGuildMembers,
	}

	// IntentGuildPresences is the intent for guild presences.
	IntentGuildPresences = Intent{
		Name:  "Guild Presences",
		Value: discordgo.IntentGuildPresences,
	}

	// IntentMessageContent is the intent for message content.
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
