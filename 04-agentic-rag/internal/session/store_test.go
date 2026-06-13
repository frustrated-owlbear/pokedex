package session

import (
	"context"
	"testing"
)

type stubEmbedder struct{}

func (stubEmbedder) Embed(_ context.Context, text string) ([]float32, error) {
	if text == "" {
		return []float32{0, 0, 0}, nil
	}
	return []float32{float32(len(text)), 0.1, 0.2}, nil
}

func (stubEmbedder) ModelName() string { return "stub" }

func TestStoreSaveAndSearch(t *testing.T) {
	t.Parallel()

	store, err := NewStore(stubEmbedder{})
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	if err := store.SaveTurn(ctx, "default", "I saw Pikachu", "Likely Pikachu near the trees.", "thought: checked gps"); err != nil {
		t.Fatalf("SaveTurn: %v", err)
	}

	results, err := store.Search(ctx, "Pikachu forest", 2)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 {
		t.Fatalf("expected search results")
	}
	if store.LastSummary() == "" {
		t.Fatalf("expected last summary")
	}
}

func TestFormatSearchResultsEmpty(t *testing.T) {
	t.Parallel()

	if FormatSearchResults(nil) == "" {
		t.Fatalf("expected fallback text")
	}
}
