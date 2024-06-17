package bot

import (
	"errors"
	"fmt"

	"github.com/disgoorg/disgo/gateway"
)

// IntentsConfig defines how intents are configured.
type IntentsConfig struct {
	// Unprivileged is whether to use unprivileged intents.
	Unprivileged bool `yaml:"unprivileged"`
	// Privileged is the list of privileged intents.
	Privileged []string `yaml:"privileged"`
}

// Validate validates the configuration.
func (c *IntentsConfig) Validate() error {
	var err error
	if len(c.Privileged) > 3 {
		err = errors.New("too many privileged intents")
	}

	for _, i := range c.Privileged {
		if _, ok := intentRegistry[Intent(i)]; !ok {
			err = errors.Join(err, fmt.Errorf("unknown intent: %q", i))
		}
	}

	return err
}

// List returns the list of intents based on the configuration.
func (c IntentsConfig) List() []gateway.Intents {
	var intents []gateway.Intents
	if c.Unprivileged {
		intents = append(intents, gateway.IntentsNonPrivileged)
	} else {
		intents = append(intents, gateway.IntentsNone)
	}

	for _, i := range c.Privileged {
		if intent, ok := intentRegistry[Intent(i)]; ok {
			intents = append(intents, intent)
		}
	}

	return intents
}

// Intent is an intent for the bot.
type Intent string

const (
	// IntentGuilds is the intent for guilds. This is a privileged intent.
	IntentGuilds Intent = "guilds"
	// IntentGuildMessages is the intent for guild messages. This is a privileged intent.
	IntentGuildMessages Intent = "guildMessages"
	// IntentDirectMessages is the intent for direct messages. This is a privileged intent.
	IntentDirectMessages Intent = "directMessages"
)

// intentRegistry is the registry of privileged intents.
var intentRegistry = map[Intent]gateway.Intents{
	IntentGuilds:         gateway.IntentGuilds,
	IntentGuildMessages:  gateway.IntentGuildMessages,
	IntentDirectMessages: gateway.IntentDirectMessages,
}
