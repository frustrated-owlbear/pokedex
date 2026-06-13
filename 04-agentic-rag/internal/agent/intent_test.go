package agent

import (
	"strings"
	"testing"
)

func TestToolHintForQuestionTeam(t *testing.T) {
	t.Parallel()

	hint := toolHintForQuestion("What is my first pokemon?")
	if hint == "" {
		t.Fatalf("expected team routing hint")
	}
	if !strings.Contains(hint, "pokemon_db") || !strings.Contains(hint, "session_memory") {
		t.Fatalf("hint should mention pokemon_db and session_memory, got %q", hint)
	}
}

func TestToolHintForQuestionMemory(t *testing.T) {
	t.Parallel()

	hint := toolHintForQuestion("What did we talk about last time?")
	if hint == "" {
		t.Fatalf("expected memory routing hint")
	}
}

func TestToolHintForQuestionFirstCaught(t *testing.T) {
	t.Parallel()

	hint := toolHintForQuestion("What was the first Pokemon I caught?")
	if !strings.Contains(hint, "sort_by") || !strings.Contains(hint, "caught_date") {
		t.Fatalf("expected caught-date routing hint, got %q", hint)
	}
}

func TestPokemonDBToolArgsForQuestion(t *testing.T) {
	t.Parallel()

	args := pokemonDBToolArgsForQuestion("What was my first caught pokemon?")
	if !strings.Contains(args, `"sort_by":"caught_date"`) || !strings.Contains(args, `"sort_order":"asc"`) {
		t.Fatalf("unexpected args %q", args)
	}

	args = pokemonDBToolArgsForQuestion("Who is my lead pokemon?")
	if !strings.Contains(args, `"sort_by":"slot"`) {
		t.Fatalf("expected slot sort args, got %q", args)
	}
}

func TestMentionsTrainerTeam(t *testing.T) {
	t.Parallel()

	if !mentionsTrainerTeam("what is my first pokemon?") {
		t.Fatal("expected team match")
	}
}
