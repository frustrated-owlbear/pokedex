package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/domain"
	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/llm"
	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/session"
	"github.com/tmc/langchaingo/llms"
)

const agentSystemPrompt = `You are a Pokédex from the Kanto region, as in Season 1 of the Pokémon anime.

You can call tools to gather context before answering:
- clock: in-game time and weather
- gps: trainer's current location
- pokemon_db: trainer's current party from the local database; filter by species, type, level, and caught date; sort with sort_by (slot or caught_date) before limit
- session_memory: search past gameplay sessions, battles, badges, and trainer habits
- record_observation: save an important gameplay event for long-term memory
- knowledge_search: Pokédex facts and Kanto lore

Tool selection rules:
- Questions about "my team", "my Pokémon", "party", or "what I own" -> pokemon_db
- First/earliest caught Pokémon -> pokemon_db with sort_by=caught_date, sort_order=asc, limit=1
- Most recently caught Pokémon -> pokemon_db with sort_by=caught_date, sort_order=desc, limit=1
- First party slot / lead Pokémon -> pokemon_db with sort_by=slot, limit=1
- Questions about past sessions, habits, badges, or "last time we battled" -> session_memory
- When the trainer reports captures, gym wins, battles, locations, or preferences -> record_observation
- Never use session_memory to answer what Pokémon the trainer currently owns

When you have enough information, reply with a final answer only (no tool calls).
Keep answers brief and matter-of-fact (one to three sentences).
Use only Kanto Season 1 knowledge unless retrieved facts say otherwise.

Session memory is managed automatically. Never ask the trainer to create, end, compact, or save sessions.
Respond naturally when remembering injuries, location changes, or topic shifts (e.g. "Got it — I'll remember that Pikachu fainted.").

Use structured session state and recent messages to resolve pronouns and references:
- "he", "she", "they" usually refer to the opponent trainer named in session state
- "my Pokémon" refers to the trainer's party
- opponent trainer Pokémon choices belong in opponent battle state, not trainer team questions

When the trainer asks which Pokémon to choose or use, treat it as a recommendation request.
Call pokemon_db to inspect the trainer's party, then recommend the best option against the known opponent Pokémon.
Do not ask which Pokémon the trainer will choose when they are asking for advice.

Only ask a short clarification when required information truly cannot be inferred from session state, recent messages, or the trainer's party.

Do not mention fainted or injured party Pokémon unless the trainer's current message is about health, switching Pokémon, or sending out a Pokémon. Opponent actions (e.g. "Brock chose Onix") do not require commenting on your party's HP.`

type Input struct {
	Prompt         string
	ImageBase64    string
	ImageMIME      string
	SessionID      string
	OnSessionReset func()
}

type SessionRecorder interface {
	EnsureActiveSession(ctx context.Context) (string, error)
	SaveTurn(ctx context.Context, sessionID, userInput, finalAnswer, traceSummary string) error
	LastSummary() string
}

