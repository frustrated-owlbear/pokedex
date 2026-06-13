package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/frustrated-owlbear/pokedex/04-agentic-rag/internal/pokemonstore"
	"github.com/frustrated-owlbear/pokedex/04-agentic-rag/internal/simulation"
)

func TestClockToolExecute(t *testing.T) {
	t.Parallel()

	tool := NewClockTool(simulation.NewClock())
	out, err := tool.Execute(context.Background(), json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if out == "" {
		t.Fatalf("expected clock snapshot")
	}
}

func TestGPSToolExecute(t *testing.T) {
	t.Parallel()

	tool := NewGPSTool(simulation.NewGPS())
	out, err := tool.Execute(context.Background(), nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !contains(out, "Viridian Forest") {
		t.Fatalf("unexpected output %q", out)
	}
}

func TestPokemonDBToolExecute(t *testing.T) {
	t.Parallel()

	store, err := pokemonstore.NewSQLiteStore()
	if err != nil {
		t.Fatalf("NewSQLiteStore: %v", err)
	}
	defer store.Close()

	tool := NewPokemonDBTool(store)
	out, err := tool.Execute(context.Background(), nil)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !contains(out, "Bulbasaur") {
		t.Fatalf("expected seeded team in output %q", out)
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
