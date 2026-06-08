package llm

import "time"

const defaultHealthTimeout = 2 * time.Second

type Settings struct {
	ModelName     string
	BaseURL       string
	Temperature   float64
	HealthTimeout time.Duration
}

type Client struct {
	settings Settings
}

func NewClient(settings Settings) *Client {
	return &Client{settings: settings}
}

func (c *Client) ModelName() string {
	return c.settings.ModelName
}
