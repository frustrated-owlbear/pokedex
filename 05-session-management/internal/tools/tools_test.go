package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/domain"
	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/pokemonstore"
	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/session"
	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/simulation"
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
	if !contains(out, "Squirtle") {
		t.Fatalf("expected most recently caught Squirtle, got %q", out)
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

func TestSessionMemoryToolExecute(t *testing.T) {
	store, err := session.NewStoreWithDSN(
		"file:test-memory-tool?mode=memory&cache=shared",
		stubSessionEmbedder{},
		testSummarizer{},
	)
	if err != nil {
		t.Fatalf("NewStoreWithDSN: %v", err)
	}
	defer store.Close()

	tool := NewSessionMemoryTool(store, 3)
	out, err := tool.Execute(context.Background(), json.RawMessage(`{"query":"Electric opponents Bulbasaur"}`))
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if contains(out, "No matching past session memories") {
		t.Fatalf("expected seeded memory in output %q", out)
	}
	if !contains(out, "session") {
		t.Fatalf("expected session memory in output %q", out)
	}
}

func TestObservationToolExecute(t *testing.T) {
	store, err := session.NewStoreWithDSN(
		"file:test-obs-tool?mode=memory&cache=shared",
		stubSessionEmbedder{},
		testSummarizer{},
	)
	if err != nil {
		t.Fatalf("NewStoreWithDSN: %v", err)
	}
	defer store.Close()

	tool := NewObservationTool(store)
	out, err := tool.Execute(context.Background(), json.RawMessage(`{"category":"battle","content":"Defeated Brock with Vine Whip"}`))
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !contains(out, "Recorded") {
		t.Fatalf("unexpected output %q", out)
	}
}

func TestObservationToolNormalizesInvalidCategory(t *testing.T) {
	store, err := session.NewStoreWithDSN(
		"file:test-obs-normalize?mode=memory&cache=shared",
		stubSessionEmbedder{},
		testSummarizer{},
	)
	if err != nil {
		t.Fatalf("NewStoreWithDSN: %v", err)
	}
	defer store.Close()

	tool := NewObservationTool(store)
	out, err := tool.Execute(context.Background(), json.RawMessage(`{"category":"trainer goal","content":"Beat Brock at Pewter Gym"}`))
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if !contains(out, "Recorded [note]") {
		t.Fatalf("expected normalized category in output %q", out)
	}
}

type stubSessionEmbedder struct{}

func (stubSessionEmbedder) Embed(_ context.Context, text string) ([]float32, error) {
	text = strings.ToLower(text)
	vec := make([]float32, 8)
	for i, ch := range text {
		vec[i%8] += float32(ch)
	}
	return vec, nil
}

func (stubSessionEmbedder) ModelName() string { return "stub" }

type testSummarizer struct{}

func (testSummarizer) Summarize(_ context.Context, observations []domain.Observation) (string, error) {
	return session.FallbackSummarize(observations), nil
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
