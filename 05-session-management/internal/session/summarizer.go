package session

import (
	"context"
	"fmt"
	"strings"

	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/domain"
	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/llm"
)

type Summarizer interface {
	Summarize(ctx context.Context, observations []domain.Observation) (string, error)
}

type LLMSummarizer struct {
	client *llm.Client
}

func NewLLMSummarizer(client *llm.Client) *LLMSummarizer {
	return &LLMSummarizer{client: client}
}

const summarizeSystemPrompt = `You summarize a Pokémon trainer's gameplay session for long-term memory.
Write 2-4 concise sentences in Kanto Season 1 Pokédex tone.

Preserve:
- important trainer goals
- Pokémon state changes (fainted, injured, healed)
- location changes
- observations that may matter later
- unresolved tasks
- battle advice already given

Do not preserve:
- irrelevant chat
- repeated assistant wording
- low-value intermediate reasoning

Do not invent facts not present in the observations.`

func (s *LLMSummarizer) Summarize(ctx context.Context, observations []domain.Observation) (string, error) {
	if len(observations) == 0 {
		return "", nil
	}
	prompt := formatObservationsForPrompt(observations)
	summary, err := s.client.Complete(ctx, summarizeSystemPrompt, prompt)
	if err != nil {
		return FallbackSummarize(observations), nil
	}
	return strings.TrimSpace(summary), nil
}

// FallbackSummarize builds a deterministic summary without an LLM.
func FallbackSummarize(observations []domain.Observation) string {
	if len(observations) == 0 {
		return ""
	}
	parts := make([]string, 0, len(observations))
	for _, obs := range observations {
		parts = append(parts, fmt.Sprintf("[%s] %s", obs.Category, obs.Content))
	}
	return strings.Join(parts, " ")
}

func formatObservationsForPrompt(observations []domain.Observation) string {
	var b strings.Builder
	b.WriteString("Session observations:\n")
	for i, obs := range observations {
		fmt.Fprintf(&b, "%d. [%s] %s\n", i+1, obs.Category, obs.Content)
	}
	b.WriteString("\nSummarize the important events from this session.")
	return b.String()
}
