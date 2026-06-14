package main

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/agent"
	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/config"
	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/domain"
	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/llm"
	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/pokemonstore"
	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/rag"
	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/session"
	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/simulation"
	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/tools"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	ollamaStatusEvent  = "ollama:status"
	agentTraceEvent    = "agent:trace"
	agentTraceResetEvent = "agent:trace-reset"
)

type SystemStatus struct {
	Label   string `json:"label"`
	Value   string `json:"value"`
	Detail  string `json:"detail,omitempty"`
	Healthy bool   `json:"healthy"`
}

type SituationView struct {
	Location string   `json:"location"`
	Region   string   `json:"region"`
	Time     string   `json:"time"`
	Period   string   `json:"period"`
	Weather  string   `json:"weather"`
	Memory   string   `json:"memory"`
	Tools    []string `json:"tools"`
	Analysis string   `json:"analysis,omitempty"`
}

// App struct
type App struct {
	ctx               context.Context
	cfg               config.Config
	llm               *llm.Client
	teamStore         *pokemonstore.SQLiteStore
	clock             *simulation.Clock
	gps               *simulation.GPS
	sessionStore      *session.Store
	retriever         *rag.Retriever
	agentLoop         *agent.Loop
	registry          *agent.Registry
	traceSteps        []agent.TraceStep
	healthcheckCancel context.CancelFunc
}

