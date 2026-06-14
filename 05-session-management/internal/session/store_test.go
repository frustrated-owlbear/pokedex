package session

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/domain"
)

type stubEmbedder struct{}

func (stubEmbedder) Embed(_ context.Context, text string) ([]float32, error) {
	text = strings.ToLower(text)
	vec := make([]float32, 8)
	for i, ch := range text {
		vec[i%8] += float32(ch)
	}
	return vec, nil
}

func (stubEmbedder) ModelName() string { return "stub" }

type stubSummarizer struct{}

func (stubSummarizer) Summarize(_ context.Context, observations []domain.Observation) (string, error) {
	return FallbackSummarize(observations), nil
}

func newTestStore(t *testing.T) *Store {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	store, err := NewStoreWithDSN(dsn, stubEmbedder{}, stubSummarizer{})
	if err != nil {
		t.Fatalf("NewStoreWithDSN: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func TestStoreSeedAndSearch(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	results, err := store.Search(ctx, "Electric Bulbasaur", 3)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 {
		t.Fatalf("expected seeded search results")
	}

	found := false
	for _, result := range results {
		lower := strings.ToLower(result.Content)
		if strings.Contains(lower, "electric") || strings.Contains(lower, "thunder") || strings.Contains(lower, "bulbasaur") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected electric-related memory, got %#v", results)
	}

	summary := store.LastSummary()
	if summary == "" || summary == "No previous sessions recorded yet." {
		t.Fatalf("expected seeded summary, got %q", summary)
	}
}

func TestStoreObservationAndEndSession(t *testing.T) {
	store := newTestStore(t)
	ctx := context.Background()

	sessionID, err := store.EnsureActiveSession(ctx)
	if err != nil {
		t.Fatalf("EnsureActiveSession: %v", err)
	}

	if err := store.AddObservation(ctx, sessionID, domain.ObservationBattle, "Defeated Misty's Staryu with Pikachu"); err != nil {
		t.Fatalf("AddObservation: %v", err)
	}

	ended, err := store.EndSession(ctx, sessionID)
	if err != nil {
		t.Fatalf("EndSession: %v", err)
	}
	if ended.Summary == "" {
		t.Fatalf("expected session summary")
	}
	if ended.EndTime == nil {
		t.Fatalf("expected ended session")
	}
}

func TestStoreListSessions(t *testing.T) {
	store := newTestStore(t)
	sessions, err := store.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(sessions) < 2 {
		t.Fatalf("expected at least 2 seeded sessions, got %d", len(sessions))
	}
}

func TestFormatSearchResultsEmpty(t *testing.T) {
	t.Parallel()

	if FormatSearchResults(nil) == "" {
		t.Fatalf("expected fallback text")
	}
}

func TestFallbackSummarize(t *testing.T) {
	t.Parallel()

	summary := FallbackSummarize([]domain.Observation{
		{Category: domain.ObservationBattle, Content: "Used Bulbasaur against Electric-types"},
	})
	if !strings.Contains(summary, "Bulbasaur") {
		t.Fatalf("unexpected summary %q", summary)
	}
}
