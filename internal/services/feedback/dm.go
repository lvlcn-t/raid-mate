package feedback

import (
	"context"

	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/discord"
	"github.com/disgoorg/disgo/rest"
	"github.com/disgoorg/snowflake/v2"
	"github.com/lvlcn-t/loggerhead/logger"
)

// dmConfig is the configuration for the DM service.
type dmConfig struct {
	// ID is the Discord user ID.
	ID string `yaml:"id" mapstructure:"id"`
}

type dm struct {
	id snowflake.ID
}

func newDM(c *dmConfig) *dm {
	return &dm{
		id: snowflake.MustParse(c.ID),
	}
}

func (s *dm) Submit(ctx context.Context, req Request, client bot.Client) error {
	log := logger.FromContext(ctx)
	dm, err := client.Rest().CreateDMChannel(s.id, rest.WithCtx(ctx))
	if err != nil {
		log.ErrorContext(ctx, "Error while creating DM channel", "error", err)
		return err
	}
	m, err := client.Rest().CreateMessage(dm.ID(), discord.NewMessageCreateBuilder().
		SetContentf("Feedback from %s in %s\n\n%s", req.User, req.Server, req.Feedback).
		SetEphemeral(false).
		Build())
	if err != nil {
		log.ErrorContext(ctx, "Error while sending message", "error", err)
		return err
	}

	log.InfoContext(ctx, "Sent direct message", "message-id", m.ID, "user", s.id)
	return nil
}
