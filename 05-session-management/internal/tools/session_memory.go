package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/session"
)

type SessionMemoryTool struct {
	store *session.Store
	topK  int
}

func NewSessionMemoryTool(store *session.Store, topK int) *SessionMemoryTool {
	if topK <= 0 {
		topK = 3
	}
	return &SessionMemoryTool{store: store, topK: topK}
}

func (t *SessionMemoryTool) Name() string { return "session_memory" }

func (t *SessionMemoryTool) Description() string {
	return "Searches past gameplay session summaries and observations. Use when the trainer asks about previous sessions, past battles, earned badges, locations visited, trainer habits, or preferences from earlier play. Do not use for current team, party, or owned Pokémon questions; use pokemon_db instead."
}

func (t *SessionMemoryTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"query": map[string]any{
				"type":        "string",
				"description": "Search query describing what to recall from past gameplay sessions",
			},
		},
		"required": []string{"query"},
	}
}

type sessionMemoryArgs struct {
	Query string `json:"query"`
}

func (t *SessionMemoryTool) Execute(ctx context.Context, arguments json.RawMessage) (string, error) {
	var args sessionMemoryArgs
	if err := json.Unmarshal(arguments, &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}
	results, err := t.store.Search(ctx, args.Query, t.topK)
	if err != nil {
		return "", err
	}
	return session.FormatSearchResults(results), nil
}
