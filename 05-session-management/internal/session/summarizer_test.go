package session

import (
	"context"
	"testing"

	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/domain"
)

func TestFallbackSummarizeJoinsObservations(t *testing.T) {
	t.Parallel()

	summary := FallbackSummarize([]domain.Observation{
		{Category: domain.ObservationBadge, Content: "Earned Thunder Badge"},
		{Category: domain.ObservationPreference, Content: "Leads with Bulbasaur"},
	})
	if summary == "" {
		t.Fatal("expected non-empty summary")
	}
}

func TestLLMSummarizerEmptyObservations(t *testing.T) {
	t.Parallel()

	summarizer := &LLMSummarizer{}
	summary, err := summarizer.Summarize(context.Background(), nil)
	if err != nil {
		t.Fatalf("Summarize: %v", err)
	}
	if summary != "" {
		t.Fatalf("expected empty summary, got %q", summary)
	}
}
