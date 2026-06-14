package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tmc/langchaingo/llms"
)

func TestGenerateWithToolsRoundTrip(t *testing.T) {
	t.Parallel()

	var callCount int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if r.URL.Path != "/api/chat" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}

		var req ollamaChatRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Fatalf("decode request: %v", err)
		}
		if req.Stream {
			t.Fatalf("expected non-streaming request")
		}
		if len(req.Tools) == 0 {
			t.Fatalf("expected tools in request")
		}

		w.Header().Set("Content-Type", "application/json")
		if callCount == 1 {
			_, _ = w.Write([]byte(`{
				"message": {
					"role": "assistant",
					"content": "Checking location.",
					"tool_calls": [{
						"id": "call_1",
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
	}))
	defer server.Close()

	client, err := NewClient(Settings{
		ModelName:     "test-model",
		BaseURL:       server.URL,
		Temperature:   0.2,
		HealthTimeout: 0,
	})
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	tools := []llms.Tool{{
		Type: "function",
		Function: &llms.FunctionDefinition{
			Name:        "gps",
			Description: "location",
			Parameters:  map[string]any{"type": "object", "properties": map[string]any{}},
		},
	}}

	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "system"),
		llms.TextParts(llms.ChatMessageTypeHuman, "Where am I?"),
	}

	resp, err := client.GenerateWithTools(context.Background(), messages, tools)
	if err != nil {
		t.Fatalf("GenerateWithTools: %v", err)
	}
	if len(resp.Choices[0].ToolCalls) != 1 {
		t.Fatalf("expected one tool call, got %#v", resp.Choices[0].ToolCalls)
	}
	if resp.Choices[0].ToolCalls[0].FunctionCall.Name != "gps" {
		t.Fatalf("unexpected tool name %q", resp.Choices[0].ToolCalls[0].FunctionCall.Name)
	}

	messages = AppendAssistantTurn(messages, resp.Choices[0])
	messages = AppendToolResults(messages, resp.Choices[0].ToolCalls, map[string]string{
		"call_1": `{"location":"Viridian Forest"}`,
	})

	resp, err = client.GenerateWithTools(context.Background(), messages, tools)
	if err != nil {
		t.Fatalf("second GenerateWithTools: %v", err)
	}
	if !strings.Contains(resp.Choices[0].Content, "Viridian Forest") {
		t.Fatalf("unexpected content %q", resp.Choices[0].Content)
	}
}

func TestAppendToolResults(t *testing.T) {
	t.Parallel()

	calls := []llms.ToolCall{{
		ID:   "call_1",
		Type: "function",
		FunctionCall: &llms.FunctionCall{
			Name:      "clock",
			Arguments: `{}`,
		},
	}}
	messages := AppendToolResults(nil, calls, map[string]string{"call_1": "18:32"})
	if len(messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(messages))
	}
	part, ok := messages[0].Parts[0].(llms.ToolCallResponse)
	if !ok {
		t.Fatalf("expected ToolCallResponse part")
	}
	if part.Content != "18:32" {
		t.Fatalf("unexpected content %q", part.Content)
	}
}
