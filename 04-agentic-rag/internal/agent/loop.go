package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/frustrated-owlbear/pokedex/04-agentic-rag/internal/llm"
	"github.com/tmc/langchaingo/llms"
)

const agentSystemPrompt = `You are a Pokédex from the Kanto region, as in Season 1 of the Pokémon anime.

You can call tools to gather context before answering. Use tools when you need:
- current time or weather (clock)
- trainer location (gps)
- owned Pokémon (pokemon_db)
- past conversations (session_memory)
- Pokémon facts and Kanto lore (knowledge_search)

When you have enough information, reply with a final answer only (no tool calls).
Keep answers brief and matter-of-fact (one to three sentences).
Use only Kanto Season 1 knowledge unless retrieved facts say otherwise.
Do not ask follow-up questions.`

type Input struct {
	Prompt        string
	ImageBase64   string
	ImageMIME     string
	SessionID     string
}

type SessionRecorder interface {
	SaveTurn(ctx context.Context, sessionID, userInput, finalAnswer, traceSummary string) error
	LastSummary() string
}

type Loop struct {
	client        *llm.Client
	registry      *Registry
	maxIterations int
	sessions      SessionRecorder
}

func NewLoop(client *llm.Client, registry *Registry, maxIterations int, sessions SessionRecorder) *Loop {
	if maxIterations <= 0 {
		maxIterations = 5
	}
	return &Loop{
		client:        client,
		registry:      registry,
		maxIterations: maxIterations,
		sessions:      sessions,
	}
}

