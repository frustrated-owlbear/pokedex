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

You can call tools to gather context before answering:
- clock: in-game time and weather
- gps: trainer's current location
- pokemon_db: trainer's current party from the local database; filter by species, type, level, and caught date; sort with sort_by (slot or caught_date) before limit
- session_memory: search past conversations only
- knowledge_search: Pokédex facts and Kanto lore

Tool selection rules:
- Questions about "my team", "my Pokémon", "party", or "what I own" -> pokemon_db
- First/earliest caught Pokémon -> pokemon_db with sort_by=caught_date, sort_order=asc, limit=1
- Most recently caught Pokémon -> pokemon_db with sort_by=caught_date, sort_order=desc, limit=1
- First party slot / lead Pokémon -> pokemon_db with sort_by=slot, limit=1
- Questions about earlier chats or "last time we discussed" -> session_memory
- Never use session_memory to answer what Pokémon the trainer currently owns

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
			Parts: []llms.ContentPart{llms.TextPart(buildSystemPrompt(question))},
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
		plannedCalls := make([]plannedToolCall, 0, len(choice.ToolCalls))
		for _, call := range choice.ToolCalls {
			if call.FunctionCall == nil {
				continue
			}
			plan := planToolCall(question, call.FunctionCall)
			plannedCalls = append(plannedCalls, plan)
			toolNames = append(toolNames, plan.name)
		}
		onTrace(NewTraceStep(StepAction, "Use Tools", formatPlannedToolArgs(plannedCalls), toolNames...))

		messages = llm.AppendAssistantTurn(messages, choice)

		results := make(map[string]string, len(choice.ToolCalls))
		planIndex := 0
		for _, call := range choice.ToolCalls {
			if call.FunctionCall == nil {
				continue
			}
			plan := plannedCalls[planIndex]
			planIndex++

			if plan.corrected {
				onTrace(NewTraceStep(
					StepEvent,
					"Route Correction",
					fmt.Sprintf(
						"Model requested %s(%s); redirected to %s(%s).",
						plan.requestedName,
						plan.requestedArgs,
						plan.name,
						plan.args,
					),
					plan.name,
				))
			}
			result, execErr := l.registry.Execute(ctx, plan.name, plan.args)
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
				plan.name,
			))
			traceSummary.WriteString(plan.name)
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

func buildSystemPrompt(question string) string {
	prompt := agentSystemPrompt
	if hint := toolHintForQuestion(question); hint != "" {
		prompt += "\n\n" + hint
	}
	return prompt
}

type plannedToolCall struct {
	requestedName string
	requestedArgs string
	name          string
	args          string
	corrected     bool
}

func planToolCall(question string, call *llms.FunctionCall) plannedToolCall {
	plan := plannedToolCall{
		requestedName: call.Name,
		requestedArgs: call.Arguments,
		name:          call.Name,
		args:          call.Arguments,
	}
	if plan.name == "session_memory" && mentionsTrainerTeam(strings.ToLower(question)) {
		plan.name = "pokemon_db"
		plan.args = pokemonDBToolArgsForQuestion(question)
		plan.corrected = true
	}
	return plan
}

func formatPlannedToolArgs(plans []plannedToolCall) string {
	parts := make([]string, 0, len(plans))
	for _, plan := range plans {
		parts = append(parts, fmt.Sprintf("%s(%s)", plan.name, plan.args))
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
