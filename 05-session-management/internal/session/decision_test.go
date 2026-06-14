package session

import (
	"strings"
	"testing"

	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/domain"
)

func demoTeam() []domain.TeamPokemon {
	return []domain.TeamPokemon{
		{Name: "Bulbasaur", HP: 42, MaxHP: 42},
		{Name: "Pikachu", HP: 35, MaxHP: 35},
		{Name: "Squirtle", HP: 40, MaxHP: 40},
	}
}

func TestDecideHeuristic_BulbasaurHealthUpdated(t *testing.T) {
	t.Parallel()

	decision, ok := TryDecideHealthUpdate(DecisionInput{
		UserMessage: "Bulbasaur health updated to 0",
		Team:        demoTeam(),
	})
	if !ok {
		t.Fatal("expected health update decision")
	}
	if decision.PokemonName != "Bulbasaur" {
		t.Fatalf("expected Bulbasaur, got %q", decision.PokemonName)
	}
	if decision.PokemonHealth == nil || *decision.PokemonHealth != 0 {
		t.Fatalf("expected HP 0, got %#v", decision.PokemonHealth)
	}
}

func TestDecideHeuristic_PikachuFainted(t *testing.T) {
	t.Parallel()

	decision := DecideHeuristic(DecisionInput{
		UserMessage: "Pikachu fainted.",
		Team:        demoTeam(),
	})

	if decision.Action != domain.ActionUpdatePokemon {
		t.Fatalf("expected update_pokemon_state, got %q", decision.Action)
	}
	if decision.PokemonName != "Pikachu" {
		t.Fatalf("expected Pikachu, got %q", decision.PokemonName)
	}
	if decision.PokemonHealth == nil || *decision.PokemonHealth != 0 {
		t.Fatalf("expected HP 0, got %#v", decision.PokemonHealth)
	}
}

func TestDecideHeuristic_ImplicitPokemon(t *testing.T) {
	t.Parallel()

	events := []domain.SessionEvent{
		{EventType: domain.EventUserMessage, Content: "Pikachu looks tired."},
	}
	decision := DecideHeuristic(DecisionInput{
		UserMessage:  "My Pokémon fainted.",
		RecentEvents: events,
		Team:         demoTeam(),
	})

	if decision.PokemonName != "Pikachu" {
		t.Fatalf("expected inferred Pikachu, got %q", decision.PokemonName)
	}
	if decision.PokemonHealth == nil || *decision.PokemonHealth != 0 {
		t.Fatalf("expected HP 0, got %#v", decision.PokemonHealth)
	}
}

func TestDecideHeuristic_ImageTopicShift(t *testing.T) {
	t.Parallel()

	events := []domain.SessionEvent{
		{EventType: domain.EventObservation, Category: domain.ObservationBattle, Content: "Discussing battle strategy against Geodude"},
	}
	decision := DecideHeuristic(DecisionInput{
		HasImage:         true,
		ImageDescription: "A forest path with tall trees",
		RecentEvents:     events,
		Team:             demoTeam(),
	})

	if !decision.ShouldCloseSession || !decision.ShouldStartNew {
		t.Fatalf("expected close and new session, got %#v", decision)
	}
	if !strings.Contains(strings.ToLower(decision.Observation), "forest") &&
		!strings.Contains(strings.ToLower(decision.Observation), "trainer") {
		t.Fatalf("expected location observation, got %q", decision.Observation)
	}
}

func TestDecideHeuristic_AmbiguousFaint(t *testing.T) {
	t.Parallel()

	events := []domain.SessionEvent{
		{EventType: domain.EventUserMessage, Content: "Bulbasaur and Pikachu are both low on HP."},
	}
	decision := DecideHeuristic(DecisionInput{
		UserMessage:  "My Pokémon fainted.",
		RecentEvents: events,
		Team:         demoTeam(),
	})

	if !decision.NeedsClarification {
		t.Fatalf("expected clarification, got %#v", decision)
	}
	if decision.PokemonHealth != nil {
		t.Fatalf("expected no HP update, got %#v", decision.PokemonHealth)
	}
}

func TestInferPokemonName(t *testing.T) {
	t.Parallel()

	name, ambiguous := InferPokemonName("Pikachu looks hurt", nil, demoTeam(), domain.BattleSessionState{})
	if name != "Pikachu" || ambiguous {
		t.Fatalf("expected Pikachu unambiguous, got %q ambiguous=%v", name, ambiguous)
	}

	events := []domain.SessionEvent{{Content: "Pikachu looks tired."}}
	name, ambiguous = InferPokemonName("My pokemon fainted", events, demoTeam(), domain.BattleSessionState{})
	if name != "Pikachu" || ambiguous {
		t.Fatalf("expected inferred Pikachu, got %q ambiguous=%v", name, ambiguous)
	}

	events = []domain.SessionEvent{{Content: "Bulbasaur and Pikachu are ready."}}
	_, ambiguous = InferPokemonName("My pokemon fainted", events, demoTeam(), domain.BattleSessionState{})
	if !ambiguous {
		t.Fatal("expected ambiguous inference")
	}
}

func TestInferPokemonName_RecommendedSquirtleAfterOnix(t *testing.T) {
	t.Parallel()

	team := demoTeam()
	team[0].HP = 0

	events := []domain.SessionEvent{
		{EventType: domain.EventUserMessage, Content: "Brock chose Onix."},
		{EventType: domain.EventAssistantMessage, Content: "Use Squirtle — Water moves are strong against Onix."},
	}
	battleState := domain.BattleSessionState{
		Activity:        "training",
		OpponentTrainer: "Brock",
		OpponentPokemon: "Onix",
		ActivePokemon:   "Squirtle",
	}

	name, ambiguous := InferPokemonName("My pokemon fainted.", events, team, battleState)
	if ambiguous {
		t.Fatal("expected unambiguous Squirtle inference")
	}
	if name != "Squirtle" {
		t.Fatalf("expected Squirtle, got %q", name)
	}

	decision, ok := TryDecideHealthUpdate(DecisionInput{
		UserMessage:  "My pokemon fainted.",
		RecentEvents: events,
		Team:         team,
		BattleState:  battleState,
	})
	if !ok {
		t.Fatal("expected health update decision")
	}
	if decision.NeedsClarification {
		t.Fatalf("expected no clarification, got %#v", decision)
	}
	if decision.PokemonName != "Squirtle" {
		t.Fatalf("expected Squirtle HP update, got %#v", decision)
	}
}

func TestParseDecisionJSON(t *testing.T) {
	t.Parallel()

	raw := "```json\n{\"action\":\"continue_session\",\"reason\":\"ongoing battle\",\"observation\":\"Trainer noted injury\"}\n```"
	decision, err := ParseDecisionJSON(raw)
	if err != nil {
		t.Fatalf("ParseDecisionJSON: %v", err)
	}
	if decision.Action != domain.ActionContinueSession {
		t.Fatalf("unexpected action %q", decision.Action)
	}
	if decision.Observation == "" {
		t.Fatal("expected observation")
	}

	if _, err := ParseDecisionJSON("not json"); err == nil {
		t.Fatal("expected parse error")
	}
}
