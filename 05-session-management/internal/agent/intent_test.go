package agent

import (
	"strings"
	"testing"
)

func TestMentionsPokemonRecommendation(t *testing.T) {
	t.Parallel()

	cases := []struct {
		question string
		want     bool
	}{
		{"What pokemon should I chose?", true},
		{"Which Pokémon should I choose?", true},
		{"Brok chose Onix.", false},
		{"He chose Onix!", false},
	}

	for _, tc := range cases {
		if got := mentionsPokemonRecommendation(tc.question); got != tc.want {
			t.Fatalf("mentionsPokemonRecommendation(%q) = %v, want %v", tc.question, got, tc.want)
		}
	}
}

func TestBattleToolHintForQuestion(t *testing.T) {
	t.Parallel()

	ctx := "Structured session state:\nOpponent Pokémon: Onix"
	hint := battleToolHintForQuestion("What pokemon should I choose?", ctx)
	if hint == "" {
		t.Fatal("expected battle tool hint")
	}
	if hint == battleToolHintForQuestion("Where am I?", ctx) {
		t.Fatal("expected empty hint for unrelated question")
	}
}

func TestHealthAdviceHintOnlyForFreshOrHealthQuestions(t *testing.T) {
	t.Parallel()

	staleCtx := "Recent session events:\n- [observation] Bulbasaur fainted. Health updated to 0."
	if hint := healthAdviceHint("I decided to train with Brock.", staleCtx); hint != "" {
		t.Fatalf("expected no stale health hint, got %q", hint)
	}

	freshCtx := staleCtx + "\n\nRecent state change: Bulbasaur fainted (HP 0). Advise visiting a Pokémon Center or switching Pokémon."
	if hint := healthAdviceHint("Got it.", freshCtx); hint == "" {
		t.Fatal("expected health hint for fresh faint state")
	}
	if hint := healthAdviceHint("Bulbasaur fainted.", staleCtx); hint == "" {
		t.Fatal("expected health hint for health-related question")
	}
}

func TestFilterSessionContextHidesStaleHealthForUnrelatedQuestions(t *testing.T) {
	t.Parallel()

	ctx := strings.Join([]string{
		"Current session events:",
		"1. [user_message] Brock chose Onix.",
		"2. [state_update] Bulbasaur health updated to 0.",
		"3. [observation] Training with Brock.",
		"",
		"Structured session state:",
		"Opponent trainer: Brock",
		"Opponent Pokémon: Onix",
	}, "\n")

	filtered := filterSessionContextForQuestion("Brock chose Onix as his battle Pokémon.", ctx)
	if strings.Contains(filtered, "Bulbasaur health updated") {
		t.Fatalf("expected stale health filtered out, got %q", filtered)
	}
	if !strings.Contains(filtered, "Opponent Pokémon: Onix") {
		t.Fatalf("expected battle state preserved, got %q", filtered)
	}

	if !strings.Contains(filterSessionContextForQuestion("Bulbasaur fainted.", ctx), "Bulbasaur health updated") {
		t.Fatal("expected health context kept for health-related question")
	}
}
