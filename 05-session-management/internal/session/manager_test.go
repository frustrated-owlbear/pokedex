package session

import (
	"context"
	"testing"

	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/domain"
	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/pokemonstore"
	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/simulation"
)

func newTestManager(t *testing.T) *Manager {
	t.Helper()
	store := newTestStore(t)
	team, err := pokemonstore.NewSQLiteStore()
	if err != nil {
		t.Fatalf("NewSQLiteStore: %v", err)
	}
	t.Cleanup(func() { _ = team.Close() })
	return NewManager(store, team, simulation.NewClock(), simulation.NewGPS(), nil)
}

func TestManagerApplyDecisionUpdatesBulbasaurHP(t *testing.T) {
	t.Parallel()

	mgr := newTestManager(t)
	ctx := context.Background()

	sessionID, err := mgr.EnsureActiveSession(ctx)
	if err != nil {
		t.Fatalf("EnsureActiveSession: %v", err)
	}

	zero := 0
	_, err = mgr.ApplyDecision(ctx, sessionID, domain.AgentSessionDecision{
		Action:        domain.ActionUpdatePokemon,
		Reason:        "Bulbasaur fainted",
		PokemonName:   "Bulbasaur",
		PokemonHealth: &zero,
	}, "Bulbasaur health updated to 0", false)
	if err != nil {
		t.Fatalf("ApplyDecision: %v", err)
	}

	pokemon, err := mgr.team.GetByName("Bulbasaur")
	if err != nil {
		t.Fatalf("GetByName: %v", err)
	}
	if pokemon.HP != 0 {
		t.Fatalf("expected Bulbasaur HP 0, got %d", pokemon.HP)
	}
}

func TestManagerApplyDecisionUpdatesPikachuHP(t *testing.T) {
	t.Parallel()

	mgr := newTestManager(t)
	ctx := context.Background()

	sessionID, err := mgr.EnsureActiveSession(ctx)
	if err != nil {
		t.Fatalf("EnsureActiveSession: %v", err)
	}

	zero := 0
	result, err := mgr.ApplyDecision(ctx, sessionID, domain.AgentSessionDecision{
		Action:        domain.ActionUpdatePokemon,
		Reason:        "Pikachu fainted",
		PokemonName:   "Pikachu",
		PokemonHealth: &zero,
		Observation:   "Pikachu fainted. Health updated to 0.",
	}, "Pikachu fainted.", false)
	if err != nil {
		t.Fatalf("ApplyDecision: %v", err)
	}
	if result.SessionID == "" {
		t.Fatal("expected session id")
	}

	pokemon, err := mgr.team.GetByName("Pikachu")
	if err != nil {
		t.Fatalf("GetByName: %v", err)
	}
	if pokemon.HP != 0 {
		t.Fatalf("expected HP 0, got %d", pokemon.HP)
	}

	events, err := mgr.store.RecentEvents(sessionID, 10)
	if err != nil {
		t.Fatalf("RecentEvents: %v", err)
	}
	foundState := false
	for _, event := range events {
		if event.EventType == domain.EventStateUpdate {
			foundState = true
			break
		}
	}
	if !foundState {
		t.Fatalf("expected state_update event, got %#v", events)
	}
}

func TestManagerCloseAndStartNewSession(t *testing.T) {
	t.Parallel()

	mgr := newTestManager(t)
	ctx := context.Background()

	sessionID, err := mgr.EnsureActiveSession(ctx)
	if err != nil {
		t.Fatalf("EnsureActiveSession: %v", err)
	}
	if err := mgr.store.AddObservation(ctx, sessionID, domain.ObservationBattle, "Battle advice against Geodude"); err != nil {
		t.Fatalf("AddObservation: %v", err)
	}

	result, err := mgr.ApplyDecision(ctx, sessionID, domain.AgentSessionDecision{
		Action:             domain.ActionStartNewSession,
		Reason:             "Location shift",
		Observation:        "Trainer appears to be in a forest area.",
		ShouldCloseSession: true,
		ShouldCompact:      true,
		ShouldStartNew:     true,
	}, "", true)
	if err != nil {
		t.Fatalf("ApplyDecision: %v", err)
	}
	if result.SessionID == sessionID {
		t.Fatalf("expected new session id, got same %q", sessionID)
	}

	sessions, err := mgr.store.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	var ended, active int
	for _, view := range sessions {
		if view.Active {
			active++
		} else if view.ID == sessionID {
			ended++
			if view.Summary == "" {
				t.Fatal("expected compacted summary on ended session")
			}
		}
	}
	if ended != 1 || active < 1 {
		t.Fatalf("expected one ended and one active session, got %#v", sessions)
	}
}

func TestStoreSaveTurnRecordsAssistantMessage(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	sessionID, err := store.EnsureActiveSession(ctx)
	if err != nil {
		t.Fatalf("EnsureActiveSession: %v", err)
	}
	if err := store.SaveTurn(ctx, sessionID, "hello", "Hi trainer", ""); err != nil {
		t.Fatalf("SaveTurn: %v", err)
	}

	events, err := store.RecentEvents(sessionID, 10)
	if err != nil {
		t.Fatalf("RecentEvents: %v", err)
	}
	if len(events) != 1 || events[0].EventType != domain.EventAssistantMessage {
		t.Fatalf("expected assistant message event, got %#v", events)
	}
}

func TestStoreAddEventPersistsEventType(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	sessionID, err := store.EnsureActiveSession(ctx)
	if err != nil {
		t.Fatalf("EnsureActiveSession: %v", err)
	}
	if err := store.AddEvent(ctx, sessionID, domain.EventStateUpdate, domain.ObservationNote, "Pikachu HP set to 0"); err != nil {
		t.Fatalf("AddEvent: %v", err)
	}

	events, err := store.listEvents(sessionID)
	if err != nil {
		t.Fatalf("listEvents: %v", err)
	}
	if len(events) != 1 || events[0].EventType != domain.EventStateUpdate {
		t.Fatalf("unexpected events %#v", events)
	}
}
