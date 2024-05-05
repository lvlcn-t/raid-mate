package services

// Collection is the collection of services.
type Collection struct {
	GitHub GitHub
	Guild  Guild
}

// NewCollection creates a new collection of services.
func NewCollection() (Collection, error) {
	gh, err := NewGitHubService()
	if err != nil {
		return Collection{}, err
	}

	g, err := NewGuildService()
	if err != nil {
		return Collection{}, err
	}

	return Collection{
		GitHub: gh,
		Guild:  g,
	}, nil
}

func (c *Collection) Connect() error {
	// TODO: connect all services
	return nil
}

func (c *Collection) Close() error {
	// TODO: close all services
	return nil
}
