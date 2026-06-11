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
