package config

import (
	"os"
	"testing"
	"time"
)

func TestLoadDefaults(t *testing.T) {
	os.Unsetenv("OLLAMA_HOST")
	os.Unsetenv("OLLAMA_MODEL")
	os.Unsetenv("OLLAMA_TEMPERATURE")
	os.Unsetenv("OLLAMA_HEALTH_TIMEOUT")
	os.Unsetenv("OLLAMA_HEALTHCHECK_INTERVAL")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.OllamaModel != "gemma3:latest" {
		t.Fatalf("OllamaModel = %q, want gemma3:latest", cfg.OllamaModel)
	}
	if cfg.OllamaTemperature != 0.8 {
		t.Fatalf("OllamaTemperature = %v, want 0.8", cfg.OllamaTemperature)
	}
	if cfg.OllamaHealthTimeout != 2*time.Second {
		t.Fatalf("OllamaHealthTimeout = %v, want 2s", cfg.OllamaHealthTimeout)
	}
	if cfg.OllamaHealthcheckEvery != 5*time.Second {
		t.Fatalf("OllamaHealthcheckEvery = %v, want 5s", cfg.OllamaHealthcheckEvery)
	}
	if got := cfg.OllamaBaseURL(); got != defaultOllamaBaseURL {
		t.Fatalf("OllamaBaseURL() = %q, want %q", got, defaultOllamaBaseURL)
	}
}

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("OLLAMA_HOST", "192.168.1.10:11434")
	t.Setenv("OLLAMA_MODEL", "llama3:latest")
	t.Setenv("OLLAMA_TEMPERATURE", "0.5")
	t.Setenv("OLLAMA_HEALTH_TIMEOUT", "3s")
	t.Setenv("OLLAMA_HEALTHCHECK_INTERVAL", "10s")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.OllamaModel != "llama3:latest" {
		t.Fatalf("OllamaModel = %q, want llama3:latest", cfg.OllamaModel)
	}
	if cfg.OllamaTemperature != 0.5 {
		t.Fatalf("OllamaTemperature = %v, want 0.5", cfg.OllamaTemperature)
	}
	if cfg.OllamaHealthTimeout != 3*time.Second {
		t.Fatalf("OllamaHealthTimeout = %v, want 3s", cfg.OllamaHealthTimeout)
	}
	if cfg.OllamaHealthcheckEvery != 10*time.Second {
		t.Fatalf("OllamaHealthcheckEvery = %v, want 10s", cfg.OllamaHealthcheckEvery)
	}
	if got := cfg.OllamaBaseURL(); got != "http://192.168.1.10:11434" {
		t.Fatalf("OllamaBaseURL() = %q, want http://192.168.1.10:11434", got)
	}
}

func TestOllamaBaseURLPreservesScheme(t *testing.T) {
	cfg := Config{OllamaHost: "https://ollama.example.com"}
	if got := cfg.OllamaBaseURL(); got != "https://ollama.example.com" {
		t.Fatalf("OllamaBaseURL() = %q, want https://ollama.example.com", got)
	}
}

func TestLoadClearsUnsetVars(t *testing.T) {
	os.Unsetenv("OLLAMA_MODEL")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.OllamaModel != "gemma3:latest" {
		t.Fatalf("OllamaModel = %q, want default", cfg.OllamaModel)
	}
}
