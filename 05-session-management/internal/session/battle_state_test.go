package session

import (
	"context"
	"strings"
	"testing"

	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/domain"
)

func TestExtractBattleState_TrainingWithBrok(t *testing.T) {
	t.Parallel()

	state := ExtractBattleState("I'm going to train with Brok.", domain.BattleSessionState{}, nil)

	if state.Activity != "training" {
		t.Fatalf("expected training activity, got %q", state.Activity)
	}
	if state.OpponentTrainer != "Brock" {
		t.Fatalf("expected Brock, got %q", state.OpponentTrainer)
	}
}

func TestExtractBattleState_PronounChoseOnix(t *testing.T) {
	t.Parallel()

	prior := domain.BattleSessionState{
		Activity:        "training",
		OpponentTrainer: "Brock",
	}
	events := []domain.SessionEvent{
		{EventType: domain.EventUserMessage, Content: "I'm going to train with Brok."},
	}
	state := ExtractBattleState("He chose Onix!", prior, events)

	if state.OpponentTrainer != "Brock" {
		t.Fatalf("expected Brock, got %q", state.OpponentTrainer)
	}
	if state.OpponentPokemon != "Onix" {
		t.Fatalf("expected Onix, got %q", state.OpponentPokemon)
	}
}

func TestExtractBattleState_NamedChoseOnix(t *testing.T) {
	t.Parallel()

	state := ExtractBattleState("Brok chose Onix.", domain.BattleSessionState{}, nil)

	if state.OpponentTrainer != "Brock" {
		t.Fatalf("expected Brock, got %q", state.OpponentTrainer)
	}
	if state.OpponentPokemon != "Onix" {
		t.Fatalf("expected Onix, got %q", state.OpponentPokemon)
	}
}

func TestExtractBattleState_RecommendationRequest(t *testing.T) {
	t.Parallel()

	prior := domain.BattleSessionState{
		Activity:        "training",
		OpponentTrainer: "Brock",
		OpponentPokemon: "Onix",
	}
	state := ExtractBattleState("What pokemon should I chose?", prior, nil)

	if state.TrainerGoal == "" {
		t.Fatal("expected trainer goal for recommendation request")
	}
	if !IsPokemonRecommendationRequest(strings.ToLower("What pokemon should I chose?")) {
		t.Fatal("expected recommendation request detection")
	}
}

func TestRecommendPokemonAgainstOnix(t *testing.T) {
	t.Parallel()

	team := []domain.TeamPokemon{
		{Name: "Bulbasaur", PrimaryType: "GRASS", HP: 42, MaxHP: 42},
		{Name: "Squirtle", PrimaryType: "WATER", HP: 40, MaxHP: 40},
	}
	answer, ok := RecommendPokemon(team, "Onix")
	if !ok {
		t.Fatal("expected recommendation")
	}
	if !strings.Contains(answer, "Squirtle") {
		t.Fatalf("expected Squirtle against Onix, got %q", answer)
	}
	if strings.Contains(strings.ToLower(answer), "which pokémon will you choose") {
		t.Fatalf("recommendation should not ask for trainer choice, got %q", answer)
	}
}

func TestBattleStatePersistenceAfterOnix(t *testing.T) {
	t.Parallel()

	mgr := newTestManager(t)
	ctx := context.Background()
	sessionID, err := mgr.EnsureActiveSession(ctx)
	if err != nil {
		t.Fatalf("EnsureActiveSession: %v", err)
	}

	if _, err := mgr.UpdateBattleState(ctx, sessionID, "I'm going to train with Brok."); err != nil {
		t.Fatalf("turn 1 UpdateBattleState: %v", err)
	}
	state, err := mgr.UpdateBattleState(ctx, sessionID, "He chose Onix!")
	if err != nil {
		t.Fatalf("turn 2 UpdateBattleState: %v", err)
	}
	if state.OpponentPokemon != "Onix" {
		t.Fatalf("turn 2 returned %#v", state)
	}

	loaded, err := mgr.BattleStateForSession(sessionID)
	if err != nil {
		t.Fatalf("BattleStateForSession: %v", err)
	}
	if loaded.OpponentPokemon != "Onix" {
		t.Fatalf("loaded state %#v", loaded)
	}
}

func TestBattleConversationFlow(t *testing.T) {
	t.Parallel()

	mgr := newTestManager(t)
	ctx := context.Background()
	sessionID, err := mgr.EnsureActiveSession(ctx)
	if err != nil {
		t.Fatalf("EnsureActiveSession: %v", err)
	}

	turns := []struct {
		message string
		check   func(domain.BattleSessionState) error
	}{
		{
			message: "I'm going to train with Brok.",
			check: func(state domain.BattleSessionState) error {
				if state.Activity != "training" || state.OpponentTrainer != "Brock" {
					t.Fatalf("turn 1 state = %#v", state)
				}
				return nil
			},
		},
		{
			message: "He chose Onix!",
			check: func(state domain.BattleSessionState) error {
				if state.OpponentTrainer != "Brock" || state.OpponentPokemon != "Onix" {
					t.Fatalf("turn 2 state = %#v", state)
				}
				return nil
			},
		},
		{
			message: "What pokemon should I chose?",
			check: func(state domain.BattleSessionState) error {
				if state.OpponentPokemon != "Onix" {
					t.Fatalf("turn 3 state = %#v", state)
				}
				answer, ok, err := mgr.TryBattleRecommendation(ctx, sessionID, "What pokemon should I chose?")
				if err != nil {
					return err
				}
				if !ok {
					t.Fatal("expected battle recommendation")
				}
				if strings.Contains(strings.ToLower(answer), "which pokémon will you choose") ||
					strings.Contains(strings.ToLower(answer), "which pokemon will you choose") {
					t.Fatalf("unexpected clarification in recommendation: %q", answer)
				}
				if !strings.Contains(answer, "Squirtle") {
					t.Fatalf("expected Squirtle recommendation, got %q", answer)
				}
				return nil
			},
		},
	}

	for _, turn := range turns {
		if err := mgr.store.AddEvent(ctx, sessionID, domain.EventUserMessage, domain.ObservationNote, turn.message); err != nil {
			t.Fatalf("AddEvent: %v", err)
		}
		state, err := mgr.UpdateBattleState(ctx, sessionID, turn.message)
		if err != nil {
			t.Fatalf("UpdateBattleState(%q): %v", turn.message, err)
		}
		if err := turn.check(state); err != nil {
			t.Fatalf("check(%q): %v", turn.message, err)
		}
	}
}

func TestLoadBattleStateFromEvents(t *testing.T) {
	t.Parallel()

	payload := `{"activity":"training","opponent_trainer":"Brock","opponent_pokemon":"Onix"}`
	events := []domain.SessionEvent{
		{EventType: domain.EventStateUpdate, Category: domain.BattleStateCategory, Content: payload},
	}
	state := LoadBattleState(events)
	if state.OpponentTrainer != "Brock" || state.OpponentPokemon != "Onix" {
		t.Fatalf("unexpected loaded state %#v", state)
	}
}
