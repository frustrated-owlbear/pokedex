// Package llm provides Ollama chat helpers for the Pokédex agent.
//
// Tool-calling turns use GenerateWithTools in toolchat.go instead of the
// stock langchaingo Ollama provider. See that file for rationale.
package llm

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/tmc/langchaingo/llms"
)

/*
GenerateWithTools posts directly to Ollama's /api/chat endpoint with native
tool definitions and parses tool_calls from the response.

Why not langchaingo's ollama.LLM?

langchaingo v0.1.14's Ollama provider only serializes TextContent and
BinaryContent message parts and never sends a "tools" field on ChatRequest.
Attempting tool turns through GenerateContent therefore cannot invoke Ollama
function calling. Upstream work to fix this lives in:

	https://github.com/tmc/langchaingo/pull/1491

Once that PR is released in a langchaingo version we depend on, migrate
GenerateWithTools to delegate to llms.WithTools on the stock provider and
delete the duplicated HTTP types in this file. Final-answer streaming should
continue to use StreamChat regardless.

References:
  - Ollama tool calling: https://docs.ollama.com/capabilities/tool-calling
  - langchaingo llms.Tool: https://pkg.go.dev/github.com/tmc/langchaingo/llms#Tool
*/

type ollamaTool struct {
	Type     string              `json:"type"`
	Function ollamaToolFunction  `json:"function"`
}

type ollamaToolFunction struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

type ollamaToolCall struct {
	ID       string                 `json:"id"`
	Type     string                 `json:"type"`
	Function ollamaToolCallFunction `json:"function"`
}

type ollamaToolCallFunction struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type ollamaChatMessage struct {
	Role      string           `json:"role"`
	Content   string           `json:"content"`
	Images    []string         `json:"images,omitempty"`
	ToolCalls []ollamaToolCall `json:"tool_calls,omitempty"`
}

type ollamaChatRequest struct {
	Model    string              `json:"model"`
	Messages []ollamaChatMessage `json:"messages"`
	Tools    []ollamaTool        `json:"tools,omitempty"`
	Stream   bool                `json:"stream"`
	Options  map[string]any      `json:"options,omitempty"`
}

type ollamaChatResponse struct {
	Message ollamaChatMessage `json:"message"`
	Done    bool              `json:"done"`
}

func (c *Client) GenerateWithTools(
	ctx context.Context,
	messages []llms.MessageContent,
	tools []llms.Tool,
) (*llms.ContentResponse, error) {
	chatMessages, err := toOllamaMessages(messages)
	if err != nil {
		return nil, err
	}

	reqBody := ollamaChatRequest{
		Model:    c.AgentModelName(),
		Messages: chatMessages,
		Tools:    toOllamaTools(tools),
		Stream:   false,
		Options: map[string]any{
			"temperature": c.settings.Temperature,
		},
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := strings.TrimRight(c.baseURL(), "/") + "/api/chat"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= http.StatusBadRequest {
		bodyStr := strings.TrimSpace(string(body))
		if strings.Contains(strings.ToLower(bodyStr), "does not support tools") {
			return nil, fmt.Errorf("%w: %s", ErrToolsNotSupported, bodyStr)
		}
		return nil, fmt.Errorf("ollama chat error %d: %s", resp.StatusCode, bodyStr)
	}

	var chatResp ollamaChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, err
	}

	choice := &llms.ContentChoice{
		Content: chatResp.Message.Content,
	}
	for _, call := range chatResp.Message.ToolCalls {
		args := string(call.Function.Arguments)
		if len(call.Function.Arguments) > 0 && call.Function.Arguments[0] != '{' {
			// Ollama may return arguments as a JSON object; normalize to string.
			args = string(call.Function.Arguments)
		}
		choice.ToolCalls = append(choice.ToolCalls, llms.ToolCall{
			ID:   call.ID,
			Type: call.Type,
			FunctionCall: &llms.FunctionCall{
				Name:      call.Function.Name,
				Arguments: args,
			},
		})
	}

	return &llms.ContentResponse{Choices: []*llms.ContentChoice{choice}}, nil
}

