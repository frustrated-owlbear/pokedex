package llm

import (
	"strings"
	"testing"

	"github.com/tmc/langchaingo/llms"
)

func TestPrepareStreamMessagesFlattensToolResults(t *testing.T) {
	t.Parallel()

	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, "system"),
		llms.TextParts(llms.ChatMessageTypeHuman, "Where am I?"),
		{
			Role: llms.ChatMessageTypeAI,
			Parts: []llms.ContentPart{
				llms.TextContent{Text: "Checking GPS."},
				llms.ToolCall{
					ID:   "call_1",
					Type: "function",
					FunctionCall: &llms.FunctionCall{
						Name:      "gps",
						Arguments: `{}`,
					},
				},
			},
		},
		{
			Role: llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{
				llms.ToolCallResponse{
					ToolCallID: "call_1",
					Name:       "gps",
					Content:    `{"location":"Viridian Forest"}`,
				},
			},
		},
	}

	flattened := PrepareStreamMessages(messages)
	if len(flattened) < 3 {
		t.Fatalf("expected flattened messages, got %d", len(flattened))
	}
	last := flattened[len(flattened)-1]
	text := textFromParts(last.Parts)
	if !strings.Contains(text, "Viridian Forest") {
		t.Fatalf("expected tool notes in stream prompt, got %q", text)
	}
}
