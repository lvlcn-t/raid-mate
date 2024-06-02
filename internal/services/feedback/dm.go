package feedback

import "context"

// dmConfig is the configuration for the DM service.
type dmConfig struct {
	// ID is the Discord user ID.
	ID string `yaml:"id" mapstructure:"id"`
}

type dm struct {
	config *dmConfig
}

func newDM(c *dmConfig) *dm {
	return &dm{
		config: c,
	}
}

func (s *dm) Submit(_ context.Context, _ string) error {
	return nil
}