func (c *Client) baseURL() string {
	if c.settings.BaseURL != "" {
		return c.settings.BaseURL
	}
	return "http://127.0.0.1:11434"
}

func toOllamaTools(tools []llms.Tool) []ollamaTool {
	out := make([]ollamaTool, 0, len(tools))
	for _, tool := range tools {
		if tool.Function == nil {
			continue
		}
		params, _ := tool.Function.Parameters.(map[string]any)
		out = append(out, ollamaTool{
			Type: "function",
			Function: ollamaToolFunction{
				Name:        tool.Function.Name,
				Description: tool.Function.Description,
				Parameters:  params,
			},
		})
	}
	return out
}

func toOllamaMessages(messages []llms.MessageContent) ([]ollamaChatMessage, error) {
	out := make([]ollamaChatMessage, 0, len(messages))
	for _, mc := range messages {
		msg := ollamaChatMessage{Role: toOllamaRole(mc.Role)}

		for _, part := range mc.Parts {
			switch p := part.(type) {
			case llms.TextContent:
				msg.Content += p.Text
			case llms.BinaryContent:
				msg.Images = append(msg.Images, base64.StdEncoding.EncodeToString(p.Data))
			case llms.ToolCall:
				msg.ToolCalls = append(msg.ToolCalls, ollamaToolCall{
					ID:   p.ID,
					Type: p.Type,
					Function: ollamaToolCallFunction{
						Name:      p.FunctionCall.Name,
						Arguments: json.RawMessage(p.FunctionCall.Arguments),
					},
				})
			case llms.ToolCallResponse:
				out = append(out, ollamaChatMessage{
					Role:    "tool",
					Content: p.Content,
				})
				continue
			default:
				return nil, fmt.Errorf("unsupported message part %T", part)
			}
		}

		if msg.Role == "tool" {
			continue
		}
		if msg.Role != "" || msg.Content != "" || len(msg.ToolCalls) > 0 {
			out = append(out, msg)
		}
	}
	return out, nil
}

func toOllamaRole(role llms.ChatMessageType) string {
	switch role {
	case llms.ChatMessageTypeSystem:
		return "system"
	case llms.ChatMessageTypeAI:
		return "assistant"
	case llms.ChatMessageTypeHuman:
		return "user"
	case llms.ChatMessageTypeTool:
		return "tool"
	default:
		return string(role)
	}
}

// AppendAssistantTurn appends an assistant message that may include tool calls.
func AppendAssistantTurn(messages []llms.MessageContent, choice *llms.ContentChoice) []llms.MessageContent {
	parts := make([]llms.ContentPart, 0, 1+len(choice.ToolCalls))
	if strings.TrimSpace(choice.Content) != "" {
		parts = append(parts, llms.TextContent{Text: choice.Content})
	}
	for _, call := range choice.ToolCalls {
		parts = append(parts, call)
	}
	if len(parts) == 0 {
		return messages
	}
	return append(messages, llms.MessageContent{
		Role:  llms.ChatMessageTypeAI,
		Parts: parts,
	})
}

// AppendToolResults appends tool response messages for each tool call.
func AppendToolResults(messages []llms.MessageContent, calls []llms.ToolCall, results map[string]string) []llms.MessageContent {
	for _, call := range calls {
		if call.FunctionCall == nil {
			continue
		}
		id := call.ID
		if id == "" {
			id = call.FunctionCall.Name
		}
		content := results[id]
		if content == "" {
			content = results[call.FunctionCall.Name]
		}
		messages = append(messages, llms.MessageContent{
			Role: llms.ChatMessageTypeTool,
			Parts: []llms.ContentPart{
				llms.ToolCallResponse{
					ToolCallID: id,
					Name:       call.FunctionCall.Name,
					Content:    content,
				},
			},
		})
	}
	return messages
}
