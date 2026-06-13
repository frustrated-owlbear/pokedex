package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
)

type prefetchTool struct {
	name string
	args json.RawMessage
}

func (l *Loop) runPrefetchFallback(
	ctx context.Context,
	input Input,
	question string,
	messages []llms.MessageContent,
	onTrace func(TraceStep),
	onChunk func(string),
) error {
	onTrace(NewTraceStep(
		StepThought,
		"Prefetch Context",
		"This model does not support native tool calling. Gathering GPS, clock, team, memory, and knowledge automatically.",
	))

	plans := []prefetchTool{
		{name: "clock", args: json.RawMessage(`{}`)},
		{name: "gps", args: json.RawMessage(`{}`)},
	}
	if mentionsTrainerTeam(strings.ToLower(question)) || question == "" {
		plans = append(plans, prefetchTool{name: "pokemon_db", args: json.RawMessage(`{}`)})
	}
	if question != "" {
		queryArgs, err := json.Marshal(map[string]string{"query": question})
		if err != nil {
			return err
		}
		if mentionsPastConversation(strings.ToLower(question)) {
			plans = append(plans, prefetchTool{name: "session_memory", args: queryArgs})
		}
		if mentionsPokemonFacts(strings.ToLower(question)) && !mentionsTrainerTeam(strings.ToLower(question)) {
			plans = append(plans, prefetchTool{name: "knowledge_search", args: queryArgs})
		}
	}
	if !mentionsTrainerTeam(strings.ToLower(question)) && question != "" {
		// Default party context for general questions unless clearly not team-related.
		hasPokemonDB := false
		for _, plan := range plans {
			if plan.name == "pokemon_db" {
				hasPokemonDB = true
				break
			}
		}
		if !hasPokemonDB {
			plans = append(plans, prefetchTool{name: "pokemon_db", args: json.RawMessage(`{}`)})
		}
	}

	toolNames := make([]string, 0, len(plans))
	var toolNotes strings.Builder
	traceSummary := strings.Builder{}
	traceSummary.WriteString("fallback prefetch\n")

	results := make([]struct {
		name   string
		result string
	}, 0, len(plans))

	for _, plan := range plans {
		tool, ok := l.registry.Get(plan.name)
		if !ok {
			continue
		}
		toolNames = append(toolNames, tool.Name())
		result, execErr := tool.Execute(ctx, plan.args)
		if execErr != nil {
			result = fmt.Sprintf("tool error: %v", execErr)
		}
		results = append(results, struct {
			name   string
			result string
		}{name: plan.name, result: result})
		fmt.Fprintf(&toolNotes, "Tool result %s: %s\n", plan.name, result)
		traceSummary.WriteString(plan.name)
		traceSummary.WriteString(": ")
		traceSummary.WriteString(truncateDetail(result, 80))
		traceSummary.WriteByte('\n')
	}

	onTrace(NewTraceStep(StepAction, "Use Tools", "Automatic prefetch for non-tool model", toolNames...))
	for _, item := range results {
		onTrace(NewTraceStep(
			StepObservation,
			"Retrieval Results",
			truncateDetail(item.result, 500),
			item.name,
		))
	}

	augmented := append([]llms.MessageContent{}, messages...)
	augmented = append(augmented, llms.TextParts(
		llms.ChatMessageTypeHuman,
		"Tool observations gathered automatically:\n"+strings.TrimSpace(toolNotes.String())+"\nProvide your final Pokédex answer now.",
	))

	onTrace(NewTraceStep(StepFinalAnswer, "Response Ready", "Streaming final answer to trainer"))
	onTrace(NewTraceStep(StepEvent, "Response Ready", "Streaming final answer to trainer"))

	var streamBuilder strings.Builder
	if err := l.client.StreamChat(ctx, augmented, func(chunk string) {
		streamBuilder.WriteString(chunk)
		onChunk(chunk)
	}); err != nil {
		return err
	}

	finalAnswer := strings.TrimSpace(streamBuilder.String())
	if finalAnswer == "" {
		finalAnswer = "I could not produce an answer with the available context."
		onChunk(finalAnswer)
	}

	if l.sessions != nil {
		sessionID := input.SessionID
		if sessionID == "" {
			sessionID = "default"
		}
		_ = l.sessions.SaveTurn(ctx, sessionID, question, finalAnswer, traceSummary.String())
	}

	return nil
}
