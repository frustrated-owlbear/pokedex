package tools

import (
	"context"
	"encoding/json"

	"github.com/frustrated-owlbear/pokedex/04-agentic-rag/internal/simulation"
)

type ClockTool struct {
	clock *simulation.Clock
}

func NewClockTool(clock *simulation.Clock) *ClockTool {
	return &ClockTool{clock: clock}
}

func (t *ClockTool) Name() string { return "clock" }

func (t *ClockTool) Description() string {
	return "Returns the current in-game time, period of day, and weather for the trainer's journey in Kanto."
}

func (t *ClockTool) Parameters() map[string]any {
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
}

func (t *ClockTool) Execute(_ context.Context, _ json.RawMessage) (string, error) {
	data, err := json.Marshal(t.clock.Snapshot())
	if err != nil {
		return "", err
	}
	return string(data), nil
}
