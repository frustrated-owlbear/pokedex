# Pokédex Workshop

A multi-stage workshop for building a local-first Pokédex desktop app with [Wails](https://wails.io/), Go, React, and [Ollama](https://ollama.com/). Each stage lives in its own folder (`01-inference` through `05-session-management`); see that stage's README for run and configuration details.

## Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| [Go](https://go.dev/) | 1.25+ | Backend and Wails bindings |
| [Node.js](https://nodejs.org/) | 18+ (LTS recommended) | Frontend build (Vite + React) |
| [Wails](https://wails.io/) | v2 | Desktop app framework |
| [Ollama](https://ollama.com/) | latest | Local LLM inference |

---

## Go

Install Go **1.25 or newer** (each stage's `go.mod` pins the exact toolchain).

### macOS

```bash
brew install go
```

Or download the installer from [go.dev/dl](https://go.dev/dl/).

### Linux

```bash
# Debian/Ubuntu
sudo apt update && sudo apt install golang-go

# Fedora
sudo dnf install golang
```

Or use the official tarball from [go.dev/dl](https://go.dev/dl/) and add `GOROOT` / `PATH` as described in the install notes.

### Windows

Download and run the MSI from [go.dev/dl](https://go.dev/dl/), or:

```powershell
winget install GoLang.Go
```

### Verify

```bash
go version   # should report go1.25 or newer
```

---

## Node.js

Node is used to build each stage's React frontend. **Node 18 LTS or newer** is recommended.

### macOS

```bash
brew install node
```

Or use a version manager:

```bash
# fnm
brew install fnm && fnm install --lts

# nvm
curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.1/install.sh | bash
nvm install --lts
```

### Linux

```bash
# Debian/Ubuntu (NodeSource LTS)
curl -fsSL https://deb.nodesource.com/setup_lts.x | sudo -E bash -
sudo apt install -y nodejs
```

Or install [fnm](https://github.com/Schniz/fnm) / [nvm](https://github.com/nvm-sh/nvm) and run `fnm install --lts` or `nvm install --lts`.

### Windows

```powershell
winget install OpenJS.NodeJS.LTS
```

Or install [fnm](https://github.com/Schniz/fnm) / [nvm-windows](https://github.com/coreybutler/nvm-windows).

### Verify

```bash
node --version   # v18.x or newer
npm --version
```

---

## Wails

[Wails v2](https://wails.io/docs/gettingstarted/installation) wraps the Go backend and web frontend in a native desktop window.

### Platform build tools

Install these **before** installing the Wails CLI.

**macOS**

```bash
xcode-select --install   # Xcode Command Line Tools (if not already installed)
```

**Linux (Debian/Ubuntu example)**

```bash
sudo apt install build-essential libgtk-3-dev libwebkit2gtk-4.1-dev
```

**Windows**

- [WebView2](https://developer.microsoft.com/en-us/microsoft-edge/webview2/) (usually pre-installed on Windows 11)
- [Build Tools for Visual Studio](https://visualstudio.microsoft.com/downloads/) with the **Desktop development with C++** workload

### Install the CLI

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
```

Ensure `$GOPATH/bin` (or `$HOME/go/bin`) is on your `PATH`.

### Verify

```bash
wails doctor
```

Fix anything `wails doctor` flags before running a stage.

---

## Ollama

Ollama serves the local models used by every workshop stage.

### Install

- **macOS / Linux / Windows:** [ollama.com/download](https://ollama.com/download)

```bash
# macOS (Homebrew)
brew install ollama
```

```powershell
# Windows
winget install Ollama.Ollama
```

### Start the server

The Ollama app starts the server automatically on macOS and Windows. On Linux you may need:

```bash
ollama serve
```

Default API: `http://127.0.0.1:11434`

### Pull required models

At minimum, stage 1 needs a chat model:

```bash
ollama pull gemma3:latest
```

Stages 3–5 also use embeddings for RAG and session search:

```bash
ollama pull nomic-embed-text
```

Later stages that use tool calling need a model that supports tools (e.g. `llama3.1:8b`).

Run app with a specific model by providing env variable

```bash
OLLAMA_MODEL=llama3.1:latest wails dev
```

### Verify

```bash
ollama list
curl http://127.0.0.1:11434/api/tags
```

---

## Quick start

Once everything is installed:

```bash
cd 01-inference    # or any stage folder
wails dev
```

For build instructions, environment variables, and stage-specific setup, open that stage's `README.md`.
