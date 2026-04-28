package api

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	cfgpkg "github.com/oddship/wg-tui/internal/config"
)

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

type Info struct {
	Version string `json:"version"`
	Ports   struct {
		SSH int `json:"ssh"`
	} `json:"ports"`
}

type Target struct {
	Name                string `json:"name"`
	Description         string `json:"description"`
	Kind                string `json:"kind"`
	ExternalHost        string `json:"external_host"`
	DefaultDatabaseName string `json:"default_database_name"`
	Group               struct {
		Name string `json:"name"`
	} `json:"group"`
}

func New(cfg cfgpkg.Config) *Client {
	return &Client{
		baseURL: strings.TrimRight(cfg.Server.URL, "/"),
		token:   cfg.Server.Token,
		http:    &http.Client{Timeout: 15 * time.Second, Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: cfg.Server.InsecureSkipTLSVerify}}},
	}
}

func (c *Client) GetInfo(ctx context.Context) (Info, error) {
	var out Info
	err := c.get(ctx, "/@warpgate/api/info", &out)
	return out, err
}

func (c *Client) GetTargets(ctx context.Context) ([]Target, error) {
	var out []Target
	err := c.get(ctx, "/@warpgate/api/targets", &out)
	return out, err
}

func (c *Client) get(ctx context.Context, path string, out any) error {
	u, err := url.Parse(c.baseURL + path)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return err
	}
	if c.token != "" {
		req.Header.Set("X-Warpgate-Token", c.token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("api %s: %s", path, resp.Status)
	}
	return json.NewDecoder(resp.Body).Decode(out)
}
