package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/domain"
	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/session"
)

type ObservationTool struct {
	store *session.Store
}

func NewObservationTool(store *session.Store) *ObservationTool {
	return &ObservationTool{store: store}
}

func (t *ObservationTool) Name() string { return "record_observation" }

func (t *ObservationTool) Description() string {
	return "Records an important gameplay observation for the current session. Use when the trainer reports captures, gym badges, battles, preferences, favorite Pokémon, locations visited, or notes worth remembering across sessions."
}

func (t *ObservationTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"category": map[string]any{
				"type":        "string",
				"description": "Observation category. Must be one of: capture, badge, battle, preference, favorite_pokemon, location, note",
				"enum": []string{
					"capture", "badge", "battle", "preference",
					"favorite_pokemon", "location", "note",
				},
			},
			"content": map[string]any{
				"type":        "string",
				"description": "What happened, in one or two sentences",
			},
		},
		"required": []string{"category", "content"},
	}
}

type observationArgs struct {
	Category string `json:"category"`
	Content  string `json:"content"`
}

func (t *ObservationTool) Execute(ctx context.Context, arguments json.RawMessage) (string, error) {
	var args observationArgs
	if err := json.Unmarshal(arguments, &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	sessionID, err := t.store.EnsureActiveSession(ctx)
	if err != nil {
		return "", err
	}
	category := domain.NormalizeObservationCategory(args.Category)
	if err := t.store.AddObservation(ctx, sessionID, category, strings.TrimSpace(args.Content)); err != nil {
		return "", err
	}
	return fmt.Sprintf("Recorded [%s] observation for session %s.", category, sessionID), nil
}
