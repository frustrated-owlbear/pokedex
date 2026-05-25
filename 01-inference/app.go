package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	ollamaStatusEvent      = "ollama:status"
	ollamaHealthcheckEvery = 5 * time.Second
)

type SystemStatus struct {
	Label   string `json:"label"`
	Value   string `json:"value"`
	Detail  string `json:"detail,omitempty"`
	Healthy bool   `json:"healthy"`
}

// App struct
type App struct {
	ctx               context.Context
	healthcheckCancel context.CancelFunc
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.startOllamaHealthcheck()
}

func (a *App) shutdown(_ context.Context) {
	if a.healthcheckCancel != nil {
		a.healthcheckCancel()
	}
}

// AskPokedex streams an LLM reply for a user prompt; chunks emit as "llm:chunk".
func (a *App) AskPokedex(prompt string) error {
	p := strings.TrimSpace(prompt)
	if p == "" {
		return fmt.Errorf("prompt is required")
	}
	return streamCompletion(a.ctx, promptQuestion(p), func(chunk string) {
		runtime.EventsEmit(a.ctx, "llm:chunk", chunk)
	})
}

func (a *App) startOllamaHealthcheck() {
	pollCtx, cancel := context.WithCancel(context.Background())
	a.healthcheckCancel = cancel

	go func() {
		ticker := time.NewTicker(ollamaHealthcheckEvery)
		defer ticker.Stop()

		a.emitOllamaStatus(pollCtx)

		for {
			select {
			case <-pollCtx.Done():
				return
			case <-ticker.C:
				a.emitOllamaStatus(pollCtx)
			}
		}
	}()
}

func (a *App) emitOllamaStatus(ctx context.Context) {
	if a.ctx == nil {
		return
	}
	runtime.EventsEmit(a.ctx, ollamaStatusEvent, currentOllamaStatus(ctx))
}

func currentOllamaStatus(ctx context.Context) SystemStatus {
	probe := checkOllamaHealth(ctx)
	status := SystemStatus{
		Label:   "Ollama",
		Value:   "Unavailable",
		Detail:  probe.Detail,
		Healthy: false,
	}

	switch {
	case probe.Reachable && probe.ModelAvailable:
		status.Value = "Running"
		status.Detail = "gemma3:latest available"
		status.Healthy = true
	case probe.Reachable:
		status.Value = "Model missing"
	default:
		if status.Detail == "" {
			status.Detail = "Ollama not reachable"
		}
	}

	return status
}
