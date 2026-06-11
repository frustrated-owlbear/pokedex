# 03-rag

Workshop stage 3: Wails desktop app with Ollama inference and an in-memory SQLite database for listing party Pokémon.

The Go backend loads configuration from environment variables, probes Ollama health, streams LLM tokens to the React UI, and serves team data from SQLite (`github.com/mattn/go-sqlite3`) running inside the Wails process. On startup the app creates the `team_pokemon` table in memory (with `caught_date` and optional `birthday`) and seeds Bulbasaur and Pidgey.

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

The main screen shows the existing inference UI plus a **MY TEAM** widget in the right sidebar, loaded from the in-memory SQLite store.

## Build

```bash
wails build
```
