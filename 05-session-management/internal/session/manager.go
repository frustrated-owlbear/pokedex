package session

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/domain"
	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/llm"
	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/pokemonstore"
	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/simulation"
)

// Manager orchestrates autonomous session lifecycle and Pokémon state updates.
type Manager struct {
	store   *Store
	team    *pokemonstore.SQLiteStore
	clock   *simulation.Clock
	gps     *simulation.GPS
	client  *llm.Client
	decider *Decider
}

func NewManager(
	store *Store,
	team *pokemonstore.SQLiteStore,
	clock *simulation.Clock,
	gps *simulation.GPS,
	client *llm.Client,
) *Manager {
	return &Manager{
		store:   store,
		team:    team,
		clock:   clock,
		gps:     gps,
		client:  client,
		decider: NewDecider(client),
	}
}

func (m *Manager) Store() *Store {
	return m.store
}

func (m *Manager) EnsureActiveSession(ctx context.Context) (string, error) {
	return m.store.EnsureActiveSession(ctx)
}

func (m *Manager) LastSummary() string {
	return m.store.LastSummary()
}

func (m *Manager) SaveTurn(ctx context.Context, sessionID, userInput, finalAnswer, traceSummary string) error {
	return m.store.SaveTurn(ctx, sessionID, userInput, finalAnswer, traceSummary)
}

func (m *Manager) GetActiveSession(ctx context.Context) (string, []domain.SessionEvent, error) {
	sessionID, err := m.store.EnsureActiveSession(ctx)
	if err != nil {
		return "", nil, err
	}
	events, err := m.store.RecentEvents(sessionID, 20)
	return sessionID, events, err
}

func (m *Manager) AppendEvent(ctx context.Context, sessionID string, eventType domain.SessionEventType, category, content string) error {
	return m.store.AddEvent(ctx, sessionID, eventType, category, content)
}

func (m *Manager) CloseSession(ctx context.Context, sessionID string) (domain.Session, error) {
	log.Printf("[session] closing session %s", sessionID)
	return m.store.EndSession(ctx, sessionID)
}

func (m *Manager) StartNewSessionFromObservation(ctx context.Context, observation string) (string, []string, error) {
	var logs []string
	if m.store.ActiveSessionID() != "" {
		if _, err := m.CloseSession(ctx, m.store.ActiveSessionID()); err != nil {
			return "", logs, err
		}
		logs = append(logs, "Closed previous session")
	}
	sessionID, err := m.store.EnsureActiveSession(ctx)
	if err != nil {
		return "", logs, err
	}
	observation = strings.TrimSpace(observation)
	if observation != "" {
		if err := m.store.AddEvent(ctx, sessionID, domain.EventObservation, domain.ObservationLocation, observation); err != nil {
			return sessionID, logs, err
		}
		logs = append(logs, fmt.Sprintf("Started new session with observation: %s", observation))
	} else {
		logs = append(logs, "Started new session")
	}
	log.Printf("[session] new active session %s", sessionID)
	return sessionID, logs, nil
}

type ApplyResult struct {
	SessionID string
	Logs      []string
}

func (m *Manager) BuildDecisionContext(ctx context.Context, sessionID, userMessage, imageDescription string, hasImage bool) (DecisionInput, error) {
	events, err := m.store.RecentEvents(sessionID, 20)
	if err != nil {
		return DecisionInput{}, err
	}

	team, err := m.team.ListTeam()
	if err != nil {
		return DecisionInput{}, err
	}

	location := ""
	if m.gps != nil {
		snap := m.gps.Snapshot()
		location = snap.Location
	}

	battleState, err := m.store.LoadBattleState(sessionID)
	if err != nil {
		return DecisionInput{}, err
	}

	return DecisionInput{
		UserMessage:      userMessage,
		ImageDescription: imageDescription,
		HasImage:         hasImage,
		SessionID:        sessionID,
		RecentEvents:     events,
		Team:             team,
		Location:         location,
		LastEndedSummary: m.store.LastSummary(),
		BattleState:      battleState,
	}, nil
}

func (m *Manager) Decide(ctx context.Context, input DecisionInput) (domain.AgentSessionDecision, error) {
	return m.decider.Decide(ctx, input)
}

