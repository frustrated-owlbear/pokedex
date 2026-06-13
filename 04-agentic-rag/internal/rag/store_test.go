package rag

import "testing"

func TestCosineSimilarityIdenticalVectors(t *testing.T) {
	t.Parallel()

	vector := []float32{1, 0, 0}
	score := CosineSimilarity(vector, vector)
	if score < 0.99 {
		t.Fatalf("expected ~1.0, got %v", score)
	}
}

func TestStoreSearchOrdersByScore(t *testing.T) {
	t.Parallel()

	store := NewStore()
	store.Add(Document{ID: "a", Vector: []float32{1, 0, 0}})
	store.Add(Document{ID: "b", Vector: []float32{0.9, 0.1, 0}})
	store.Add(Document{ID: "c", Vector: []float32{0, 1, 0}})

	results := store.Search([]float32{1, 0, 0}, 2)
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].ID != "a" {
		t.Fatalf("expected first result a, got %s", results[0].ID)
	}
}

func TestFormatResultsEmpty(t *testing.T) {
	t.Parallel()

	if FormatResults(nil) == "" {
		t.Fatalf("expected non-empty fallback text")
	}
}