// SessionManager extends session recording with autonomous lifecycle control.
type SessionManager interface {
	SessionRecorder
	BuildDecisionContext(ctx context.Context, sessionID, userMessage, imageDescription string, hasImage bool) (session.DecisionInput, error)
	Decide(ctx context.Context, input session.DecisionInput) (domain.AgentSessionDecision, error)
	ApplyDecision(ctx context.Context, sessionID string, decision domain.AgentSessionDecision, userMessage string, hasImage bool) (session.ApplyResult, error)
	SessionContextPrompt(sessionID string) (string, error)
	UpdateBattleState(ctx context.Context, sessionID, userMessage string) (domain.BattleSessionState, error)
	TryBattleRecommendation(ctx context.Context, sessionID, userMessage string) (string, bool, error)
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

	if l.sessions != nil {
		sessionID, err := l.sessions.EnsureActiveSession(ctx)
		if err != nil {
			return err
		}
		input.SessionID = sessionID
	}

	sessionContext := ""
	if mgr, ok := l.sessions.(SessionManager); ok && mgr != nil {
		var err error
		sessionContext, input.SessionID, err = l.runSessionLifecycle(
			ctx,
			mgr,
			input,
			question,
			imageData,
			onTrace,
			onChunk,
		)
		if err != nil {
			return err
		}
		if sessionContext == "__clarification__" || sessionContext == "__recommendation__" {
			return nil
		}
	}

	onTrace(NewTraceStep(StepEvent, "New Observation", question))
	onTrace(NewTraceStep(StepEvent, "GPS Update", "Location context available via gps tool"))

	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextPart(buildSystemPrompt(question, sessionContext))},
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
			finalAnswer := strings.TrimSpace(choice.Content)
			onTrace(NewTraceStep(StepFinalAnswer, "Response Ready", truncateDetail(finalAnswer, 200)))
			onTrace(NewTraceStep(StepEvent, "Response Ready", "Streaming final answer to trainer"))
			if finalAnswer != "" {
				onChunk(finalAnswer)
			} else {
				var streamBuilder strings.Builder
				streamErr := l.client.StreamChat(ctx, llm.PrepareStreamMessages(messages), func(chunk string) {
					streamBuilder.WriteString(chunk)
					onChunk(chunk)
				})
				finalAnswer = strings.TrimSpace(streamBuilder.String())
				if finalAnswer == "" && streamErr != nil {
					return streamErr
				}
			}
			if finalAnswer == "" {
				finalAnswer = "I could not produce an answer with the available context."
				onChunk(finalAnswer)
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

func (l *Loop) runSessionLifecycle(
	ctx context.Context,
	mgr SessionManager,
	input Input,
	question string,
	imageData []byte,
	onTrace func(TraceStep),
	onChunk func(string),
) (string, string, error) {
	mime := strings.TrimSpace(input.ImageMIME)
	if mime == "" && len(imageData) > 0 {
		mime = "image/jpeg"
	}

	imageDescription := ""
	if len(imageData) > 0 {
		ctxSummary, _ := mgr.SessionContextPrompt(input.SessionID)
		desc, err := l.client.DescribeImage(ctx, imageData, mime, ctxSummary)
		if err != nil {
			imageDescription = "Trainer shared an image."
		} else {
			imageDescription = desc
		}
		onTrace(NewTraceStep(StepEvent, "Image Analysis", truncateDetail(imageDescription, 200)))
	}

	decisionInput, err := mgr.BuildDecisionContext(ctx, input.SessionID, question, imageDescription, len(imageData) > 0)
	if err != nil {
		return "", input.SessionID, err
	}

	decision, err := mgr.Decide(ctx, decisionInput)
	if err != nil {
		return "", input.SessionID, err
	}

	applyResult, err := mgr.ApplyDecision(ctx, input.SessionID, decision, question, len(imageData) > 0)
	if err != nil {
		return "", input.SessionID, err
	}

	if applyResult.SessionID != input.SessionID && input.OnSessionReset != nil {
		input.OnSessionReset()
	}

	for _, logLine := range applyResult.Logs {
		title := sessionTraceTitle(logLine)
		onTrace(NewTraceStep(StepEvent, title, logLine))
	}

	sessionID := applyResult.SessionID
	if decision.NeedsClarification {
		state, _ := mgr.UpdateBattleState(ctx, sessionID, question)
		if !state.IsEmpty() {
			if answer, ok, err := mgr.TryBattleRecommendation(ctx, sessionID, question); err == nil && ok {
				onTrace(NewTraceStep(StepFinalAnswer, "Battle Recommendation", answer))
				onChunk(answer)
				_ = mgr.SaveTurn(ctx, sessionID, question, answer, "")
				return "__recommendation__", sessionID, nil
			}
		}
		answer := strings.TrimSpace(decision.ClarificationPrompt)
		if answer == "" {
			answer = "Which Pokémon are you referring to?"
		}
		onTrace(NewTraceStep(StepFinalAnswer, "Clarification Needed", answer))
		onChunk(answer)
		_ = mgr.SaveTurn(ctx, sessionID, question, answer, "")
		return "__clarification__", sessionID, nil
	}

	state, err := mgr.UpdateBattleState(ctx, sessionID, question)
	if err != nil {
		return "", sessionID, err
	}
	if !state.IsEmpty() {
		onTrace(NewTraceStep(StepEvent, "Session State", state.PromptSummary()))
	}
	if answer, ok, err := mgr.TryBattleRecommendation(ctx, sessionID, question); err != nil {
		return "", sessionID, err
	} else if ok {
		onTrace(NewTraceStep(StepFinalAnswer, "Battle Recommendation", answer))
		onChunk(answer)
		_ = mgr.SaveTurn(ctx, sessionID, question, answer, "")
		return "__recommendation__", sessionID, nil
	}

	sessionContext, err := mgr.SessionContextPrompt(sessionID)
	if err != nil {
		return "", sessionID, err
	}
	if decision.PokemonHealth != nil && *decision.PokemonHealth == 0 && decision.PokemonName != "" {
		sessionContext += fmt.Sprintf(
			"\n\nRecent state change: %s fainted (HP 0). Advise visiting a Pokémon Center or switching Pokémon.",
			decision.PokemonName,
		)
	}
	return sessionContext, sessionID, nil
}

func sessionTraceTitle(logLine string) string {
	switch {
	case strings.HasPrefix(logLine, "State Update"):
		return "State Update"
	case strings.HasPrefix(logLine, "Close Session"):
		return "Close Session"
	case strings.HasPrefix(logLine, "Started new session"), strings.HasPrefix(logLine, "Observation:"):
		return "Start Session"
	case strings.HasPrefix(logLine, "ask_clarification"):
		return "Session Decision"
	default:
		return "Session Decision"
	}
}

func buildSystemPrompt(question, sessionContext string) string {
	prompt := agentSystemPrompt
	if sessionContext = strings.TrimSpace(sessionContext); sessionContext != "" {
		sessionContext = filterSessionContextForQuestion(question, sessionContext)
		prompt += "\n\nActive session context:\n" + sessionContext
	}
	if hint := battleToolHintForQuestion(question, sessionContext); hint != "" {
		prompt += "\n\n" + hint
	}
	if hint := healthAdviceHint(question, sessionContext); hint != "" {
		prompt += "\n\n" + hint
	}
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
	if plan.name == "session_memory" && shouldRedirectSessionMemoryToTeamDB(question, call.Arguments) {
		plan.name = "pokemon_db"
		plan.args = pokemonDBToolArgsForQuestion(question)
		plan.corrected = true
	}
	return plan
}

func shouldRedirectSessionMemoryToTeamDB(question, toolArgs string) bool {
	q := strings.ToLower(question)
	if mentionsPastHabits(q) {
		return false
	}
	if mentionsPokemonHealthUpdate(q) || toolArgsMentionHealthUpdate(toolArgs) {
		return false
	}
	return mentionsTrainerTeam(q)
}

func toolArgsMentionHealthUpdate(args string) bool {
	lower := strings.ToLower(args)
	keywords := []string{"fainted", "faint", "injured", "hurt", "health", "hp updated", "health updated", "knocked out"}
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
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
