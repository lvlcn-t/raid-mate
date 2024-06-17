package feedback

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/disgoorg/disgo/bot"
	gh "github.com/google/go-github/v62/github"
	"github.com/lvlcn-t/loggerhead/logger"
)

// githubConfig is the configuration for the GitHub service.
type githubConfig struct {
	// Owner is the owner of the repository.
	Owner string `yaml:"owner" mapstructure:"owner"`
	// Repo is the repository name.
	Repo string `yaml:"repo" mapstructure:"repo"`
	// Token is the GitHub token to authenticate.
	Token string `yaml:"token" mapstructure:"token"`
}

func (c *githubConfig) Validate() error {
	var err error
	if c.Owner == "" {
		err = errors.New("owner is required")
	}
	if c.Repo == "" {
		err = errors.Join(err, errors.New("repo is required"))
	}
	if c.Token == "" {
		err = errors.Join(err, errors.New("token is required"))
	}
	return err
}

type github struct {
	config *githubConfig
	client githubAPI
}

func newGitHub(c *githubConfig) *github {
	return &github{
		config: c,
		client: newGitHubClient(c.Token),
	}
}

func (s *github) Submit(ctx context.Context, req Request, _ bot.Client) error {
	log := logger.FromContext(ctx)
	r := &reqIssue{
		Title:  fmt.Sprintf("Feedback from %s", req.Username),
		Body:   fmt.Sprintf("Feedback from %s in %s\n\n%s", req.Username, req.Server, req.Feedback),
		Labels: []string{"feedback"},
	}

	resp, err := s.createIssue(ctx, r)
	if err != nil {
		log.ErrorContext(ctx, "Error while creating issue", "error", err)
		return err
	}

	log.InfoContext(ctx, "Issue created", "issue-id", resp.IssueId, "url", resp.Url)
	return nil
}

type reqIssue struct {
	Title  string
	Body   string
	Labels []string
}

type respIssue struct {
	IssueId int32
	Url     string
	Message string
}

func (s *github) createIssue(ctx context.Context, req *reqIssue) (*respIssue, error) {
	repo, err := s.client.GetRepository(ctx, s.config.Owner, s.config.Repo)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repo: %v", err)
	}

	reqIssue := &gh.IssueRequest{Title: &req.Title, Body: &req.Body, Labels: &req.Labels}
	issue, err := s.client.CreateIssue(ctx, repo, reqIssue)
	if err != nil {
		return nil, fmt.Errorf("failed to create issue: %v", err)
	}

	return &respIssue{
		IssueId: int32(issue.GetID()),
		Url:     issue.GetHTMLURL(),
		Message: "Issue created successfully",
	}, nil
}

type githubAPI interface {
	GetRepository(ctx context.Context, owner, repo string) (*gh.Repository, error)
	CreateIssue(ctx context.Context, repo *gh.Repository, issue *gh.IssueRequest) (*gh.Issue, error)
}

type ghClient struct{ *gh.Client }

func newGitHubClient(token string) *ghClient {
	return &ghClient{
		Client: gh.NewClient(nil).WithAuthToken(token),
	}
}

func (g *ghClient) GetRepository(ctx context.Context, owner, repo string) (*gh.Repository, error) {
	r, resp, err := g.Client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		if resp.StatusCode >= http.StatusBadRequest {
			return nil, fmt.Errorf("HTTP error %d: %w", resp.StatusCode, err)
		}
		return nil, err
	}
	return r, nil
}

func (g *ghClient) CreateIssue(ctx context.Context, repo *gh.Repository, issue *gh.IssueRequest) (*gh.Issue, error) {
	if repo.Owner == nil {
		return nil, errors.New("repository owner is nil")
	}

	i, resp, err := g.Client.Issues.Create(ctx, repo.Owner.GetLogin(), repo.GetName(), issue)
	if err != nil {
		if resp.StatusCode >= http.StatusBadRequest {
			return nil, fmt.Errorf("HTTP error %d: %v", resp.StatusCode, err)
		}
		return nil, err
	}
	return i, nil
}
