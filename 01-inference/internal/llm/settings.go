package llm

import (
	"net/http"
	"time"

	"github.com/tmc/langchaingo/llms/ollama"
)

const defaultHealthTimeout = 2 * time.Second

type Settings struct {
	ModelName     string
	BaseURL       string
	Temperature   float64
	HealthTimeout time.Duration
}

type Client struct {
	settings Settings
	model    *ollama.LLM
	http     *http.Client
}

func NewClient(settings Settings) (*Client, error) {
	httpClient := &http.Client{}

	opts := []ollama.Option{
		ollama.WithModel(settings.ModelName),
		ollama.WithHTTPClient(httpClient),
	}
	if settings.BaseURL != "" {
		opts = append(opts, ollama.WithServerURL(settings.BaseURL))
	}

	model, err := ollama.New(opts...)
	if err != nil {
		return nil, err
	}

	return &Client{
		settings: settings,
		model:    model,
		http:     httpClient,
	}, nil
}

func (c *Client) ModelName() string {
	return c.settings.ModelName
}
