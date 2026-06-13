# 04-agentic-rag

Workshop stage 4: Agentic RAG Pokédex with a bounded tool-calling agent loop, pluggable tools, execution trace UI, and in-memory vector retrieval.

The Go backend uses Ollama (via langchaingo), native Ollama tool calling for the agent loop, in-memory SQLite for party Pokémon and session memory, and an embedded Kanto knowledge corpus for RAG.

## Prerequisites

- Go 1.25
- Node
- [Ollama](https://ollama.com/) running locally with a tool-capable chat model (e.g. `gemma3:latest`) and `nomic-embed-text` for embeddings

## Configuration

Optional environment variables (defaults shown):

| Variable | Default | Description |
|---|---|---|
| `OLLAMA_HOST` | `127.0.0.1:11434` | Ollama server host or URL |
| `OLLAMA_MODEL` | `gemma3:latest` | Chat model name |
| `AGENT_MODEL` | (empty) | Override model for tool-calling turns |
| `OLLAMA_EMBED_MODEL` | `nomic-embed-text` | Embedding model for RAG and session search |
| `AGENT_MAX_ITERATIONS` | `5` | Maximum agent loop iterations |
| `RAG_TOP_K` | `4` | Knowledge search result count |
| `OLLAMA_TEMPERATURE` | `0.8` | Sampling temperature |
| `OLLAMA_HEALTH_TIMEOUT` | `2s` | Health probe timeout |
| `OLLAMA_HEALTHCHECK_INTERVAL` | `5s` | UI status poll interval |

## Development

```bash
wails dev
```

The main screen shows the Agent Feed timeline, inference composer, current situation widgets, and team sidebar.

## Build

```bash
wails build
```
