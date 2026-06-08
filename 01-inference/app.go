package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/frustrated-owlbear/pokedex/01-inference/internal/config"
	"github.com/frustrated-owlbear/pokedex/01-inference/internal/llm"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const ollamaStatusEvent = "ollama:status"

type SystemStatus struct {
	Label   string `json:"label"`
	Value   string `json:"value"`
	Detail  string `json:"detail,omitempty"`
	Healthy bool   `json:"healthy"`
}

// App struct
type App struct {
	ctx               context.Context
	cfg               config.Config
	llm               *llm.Client
	healthcheckCancel context.CancelFunc
}

// NewApp creates a new App application struct
func NewApp(cfg config.Config) (*App, error) {
	client, err := llm.NewClient(llm.Settings{
		ModelName:     cfg.OllamaModel,
		BaseURL:       cfg.OllamaBaseURL(),
		Temperature:   cfg.OllamaTemperature,
		HealthTimeout: cfg.OllamaHealthTimeout,
	})
	if err != nil {
		return nil, err
	}

	return &App{
		cfg: cfg,
		llm: client,
	}, nil
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
	return a.llm.StreamCompletion(a.ctx, llm.Prompt(p), func(chunk string) {
		runtime.EventsEmit(a.ctx, "llm:chunk", chunk)
	})
}

func (a *App) startOllamaHealthcheck() {
	pollCtx, cancel := context.WithCancel(context.Background())
	a.healthcheckCancel = cancel

	interval := a.cfg.OllamaHealthcheckEvery
	if interval <= 0 {
		interval = 5 * time.Second
	}

	go func() {
		ticker := time.NewTicker(interval)
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
	runtime.EventsEmit(a.ctx, ollamaStatusEvent, a.currentOllamaStatus(ctx))
}

func (a *App) currentOllamaStatus(ctx context.Context) SystemStatus {
	probe := a.llm.CheckHealth(ctx)
	status := SystemStatus{
		Label:   "Ollama",
		Value:   "Unavailable",
		Detail:  probe.Detail,
		Healthy: false,
	}

	switch {
	case probe.Reachable && probe.ModelAvailable:
		status.Value = "Running"
		status.Detail = a.llm.ModelName() + " available"
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
