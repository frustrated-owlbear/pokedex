package llm

import (
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
)

// PrepareStreamMessages converts an agent conversation with tool calls into
// text-only messages compatible with the stock langchaingo Ollama StreamChat path.
func PrepareStreamMessages(messages []llms.MessageContent) []llms.MessageContent {
	out := make([]llms.MessageContent, 0, len(messages)+1)
	var toolNotes strings.Builder

	for _, mc := range messages {
		switch mc.Role {
		case llms.ChatMessageTypeSystem, llms.ChatMessageTypeHuman:
			out = append(out, flattenMessage(mc))
		case llms.ChatMessageTypeAI:
			text := textFromParts(mc.Parts)
			if text != "" {
				out = append(out, llms.TextParts(llms.ChatMessageTypeAI, text))
			}
			for _, part := range mc.Parts {
				if call, ok := part.(llms.ToolCall); ok && call.FunctionCall != nil {
					fmt.Fprintf(&toolNotes, "Tool call %s(%s)\n", call.FunctionCall.Name, call.FunctionCall.Arguments)
				}
			}
		case llms.ChatMessageTypeTool:
			for _, part := range mc.Parts {
				if resp, ok := part.(llms.ToolCallResponse); ok {
					fmt.Fprintf(&toolNotes, "Tool result %s: %s\n", resp.Name, resp.Content)
				}
			}
		}
	}

	if toolNotes.Len() > 0 {
		out = append(out, llms.TextParts(
			llms.ChatMessageTypeHuman,
			"Tool observations gathered so far:\n"+toolNotes.String()+"\nProvide your final Pokédex answer now.",
		))
	}

	return out
}

func flattenMessage(mc llms.MessageContent) llms.MessageContent {
	parts := make([]llms.ContentPart, 0, len(mc.Parts))
	for _, part := range mc.Parts {
		switch p := part.(type) {
		case llms.TextContent:
			parts = append(parts, p)
		case llms.BinaryContent:
			parts = append(parts, p)
		}
	}
	return llms.MessageContent{Role: mc.Role, Parts: parts}
}

func textFromParts(parts []llms.ContentPart) string {
	var b strings.Builder
	for _, part := range parts {
		if text, ok := part.(llms.TextContent); ok {
			b.WriteString(text.Text)
		}
	}
	return strings.TrimSpace(b.String())
}
