# 01-inference

Workshop stage 1: connect a Wails desktop app to a local Ollama instance and stream LLM responses for a basic prompt.

The Go backend loads configuration from environment variables, probes Ollama health, and streams tokens to the React UI as they arrive.

## Prerequisites

- Go 1.25
- Node
- [Ollama](https://ollama.com/) running locally with `gemma3:latest` pulled

## Configuration

Optional environment variables (defaults shown):

| Variable | Default | Description |
|---|---|---|
| `OLLAMA_HOST` | `127.0.0.1:11434` | Ollama server host or URL |
| `OLLAMA_MODEL` | `gemma3:latest` | Model name |
| `OLLAMA_TEMPERATURE` | `0.8` | Sampling temperature |
| `OLLAMA_HEALTH_TIMEOUT` | `2s` | Health probe timeout |
| `OLLAMA_HEALTHCHECK_INTERVAL` | `5s` | UI status poll interval |

## Development

```bash
wails dev
```

## Build

```bash
wails build
```
