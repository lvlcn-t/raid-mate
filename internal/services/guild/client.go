package guild

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const (
	baseURL  = "https://www.warcraftlogs.com"
	msPerSec = 1000
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

func (c *client) FetchReport(ctx context.Context, guild Request, date time.Time) (reports []report, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/v1/reports/guild/%s/%s/%s", baseURL, guild.Name, guild.ServerName, guild.ServerRegion), http.NoBody)
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
		return nil, errors.New("unexpected status code")
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