func (l *Loop) Run(
	ctx context.Context,
	input Input,
	onTrace func(TraceStep),
	onChunk func(string),
) error {
	if onTrace == nil {
		onTrace = func(TraceStep) {}
	}
	if onChunk == nil {
		onChunk = func(string) {}
	}

	question := strings.TrimSpace(input.Prompt)
	imageData, _, err := llm.DecodeImageInput(input.ImageBase64, input.ImageMIME)
	if err != nil {
		return err
	}
	if question == "" && len(imageData) == 0 {
		return llm.ErrEmptyInput
	}

	onTrace(NewTraceStep(StepEvent, "New Observation", question))
	onTrace(NewTraceStep(StepEvent, "GPS Update", "Location context available via gps tool"))

	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextPart(agentSystemPrompt)},
		},
	}

	var humanParts []llms.ContentPart
	if len(imageData) > 0 {
		mime := strings.TrimSpace(input.ImageMIME)
		if mime == "" {
			mime = "image/jpeg"
		}
		humanParts = append(humanParts, llms.BinaryPart(mime, imageData))
	}
	if question != "" {
		humanParts = append(humanParts, llms.TextPart(question))
	}
	messages = append(messages, llms.MessageContent{
		Role:  llms.ChatMessageTypeHuman,
		Parts: humanParts,
	})

	toolDefs := l.registry.Definitions()
	var streamed bool
	traceSummary := strings.Builder{}

	for iteration := 0; iteration < l.maxIterations; iteration++ {
		resp, err := l.client.GenerateWithTools(ctx, messages, toolDefs)
		if err != nil {
			if llm.IsToolsNotSupported(err) {
				return l.runPrefetchFallback(ctx, input, question, messages, onTrace, onChunk)
			}
			return err
		}
		if len(resp.Choices) == 0 {
			return errors.New("empty model response")
		}
		choice := resp.Choices[0]

		if strings.TrimSpace(choice.Content) != "" {
			onTrace(NewTraceStep(StepThought, "Analyze Situation", choice.Content))
			traceSummary.WriteString("thought: ")
			traceSummary.WriteString(truncateDetail(choice.Content, 120))
			traceSummary.WriteByte('\n')
		}

		if len(choice.ToolCalls) == 0 {
			onTrace(NewTraceStep(StepFinalAnswer, "Response Ready", truncateDetail(choice.Content, 200)))
			onTrace(NewTraceStep(StepEvent, "Response Ready", "Streaming final answer to trainer"))
			var streamBuilder strings.Builder
			streamErr := l.client.StreamChat(ctx, llm.PrepareStreamMessages(messages), func(chunk string) {
				streamBuilder.WriteString(chunk)
				onChunk(chunk)
			})
			finalAnswer := strings.TrimSpace(streamBuilder.String())
			if streamErr != nil || finalAnswer == "" {
				finalAnswer = strings.TrimSpace(choice.Content)
				if finalAnswer != "" {
					onChunk(finalAnswer)
				}
			}
			if finalAnswer == "" && streamErr != nil {
				return streamErr
			}
			streamed = true
			if l.sessions != nil {
				sessionID := input.SessionID
				if sessionID == "" {
					sessionID = "default"
				}
				_ = l.sessions.SaveTurn(ctx, sessionID, question, finalAnswer, traceSummary.String())
			}
			return nil
		}

		toolNames := make([]string, 0, len(choice.ToolCalls))
		for _, call := range choice.ToolCalls {
			if call.FunctionCall != nil {
				toolNames = append(toolNames, call.FunctionCall.Name)
			}
		}
		onTrace(NewTraceStep(StepAction, "Use Tools", formatToolArgs(choice.ToolCalls), toolNames...))

		messages = llm.AppendAssistantTurn(messages, choice)

		results := make(map[string]string, len(choice.ToolCalls))
		for _, call := range choice.ToolCalls {
			if call.FunctionCall == nil {
				continue
			}
			result, execErr := l.registry.Execute(ctx, call.FunctionCall.Name, call.FunctionCall.Arguments)
			if execErr != nil {
				result = fmt.Sprintf("tool error: %v", execErr)
			}
			id := call.ID
			if id == "" {
				id = call.FunctionCall.Name
			}
			results[id] = result
			onTrace(NewTraceStep(
				StepObservation,
				"Retrieval Results",
				truncateDetail(result, 500),
				call.FunctionCall.Name,
			))
			traceSummary.WriteString(call.FunctionCall.Name)
			traceSummary.WriteString(": ")
			traceSummary.WriteString(truncateDetail(result, 80))
			traceSummary.WriteByte('\n')
		}
		messages = llm.AppendToolResults(messages, choice.ToolCalls, results)
	}

	if !streamed {
		fallback := "I could not gather enough information in time. Please try a simpler question."
		onTrace(NewTraceStep(StepEvent, "Iteration Limit", fmt.Sprintf("Stopped after %d iterations", l.maxIterations)))
		onTrace(NewTraceStep(StepFinalAnswer, "Response Ready", fallback))
		onChunk(fallback)
		if l.sessions != nil {
			sessionID := input.SessionID
			if sessionID == "" {
				sessionID = "default"
			}
			_ = l.sessions.SaveTurn(ctx, sessionID, question, fallback, traceSummary.String())
		}
	}

	return nil
}

func formatToolArgs(calls []llms.ToolCall) string {
	parts := make([]string, 0, len(calls))
	for _, call := range calls {
		if call.FunctionCall == nil {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s(%s)", call.FunctionCall.Name, call.FunctionCall.Arguments))
	}
	return strings.Join(parts, ", ")
}

// Situation describes live simulation state for the UI.
type Situation struct {
	Location string `json:"location"`
	Region   string `json:"region"`
	Time     string `json:"time"`
	Period   string `json:"period"`
	Weather  string `json:"weather"`
	Memory   string `json:"memory"`
	Tools    []string `json:"tools"`
}

func BuildSituation(clockJSON, gpsJSON, memory string, tools []string) Situation {
	situation := Situation{
		Memory: memory,
		Tools:  tools,
	}
	var clock struct {
		Time    string `json:"time"`
		Period  string `json:"period"`
		Weather string `json:"weather"`
	}
	_ = json.Unmarshal([]byte(clockJSON), &clock)
	situation.Time = clock.Time
	situation.Period = clock.Period
	situation.Weather = clock.Weather

	var gps struct {
		Location string `json:"location"`
		Region   string `json:"region"`
	}
	_ = json.Unmarshal([]byte(gpsJSON), &gps)
	situation.Location = gps.Location
	situation.Region = gps.Region
	return situation
}