func (m *Manager) ApplyDecision(ctx context.Context, sessionID string, decision domain.AgentSessionDecision, userMessage string, hasImage bool) (ApplyResult, error) {
	result := ApplyResult{SessionID: sessionID}
	logDecision := func(msg string) {
		log.Printf("[session] %s", msg)
		result.Logs = append(result.Logs, msg)
	}

	if decision.NeedsClarification {
		if userMessage = strings.TrimSpace(userMessage); userMessage != "" {
			if err := m.store.AddEvent(ctx, sessionID, domain.EventUserMessage, domain.ObservationNote, userMessage); err != nil {
				return result, err
			}
		}
		logDecision(fmt.Sprintf("ask_clarification: %s", decision.Reason))
		return result, nil
	}

	if userMessage = strings.TrimSpace(userMessage); userMessage != "" {
		if err := m.store.AddEvent(ctx, sessionID, domain.EventUserMessage, domain.ObservationNote, userMessage); err != nil {
			return result, err
		}
	}

	if decision.PokemonName != "" && decision.PokemonHealth != nil {
		if err := m.team.UpdateHPByName(decision.PokemonName, *decision.PokemonHealth); err != nil {
			return result, err
		}
		stateContent := fmt.Sprintf("%s health updated to %d.", decision.PokemonName, *decision.PokemonHealth)
		if err := m.store.AddEvent(ctx, sessionID, domain.EventStateUpdate, domain.ObservationNote, stateContent); err != nil {
			return result, err
		}
		logDecision(fmt.Sprintf("State Update: %s HP → %d", decision.PokemonName, *decision.PokemonHealth))
	}

	if observation := strings.TrimSpace(decision.Observation); observation != "" {
		eventType := domain.EventObservation
		category := observationCategory(decision, hasImage)
		if hasImage && userMessage == "" {
			eventType = domain.EventImageObservation
		}
		if err := m.store.AddEvent(ctx, sessionID, eventType, category, observation); err != nil {
			return result, err
		}
		logDecision(fmt.Sprintf("Observation: %s", observation))
	}

	if decision.ShouldCloseSession || decision.ShouldCompact {
		if _, err := m.CloseSession(ctx, sessionID); err != nil {
			return result, err
		}
		logDecision(fmt.Sprintf("Close Session: %s", decision.Reason))
		sessionID = ""
	}

	if decision.ShouldStartNew {
		var obs string
		if strings.TrimSpace(decision.Observation) != "" && (decision.ShouldCloseSession || decision.ShouldCompact) {
			obs = decision.Observation
		}
		newID, startLogs, err := m.startNewSession(ctx, obs)
		if err != nil {
			return result, err
		}
		sessionID = newID
		for _, line := range startLogs {
			logDecision(line)
		}
	}

	if sessionID == "" {
		var err error
		sessionID, err = m.store.EnsureActiveSession(ctx)
		if err != nil {
			return result, err
		}
	}

	action := decision.Action
	if action == "" {
		action = domain.ActionContinueSession
	}
	logDecision(fmt.Sprintf("Session Decision: %s — %s", action, decision.Reason))

	result.SessionID = sessionID
	return result, nil
}

func (m *Manager) startNewSession(ctx context.Context, observation string) (string, []string, error) {
	m.store.activeID = ""
	return m.StartNewSessionFromObservation(ctx, observation)
}

func observationCategory(decision domain.AgentSessionDecision, hasImage bool) string {
	lower := strings.ToLower(decision.Observation + " " + decision.Reason)
	if strings.Contains(lower, "battle") || strings.Contains(lower, "gym") {
		return domain.ObservationBattle
	}
	if hasImage || strings.Contains(lower, "forest") || strings.Contains(lower, "route") || strings.Contains(lower, "location") {
		return domain.ObservationLocation
	}
	return domain.ObservationNote
}

func (m *Manager) SessionContextPrompt(sessionID string) (string, error) {
	summary, events, err := m.store.ActiveSessionContext(sessionID)
	if err != nil {
		return "", err
	}
	state, err := m.store.LoadBattleState(sessionID)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString(summary)
	if summary := state.PromptSummary(); summary != "" {
		b.WriteString("\n\nStructured session state:\n")
		b.WriteString(summary)
	}
	if len(events) > 0 {
		b.WriteString("\n\nRecent session events:\n")
		start := len(events) - 8
		if start < 0 {
			start = 0
		}
		for _, event := range events[start:] {
			fmt.Fprintf(&b, "- [%s] %s\n", event.EventType, event.Content)
		}
	}
	return strings.TrimSpace(b.String()), nil
}
