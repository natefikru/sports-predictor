package kalshi

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)


const baseURL = "https://external-api.kalshi.com/trade-api/v2"

type Client struct {
	http *http.Client
}

func NewClient() *Client {
	return &Client{
		http: &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) get(path string, params map[string]string) ([]byte, error) {
	u, err := url.Parse(baseURL + path)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()

	// Retry once on 429 after a backoff pause.
	for attempt := 0; attempt <= 1; attempt++ {
		if attempt > 0 {
			time.Sleep(1500 * time.Millisecond)
		}
		resp, err := c.http.Get(u.String())
		if err != nil {
			return nil, err
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}
		if resp.StatusCode == 429 && attempt == 0 {
			continue // retry after backoff
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			preview := body
			if len(preview) > 200 {
				preview = preview[:200]
			}
			return nil, fmt.Errorf("kalshi API %s %d: %s", path, resp.StatusCode, preview)
		}
		return body, nil
	}
	return nil, fmt.Errorf("kalshi API %s: rate limited after retry", path)
}
