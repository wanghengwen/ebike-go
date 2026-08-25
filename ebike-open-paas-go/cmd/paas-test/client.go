package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Envelope is the Xiaoan success/failure body. Business errors still return HTTP 200.
type Envelope struct {
	Success bool            `json:"success"`
	Data    json.RawMessage `json:"data,omitempty"`
	Error   *struct {
		ErrorMessage string `json:"errormessage"`
		Promot       string `json:"promot"`
	} `json:"error,omitempty"`
}

// Client talks to the open PaaS the way a third party would.
type Client struct {
	cfg     Config
	http    *http.Client
	verbose bool
}

func newClient(cfg Config) *Client {
	return &Client{
		cfg: cfg,
		http: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 5 {
					return fmt.Errorf("stopped after %d redirects (likely an Ingress ssl-redirect loop; check force-ssl-redirect / X-Forwarded-Proto)", len(via))
				}
				if len(via) > 0 && req.URL.String() == via[len(via)-1].URL.String() {
					return fmt.Errorf("redirect loop to %s (Ingress is 308ing the same URL; remove force-ssl-redirect or fix X-Forwarded-Proto)", req.URL)
				}
				return nil
			},
		},
	}
}

func (c *Client) get(path string, query url.Values) (*Envelope, int, error) {
	if query == nil {
		query = url.Values{}
	}
	query.Set("agentId", c.cfg.AgentID)
	return c.do(http.MethodGet, path+"?"+query.Encode(), nil)
}

func (c *Client) delete(path string, query url.Values) (*Envelope, int, error) {
	if query == nil {
		query = url.Values{}
	}
	query.Set("agentId", c.cfg.AgentID)
	return c.do(http.MethodDelete, path+"?"+query.Encode(), nil)
}

func (c *Client) post(path string, body map[string]interface{}) (*Envelope, int, error) {
	if body == nil {
		body = map[string]interface{}{}
	}
	body["agentId"] = c.cfg.AgentID
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, 0, err
	}
	return c.do(http.MethodPost, path, raw)
}

func (c *Client) do(method, path string, body []byte) (*Envelope, int, error) {
	u := c.cfg.BaseURL + path
	var rdr io.Reader
	if body != nil {
		rdr = bytes.NewReader(body)
	}
	req, err := http.NewRequest(method, u, rdr)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("xc-access-token", c.cfg.Token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, resp.StatusCode, err
	}
	if c.verbose {
		fmt.Printf("→ %s %s\n← %d %s\n", method, u, resp.StatusCode, truncate(string(raw), 2000))
	}

	var env Envelope
	if err := json.Unmarshal(raw, &env); err != nil {
		return nil, resp.StatusCode, fmt.Errorf("response is not JSON (status %d): %s", resp.StatusCode, truncate(string(raw), 200))
	}
	return &env, resp.StatusCode, nil
}

func (e *Envelope) errorType() string {
	if e == nil || e.Error == nil {
		return ""
	}
	return e.Error.ErrorMessage
}

func (e *Envelope) promot() string {
	if e == nil || e.Error == nil {
		return ""
	}
	return e.Error.Promot
}

func printEnvelope(label string, env *Envelope, status int) {
	pretty, _ := json.MarshalIndent(env, "", "  ")
	fmt.Printf("[%s] http=%d\n%s\n", label, status, pretty)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func mustHTTPS(raw string) error {
	raw = strings.TrimSpace(raw)
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("callback url is not a valid URL: %w", err)
	}
	if u.Scheme != "https" {
		return fmt.Errorf("callback url must use https (got %q)", u.Scheme)
	}
	if u.Hostname() == "" {
		return fmt.Errorf("callback url has no host")
	}
	return nil
}
