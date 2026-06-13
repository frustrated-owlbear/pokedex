package agent

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/frustrated-owlbear/pokedex/04-agentic-rag/internal/llm"
	"github.com/frustrated-owlbear/pokedex/04-agentic-rag/internal/simulation"
	"github.com/frustrated-owlbear/pokedex/04-agentic-rag/internal/tools"
)

func TestLoopExecutesToolThenStreamsFinalAnswer(t *testing.T) {
	t.Parallel()

	var toolChatCalls int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/chat":
			toolChatCalls++
			if toolChatCalls == 1 {
				_, _ = w.Write([]byte(`{
					"message": {
						"role": "assistant",
						"content": "I should check location.",
						"tool_calls": [{
							"id": "call_gps",
							"type": "function",
							"function": {"name": "gps", "arguments": {}}
						}]
					},
					"done": true
				}`))
				return
			}
			_, _ = w.Write([]byte(`{
				"message": {"role": "assistant", "content": "You are in Viridian Forest."},
				"done": true
			}`))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer server.Close()

	streamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"message":{"role":"assistant","content":"You are in Viridian Forest."},"done":true}`))
	}))
	defer streamServer.Close()

	client, err := llm.NewClient(llm.Settings{
		ModelName:   "test-model",
		BaseURL:     server.URL,
		Temperature: 0.1,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	registry := NewRegistry(tools.NewGPSTool(simulation.NewGPS()))
	loop := NewLoop(client, registry, 3, nil)

	var steps []TraceStep
	var chunks strings.Builder
	err = loop.Run(
		context.Background(),
		Input{Prompt: "Where am I?"},
		func(step TraceStep) { steps = append(steps, step) },
		func(chunk string) { chunks.WriteString(chunk) },
	)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	if toolChatCalls == 0 {
		t.Fatalf("expected tool chat call")
	}
	if !strings.Contains(chunks.String(), "Viridian Forest") {
		t.Fatalf("expected final content, got %q", chunks.String())
	}

	foundAction := false
	for _, step := range steps {
		if step.Kind == StepAction {
			foundAction = true
		}
	}
	if !foundAction {
		t.Fatalf("expected action trace step, got %#v", steps)
	}
}

func TestBuildSituation(t *testing.T) {
	t.Parallel()

	situation := BuildSituation(
		`{"time":"18:32","period":"Evening","weather":"Clear"}`,
		`{"location":"Viridian Forest","region":"Kanto"}`,
		"Last session summary",
		[]string{"gps", "clock"},
	)
	if situation.Location != "Viridian Forest" || situation.Time != "18:32" {
		t.Fatalf("unexpected situation %#v", situation)
	}
}

func TestRegistryNames(t *testing.T) {
	t.Parallel()

	registry := NewRegistry(tools.NewClockTool(simulation.NewClock()))
	names := registry.Names()
	if len(names) != 1 || names[0] != "clock" {
		t.Fatalf("unexpected names %#v", names)
	}
	data, _ := json.Marshal(registry.Definitions())
	if len(data) == 0 {
		t.Fatalf("expected definitions json")
	}
}
