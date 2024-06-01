package bot

import (
	"github.com/disgoorg/disgo/gateway"
)

// IntentsConfig defines how intents are configured.
type IntentsConfig struct {
	// Unprivileged is whether to use unprivileged intents.
	Unprivileged bool `yaml:"unprivileged"`
	// Privileged is the list of privileged intents.
	Privileged []string `yaml:"privileged"`
}

// List returns the list of intents based on the configuration.
func (c IntentsConfig) List() []gateway.Intents {
	var intents []gateway.Intents
	if c.Unprivileged {
		intents = append(intents, gateway.IntentsNonPrivileged)
	}

	for _, i := range c.Privileged {
		if intent, ok := intentRegistry[Intent(i)]; ok {
			intents = append(intents, intent)
		}
	}

	return intents
}

type Intent string

const (
	IntentGuilds         Intent = "guilds"
	IntentGuildMessages  Intent = "guildMessages"
	IntentDirectMessages Intent = "directMessages"
)

var intentRegistry = map[Intent]gateway.Intents{
	IntentGuilds:         gateway.IntentGuilds,
	IntentGuildMessages:  gateway.IntentGuildMessages,
	IntentDirectMessages: gateway.IntentDirectMessages,
}
