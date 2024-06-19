package feedback

import (
	"context"
	"errors"
	"fmt"

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

func (c *dmConfig) Validate() error {
	var err error
	if c.ID == "" {
		err = errors.New("services.feedback.dm.id is required")
	}

	_, pErr := snowflake.Parse(c.ID)
	if pErr != nil {
		err = errors.Join(err, fmt.Errorf("services.feedback.dm.id is invalid: %w", pErr))
	}
	return err
}

type dm struct {
	id snowflake.ID
}

func newDM(c *dmConfig) *dm {
	id, err := snowflake.Parse(c.ID)
	if err != nil {
		id = snowflake.MustParse("null")
	}

	return &dm{
		id: id,
	}
}

func (s *dm) Submit(ctx context.Context, req Request, client bot.Client) error {
	if client == nil {
		return nil
	}

	log := logger.FromContext(ctx)
	dm, err := client.Rest().CreateDMChannel(s.id, rest.WithCtx(ctx))
	if err != nil {
		log.ErrorContext(ctx, "Error while creating DM channel", "error", err)
		return err
	}

	am := discord.DefaultAllowedMentions
	am.Users = append(am.Users, req.UserID)
	m, err := client.Rest().CreateMessage(dm.ID(), discord.NewMessageCreateBuilder().
		SetAllowedMentions(&am).
		SetContentf("Feedback from <@%s> in %s\n\n%s", req.Username, req.Server, req.Feedback).
		SetEphemeral(false).
		Build())
	if err != nil {
		log.ErrorContext(ctx, "Error while sending message", "error", err)
		return err
	}

	log.InfoContext(ctx, "Sent direct message", "message-id", m.ID, "user", s.id)
	return nil
}