// NewApp creates a new App application struct
func NewApp(cfg config.Config) (*App, error) {
	client, err := llm.NewClient(llm.Settings{
		ModelName:      cfg.OllamaModel,
		AgentModelName: cfg.AgentModelName(),
		BaseURL:        cfg.OllamaBaseURL(),
		Temperature:    cfg.OllamaTemperature,
		HealthTimeout:  cfg.OllamaHealthTimeout,
	})
	if err != nil {
		return nil, err
	}

	store, err := pokemonstore.NewSQLiteStore()
	if err != nil {
		return nil, err
	}

	embedder, err := rag.NewEmbedder(cfg.OllamaBaseURL(), cfg.OllamaEmbedModel)
	if err != nil {
		return nil, err
	}

	retriever := rag.NewRetriever(embedder, rag.NewStore(), cfg.RAGTopK)
	if err := retriever.Bootstrap(context.Background()); err != nil {
		log.Printf("rag bootstrap failed: %v", err)
	}

	sessionStore, err := session.NewStore(embedder, session.NewLLMSummarizer(client))
	if err != nil {
		return nil, err
	}

	clock := simulation.NewClock()
	gps := simulation.NewGPS()

	sessionManager := session.NewManager(sessionStore, store, clock, gps, client)

	registry := agent.NewDefaultRegistry(agent.Dependencies{
		TeamStore: store,
		Clock:     clock,
		GPS:       gps,
		Sessions:  sessionStore,
		Retriever: retriever,
		RAGTopK:   cfg.RAGTopK,
	})

	agentLoop := agent.NewLoop(client, registry, cfg.AgentMaxIterations, sessionManager)

	return &App{
		cfg:          cfg,
		llm:          client,
		teamStore:    store,
		clock:        clock,
		gps:          gps,
		sessionStore: sessionStore,
		retriever:    retriever,
		agentLoop:    agentLoop,
		registry:     registry,
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
	if a.sessionStore != nil {
		if err := a.sessionStore.EndActiveSessionIfNeeded(a.ctx); err != nil {
			runtime.LogError(a.ctx, err.Error())
		}
		if err := a.sessionStore.Close(); err != nil {
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

// GetCurrentSituation returns GPS, clock, memory, and tool metadata for the UI.
func (a *App) GetCurrentSituation() SituationView {
	clockTool := tools.NewClockTool(a.clock)
	gpsTool := tools.NewGPSTool(a.gps)

	clockJSON, _ := clockTool.Execute(a.ctx, nil)
	gpsJSON, _ := gpsTool.Execute(a.ctx, nil)

	memory := "No previous sessions recorded yet."
	if a.sessionStore != nil {
		memory = a.sessionStore.LastSummary()
	}

	toolsList := []string{}
	if a.registry != nil {
		toolsList = a.registry.Names()
	}

	view := SituationView{
		Location: a.gps.Snapshot().Location,
		Region:   a.gps.Snapshot().Region,
		Time:     a.clock.Snapshot().Time,
		Period:   a.clock.Snapshot().Period,
		Weather:  a.clock.Snapshot().Weather,
		Memory:   memory,
		Tools:    toolsList,
	}

	_ = clockJSON
	_ = gpsJSON
	return view
}

// ListRecentTraces returns trace steps from the latest agent run.
func (a *App) ListRecentTraces() []agent.TraceStep {
	return append([]agent.TraceStep(nil), a.traceSteps...)
}

// ListSessions returns gameplay sessions for the UI.
func (a *App) ListSessions() ([]session.SessionView, error) {
	if a.sessionStore == nil {
		return nil, errors.New("session store unavailable")
	}
	return a.sessionStore.ListSessions()
}

// EndGameplaySession summarizes the active session and starts a new one.
func (a *App) EndGameplaySession() (session.SessionView, error) {
	if a.sessionStore == nil {
		return session.SessionView{}, errors.New("session store unavailable")
	}

	sessionID, err := a.sessionStore.EnsureActiveSession(a.ctx)
	if err != nil {
		return session.SessionView{}, err
	}

	ended, err := a.sessionStore.EndSession(a.ctx, sessionID)
	if err != nil {
		runtime.LogError(a.ctx, err.Error())
		return session.SessionView{}, err
	}

	if _, err := a.sessionStore.EnsureActiveSession(a.ctx); err != nil {
		runtime.LogError(a.ctx, err.Error())
	}

	return toSessionView(ended), nil
}

func toSessionView(ended domain.Session) session.SessionView {
	view := session.SessionView{
		ID:         ended.ID,
		StartedAt:  ended.StartTime.Format(time.RFC3339),
		Summary:    ended.Summary,
		EventCount: len(ended.Events),
	}
	if ended.EndTime != nil {
		view.EndedAt = ended.EndTime.Format(time.RFC3339)
	} else {
		view.Active = true
	}
	return view
}

// AskPokedex runs the agent loop; trace steps emit as "agent:trace" and the final answer as "llm:chunk".
func (a *App) AskPokedex(prompt, imageBase64, imageMIME string) error {
	return a.agentLoop.Run(
		a.ctx,
		agent.Input{
			Prompt:         prompt,
			ImageBase64:    imageBase64,
			ImageMIME:      imageMIME,
			OnSessionReset: a.resetAgentTrace,
		},
		func(step agent.TraceStep) {
			a.traceSteps = append(a.traceSteps, step)
			runtime.EventsEmit(a.ctx, agentTraceEvent, step)
		},
		func(chunk string) {
			runtime.EventsEmit(a.ctx, "llm:chunk", chunk)
		},
	)
}

func (a *App) resetAgentTrace() {
	a.traceSteps = nil
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, agentTraceResetEvent, nil)
	}
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

		a.emitSystemStatus(pollCtx)

		for {
			select {
			case <-pollCtx.Done():
				return
			case <-ticker.C:
				a.emitSystemStatus(pollCtx)
			}
		}
	}()
}

func (a *App) emitSystemStatus(ctx context.Context) {
	if a.ctx == nil {
		return
	}
	for _, status := range a.currentSystemStatuses(ctx) {
		runtime.EventsEmit(a.ctx, ollamaStatusEvent, status)
	}
}

func (a *App) currentSystemStatuses(ctx context.Context) []SystemStatus {
	statuses := []SystemStatus{a.currentOllamaStatus(ctx)}

	vectorStatus := SystemStatus{
		Label:   "Vector DB",
		Value:   "Disconnected",
		Detail:  "Knowledge index unavailable",
		Healthy: false,
	}
	if a.retriever != nil && a.retriever.Ready() {
		vectorStatus.Value = "Connected"
		vectorStatus.Detail = "Kanto knowledge corpus indexed"
		vectorStatus.Healthy = true
	}
	statuses = append(statuses, vectorStatus)

	memoryStatus := SystemStatus{
		Label:   "Memory",
		Value:   "Unavailable",
		Detail:  "Session store unavailable",
		Healthy: false,
	}
	if a.sessionStore != nil && a.sessionStore.Ready() {
		memoryStatus.Value = "OK"
		memoryStatus.Detail = "Session memory ready"
		memoryStatus.Healthy = true
	}
	statuses = append(statuses, memoryStatus)

	return statuses
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
