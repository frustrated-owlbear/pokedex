package agent

import (
	"context"
	"encoding/json"
	"testing"
)

type stubTool struct {
	name   string
	result string
	called bool
}

func (s *stubTool) Name() string { return s.name }

func (s *stubTool) Description() string { return "stub" }

func (s *stubTool) Parameters() map[string]any {
	return map[string]any{"type": "object", "properties": map[string]any{}}
}

func (s *stubTool) Execute(_ context.Context, _ json.RawMessage) (string, error) {
	s.called = true
	return s.result, nil
}

func TestRegistryDefinitionsAndExecute(t *testing.T) {
	t.Parallel()

	tool := &stubTool{name: "clock", result: `{"time":"18:32"}`}
	registry := NewRegistry(tool)

	defs := registry.Definitions()
	if len(defs) != 1 || defs[0].Function.Name != "clock" {
		t.Fatalf("unexpected definitions %#v", defs)
	}

	out, err := registry.Execute(context.Background(), "clock", `{}`)
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}
	if out != tool.result {
		t.Fatalf("unexpected output %q", out)
	}
	if !tool.called {
		t.Fatalf("expected tool to be called")
	}
}

func TestNewTraceStep(t *testing.T) {
	t.Parallel()

	step := NewTraceStep(StepThought, "Analyze Situation", "Checking context")
	if step.Kind != StepThought {
		t.Fatalf("unexpected kind %q", step.Kind)
	}
	if step.Timestamp == "" {
		t.Fatalf("expected timestamp")
	}
}
