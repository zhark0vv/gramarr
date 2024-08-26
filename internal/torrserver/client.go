package torrserver

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-resty/resty/v2"
)

type Client struct {
	username string
	password string
	baseURL  string
	client   *resty.Client
}

func New(c Config) (*Client, error) {
	if c.Hostname == "" {
		return nil, fmt.Errorf("hostname is empty")
	}

	baseURL := createApiURL(c)

	r := resty.New()
	r.SetBaseURL(baseURL)
	r.SetHeader("Accept", "application/json")
	if c.Username != "" && c.Password != "" {
		r.SetBasicAuth(c.Username, c.Password)
	}

	client := &Client{
		username: c.Username,
		password: c.Password,
		baseURL:  baseURL,
		client:   r,
	}
	return client, nil
}

func createApiURL(c Config) string {
	c.Hostname = strings.TrimPrefix(c.Hostname, "http://")
	c.Hostname = strings.TrimPrefix(c.Hostname, "https://")

	u := url.URL{}

	u.Scheme = "http"
	u.Host = c.Hostname
	if c.Port != 80 {
		u.Host = fmt.Sprintf("%s:%d", c.Hostname, c.Port)
	}

	return u.String()
}

func (c *Client) AddTorrent(torrentLink, posterLink string) error {
	resp, err := c.client.R().
		SetBody(map[string]any{
			"action":     "add",
			"link":       torrentLink,
			"poster":     posterLink,
			"save_to_db": true,
		}).
		Post("torrents")

	if err != nil {
		return fmt.Errorf("error adding torrent to torrent endpoint: %w", err)
	}
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("error adding torrent to torrent endpoint: status %d", resp.StatusCode())
	}

	return nil
}
