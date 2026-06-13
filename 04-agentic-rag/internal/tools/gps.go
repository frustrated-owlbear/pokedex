package tools

import (
	"context"
	"encoding/json"

	"github.com/frustrated-owlbear/pokedex/04-agentic-rag/internal/simulation"
)

type GPSTool struct {
	gps *simulation.GPS
}

func NewGPSTool(gps *simulation.GPS) *GPSTool {
	return &GPSTool{gps: gps}
}

func (t *GPSTool) Name() string { return "gps" }

func (t *GPSTool) Description() string {
	return "Returns the trainer's current simulated location in the Kanto region."
}

func (t *GPSTool) Parameters() map[string]any {
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
}

func (t *GPSTool) Execute(_ context.Context, _ json.RawMessage) (string, error) {
	data, err := json.Marshal(t.gps.Snapshot())
	if err != nil {
		return "", err
	}
	return string(data), nil
}
