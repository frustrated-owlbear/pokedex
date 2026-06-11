package main

import (
	"context"
	"errors"
	"time"

	"github.com/frustrated-owlbear/pokedex/03-rag/internal/config"
	"github.com/frustrated-owlbear/pokedex/03-rag/internal/domain"
	"github.com/frustrated-owlbear/pokedex/03-rag/internal/llm"
	"github.com/frustrated-owlbear/pokedex/03-rag/internal/pokemonstore"
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
	teamStore         *pokemonstore.SQLiteStore
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

	store, err := pokemonstore.NewSQLiteStore()
	if err != nil {
		return nil, err
	}

	return &App{
		cfg:       cfg,
		llm:       client,
		teamStore: store,
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
	if a.teamStore != nil {
		if err := a.teamStore.Close(); err != nil {
			runtime.LogError(a.ctx, err.Error())
		}
	}
}

// ListMyTeam returns the trainer's current party from the in-memory SQLite store.
func (a *App) ListMyTeam() ([]domain.TeamPokemon, error) {
	if a.teamStore == nil {
		return nil, errors.New("team store unavailable")
	}

	team, err := a.teamStore.ListTeam()
	if err != nil {
		runtime.LogError(a.ctx, err.Error())
		return nil, err
	}

	return team, nil
}

// GetTeamMember returns one party Pokémon by id.
func (a *App) GetTeamMember(id int) (domain.TeamPokemon, error) {
	if a.teamStore == nil {
		return domain.TeamPokemon{}, errors.New("team store unavailable")
	}

	pokemon, err := a.teamStore.GetTeamMember(id)
	if err != nil {
		runtime.LogError(a.ctx, err.Error())
		return domain.TeamPokemon{}, err
	}

	return pokemon, nil
}

// AskPokedex streams an LLM reply for a user prompt; chunks emit as "llm:chunk".
// imageBase64 and imageMIME are optional; pass empty strings when no image is attached.
func (a *App) AskPokedex(prompt, imageBase64, imageMIME string) error {
	team, err := a.ListMyTeam()
	if err != nil {
		runtime.LogError(a.ctx, err.Error())
		team = nil
	}

	messages, err := llm.BuildMessages(prompt, imageBase64, imageMIME, team)
	if err != nil {
		return err
	}
	return a.llm.StreamChat(a.ctx, messages, func(chunk string) {
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
