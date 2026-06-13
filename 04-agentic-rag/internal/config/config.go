package config

import (
	"strings"
	"time"

	env "github.com/Netflix/go-env"
)

const defaultOllamaBaseURL = "http://127.0.0.1:11434"

type Config struct {
	OllamaHost             string        `env:"OLLAMA_HOST"`
	OllamaModel            string        `env:"OLLAMA_MODEL,default=gemma3:latest"`
	AgentModel             string        `env:"AGENT_MODEL"`
	OllamaEmbedModel       string        `env:"OLLAMA_EMBED_MODEL,default=nomic-embed-text"`
	AgentMaxIterations     int           `env:"AGENT_MAX_ITERATIONS,default=5"`
	RAGTopK                int           `env:"RAG_TOP_K,default=4"`
	OllamaTemperature      float64       `env:"OLLAMA_TEMPERATURE,default=0.8"`
	OllamaHealthTimeout    time.Duration `env:"OLLAMA_HEALTH_TIMEOUT,default=2s"`
	OllamaHealthcheckEvery time.Duration `env:"OLLAMA_HEALTHCHECK_INTERVAL,default=5s"`
}

func Load() (Config, error) {
	var cfg Config
	_, err := env.UnmarshalFromEnviron(&cfg)
	return cfg, err
}

func (c Config) OllamaBaseURL() string {
	host := strings.TrimSpace(c.OllamaHost)
	if host == "" {
		return defaultOllamaBaseURL
	}
	if !strings.Contains(host, "://") {
		host = "http://" + host
	}
	return strings.TrimRight(host, "/")
}

// AgentModelName returns the model used for tool-calling agent turns.
func (c Config) AgentModelName() string {
	if model := strings.TrimSpace(c.AgentModel); model != "" {
		return model
	}
	return strings.TrimSpace(c.OllamaModel)
}
