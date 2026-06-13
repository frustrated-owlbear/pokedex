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

func TestPokemonDBToolLimitOffset(t *testing.T) {
	t.Parallel()

	store, err := pokemonstore.NewSQLiteStore()
	if err != nil {
		t.Fatalf("NewSQLiteStore: %v", err)
	}
	defer store.Close()

	tool := NewPokemonDBTool(store)
	out, err := tool.Execute(context.Background(), json.RawMessage(`{"sort_by":"slot","limit":"1","offset":"0"}`))
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if contains(out, "Pidgey") {
		t.Fatalf("expected only first party member, got %q", out)
	}
	if !contains(out, "Bulbasaur") {
		t.Fatalf("expected Bulbasaur in output %q", out)
	}
	if !contains(out, `"count":1`) {
		t.Fatalf("expected count metadata, got %q", out)
	}
}

func TestPokemonDBToolSortByCaughtDate(t *testing.T) {
	t.Parallel()

	store, err := pokemonstore.NewSQLiteStore()
	if err != nil {
		t.Fatalf("NewSQLiteStore: %v", err)
	}
	defer store.Close()

	tool := NewPokemonDBTool(store)
	out, err := tool.Execute(context.Background(), json.RawMessage(`{"sort_by":"caught_date","sort_order":"desc","limit":1}`))
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !contains(out, "Pidgey") {
		t.Fatalf("expected most recently caught Pidgey, got %q", out)
	}
	if contains(out, "Bulbasaur") {
		t.Fatalf("expected Bulbasaur to be excluded, got %q", out)
	}
}

func TestPokemonDBToolFilterByPrimaryType(t *testing.T) {
	t.Parallel()

	store, err := pokemonstore.NewSQLiteStore()
	if err != nil {
		t.Fatalf("NewSQLiteStore: %v", err)
	}
	defer store.Close()

	tool := NewPokemonDBTool(store)
	out, err := tool.Execute(context.Background(), json.RawMessage(`{"primary_type":"grass"}`))
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !contains(out, "Bulbasaur") {
		t.Fatalf("expected grass-type Bulbasaur, got %q", out)
	}
	if contains(out, "Pidgey") {
		t.Fatalf("expected Pidgey to be filtered out, got %q", out)
	}
	if !contains(out, `"count":1`) {
		t.Fatalf("expected one match, got %q", out)
	}
}

func TestPokemonDBToolFilterByLevel(t *testing.T) {
	t.Parallel()

	store, err := pokemonstore.NewSQLiteStore()
	if err != nil {
		t.Fatalf("NewSQLiteStore: %v", err)
	}
	defer store.Close()

	tool := NewPokemonDBTool(store)
	out, err := tool.Execute(context.Background(), json.RawMessage(`{"min_level":12,"max_level":12}`))
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !contains(out, "Pidgey") {
		t.Fatalf("expected level-12 Pidgey, got %q", out)
	}
	if contains(out, "Bulbasaur") {
		t.Fatalf("expected Bulbasaur to be filtered out, got %q", out)
	}
}

func TestPokemonDBToolFilterByName(t *testing.T) {
	t.Parallel()

	store, err := pokemonstore.NewSQLiteStore()
	if err != nil {
		t.Fatalf("NewSQLiteStore: %v", err)
	}
	defer store.Close()

	tool := NewPokemonDBTool(store)
	out, err := tool.Execute(context.Background(), json.RawMessage(`{"name":"pidge"}`))
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !contains(out, "Pidgey") {
		t.Fatalf("expected Pidgey by partial name, got %q", out)
	}
	if contains(out, "Bulbasaur") {
		t.Fatalf("expected Bulbasaur to be filtered out, got %q", out)
	}
}

func TestPokemonDBToolFilterByCaughtDate(t *testing.T) {
	t.Parallel()

	store, err := pokemonstore.NewSQLiteStore()
	if err != nil {
		t.Fatalf("NewSQLiteStore: %v", err)
	}
	defer store.Close()

	tool := NewPokemonDBTool(store)
	out, err := tool.Execute(context.Background(), json.RawMessage(`{"caught_date":"2024-03-01"}`))
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !contains(out, "Pidgey") {
		t.Fatalf("expected Pidgey caught on 2024-03-01, got %q", out)
	}
	if contains(out, "Bulbasaur") {
		t.Fatalf("expected Bulbasaur to be filtered out, got %q", out)
	}
}

func TestPokemonDBToolFilterByCaughtDateRange(t *testing.T) {
	t.Parallel()

	store, err := pokemonstore.NewSQLiteStore()
	if err != nil {
		t.Fatalf("NewSQLiteStore: %v", err)
	}
	defer store.Close()

	tool := NewPokemonDBTool(store)
	out, err := tool.Execute(context.Background(), json.RawMessage(`{"caught_before":"2024-02-28"}`))
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !contains(out, "Bulbasaur") {
		t.Fatalf("expected Bulbasaur caught before March, got %q", out)
	}
	if contains(out, "Pidgey") {
		t.Fatalf("expected Pidgey to be filtered out, got %q", out)
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
