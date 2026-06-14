package session

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/domain"
	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/llm"
)

type DecisionInput struct {
	UserMessage      string
	ImageDescription string
	HasImage         bool
	SessionID        string
	RecentEvents     []domain.SessionEvent
	Team             []domain.TeamPokemon
	Location         string
	LastEndedSummary string
	BattleState      domain.BattleSessionState
}

type Decider struct {
	client *llm.Client
}

func NewDecider(client *llm.Client) *Decider {
	return &Decider{client: client}
}

const decisionSystemPrompt = `You decide how a Pokédex should manage the trainer's gameplay session before answering.
Return ONLY valid JSON matching this schema:
{
  "action": "continue_session|add_observation|close_session|compact_session|start_new_session|update_pokemon_state|ask_clarification",
  "reason": "brief explanation",
  "observation": "optional observation text",
  "should_compact": false,
  "should_close_session": false,
  "should_start_new": false,
  "pokemon_name": "optional",
  "pokemon_health": null,
  "needs_clarification": false,
  "clarification_prompt": "optional"
}

Rules:
- Continue the current session when the message fits ongoing goals.
- Record observations for injuries, battles, locations, and status changes.
- Close and compact when the topic shifts significantly, a goal is resolved, or context is stale.
- Start a new session after a major location/context shift (especially image-only location updates).
- Update pokemon_health when fainted (0) or clearly injured; infer pokemon from session context when possible.
- Ask clarification only when a state change is implied but the Pokémon cannot be inferred.
- Never ask the trainer to manually manage sessions.`

func (d *Decider) Decide(ctx context.Context, input DecisionInput) (domain.AgentSessionDecision, error) {
	if decision, ok := TryDecideHealthUpdate(input); ok {
		return decision, nil
	}
	if d.client != nil {
		decision, err := d.DecideWithLLM(ctx, input)
		if err == nil {
			if override, ok := TryDecideHealthUpdate(input); ok {
				return override, nil
			}
			return decision, nil
		}
	}
	return DecideHeuristic(input), nil
}

func (d *Decider) DecideWithLLM(ctx context.Context, input DecisionInput) (domain.AgentSessionDecision, error) {
	prompt := formatDecisionPrompt(input)
	raw, err := d.client.Complete(ctx, decisionSystemPrompt, prompt)
	if err != nil {
		return domain.AgentSessionDecision{}, err
	}
	return ParseDecisionJSON(raw)
}

func formatDecisionPrompt(input DecisionInput) string {
	var b strings.Builder
	b.WriteString("Trainer message: ")
	if input.UserMessage != "" {
		b.WriteString(input.UserMessage)
	} else if input.HasImage {
		b.WriteString("(image only, no text)")
	}
	b.WriteByte('\n')
	if input.ImageDescription != "" {
		b.WriteString("Image description: ")
		b.WriteString(input.ImageDescription)
		b.WriteByte('\n')
	}
	if input.Location != "" {
		b.WriteString("Current GPS location: ")
		b.WriteString(input.Location)
		b.WriteByte('\n')
	}
	if input.LastEndedSummary != "" {
		b.WriteString("Last ended session summary: ")
		b.WriteString(input.LastEndedSummary)
		b.WriteByte('\n')
	}
	if len(input.Team) > 0 {
		b.WriteString("Trainer team: ")
		names := make([]string, 0, len(input.Team))
		for _, p := range input.Team {
			names = append(names, fmt.Sprintf("%s (%d/%d HP)", p.Name, p.HP, p.MaxHP))
		}
		b.WriteString(strings.Join(names, ", "))
		b.WriteByte('\n')
	}
	if len(input.RecentEvents) > 0 {
		b.WriteString("Recent session events:\n")
		for _, event := range input.RecentEvents {
			fmt.Fprintf(&b, "- [%s] %s\n", event.EventType, event.Content)
		}
	}
	if summary := input.BattleState.PromptSummary(); summary != "" {
		b.WriteString("\nStructured battle state:\n")
		b.WriteString(summary)
		b.WriteByte('\n')
	}
	b.WriteString("\nDecide the session action JSON.")
	return b.String()
}

func ParseDecisionJSON(raw string) (domain.AgentSessionDecision, error) {
	raw = strings.TrimSpace(raw)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var decision domain.AgentSessionDecision
	if err := json.Unmarshal([]byte(raw), &decision); err != nil {
		return domain.AgentSessionDecision{}, err
	}
	normalizeDecision(&decision)
	return decision, nil
}

func normalizeDecision(decision *domain.AgentSessionDecision) {
	switch decision.Action {
	case domain.ActionUpdatePokemon:
		if decision.PokemonHealth == nil && strings.Contains(strings.ToLower(decision.Reason+decision.Observation), "faint") {
			zero := 0
			decision.PokemonHealth = &zero
		}
	case domain.ActionAskClarification:
		decision.NeedsClarification = true
	case domain.ActionCloseSession, domain.ActionCompactSession:
		decision.ShouldCloseSession = true
		decision.ShouldCompact = decision.Action == domain.ActionCompactSession || decision.ShouldCompact
	case domain.ActionStartNewSession:
		decision.ShouldStartNew = true
	case domain.ActionAddObservation:
		if decision.Observation == "" {
			decision.Observation = decision.Reason
		}
	}
}

func DecideHeuristic(input DecisionInput) domain.AgentSessionDecision {
	msg := strings.TrimSpace(input.UserMessage)
	lower := strings.ToLower(msg)

	if input.HasImage && msg == "" {
		return decideImageOnly(input)
	}

	if mentionsFainted(lower) || mentionsHurt(lower) || mentionsHPChange(lower) {
		return decideHealthUpdate(input, lower)
	}

	if msg != "" {
		return domain.AgentSessionDecision{
			Action:      domain.ActionContinueSession,
			Reason:      "Message continues the current session",
			Observation: msg,
		}
	}

	return domain.AgentSessionDecision{
		Action: domain.ActionContinueSession,
		Reason: "No session change needed",
	}
}

func decideImageOnly(input DecisionInput) domain.AgentSessionDecision {
	desc := strings.ToLower(input.ImageDescription)
	battleContext := sessionMentionsBattle(input.RecentEvents)
	locationImage := strings.Contains(desc, "forest") ||
		strings.Contains(desc, "path") ||
		strings.Contains(desc, "route") ||
		strings.Contains(desc, "road") ||
		strings.Contains(desc, "field")

	if battleContext && locationImage {
		obs := "Trainer appears to have moved to a new outdoor area."
		if input.ImageDescription != "" {
			obs = fmt.Sprintf("Trainer appears to be in: %s", strings.TrimSpace(input.ImageDescription))
		}
		return domain.AgentSessionDecision{
			Action:             domain.ActionStartNewSession,
			Reason:             "Image suggests a location shift unrelated to the prior battle topic",
			Observation:        obs,
			ShouldCloseSession: true,
			ShouldCompact:      true,
			ShouldStartNew:     true,
		}
	}

	obs := input.ImageDescription
	if obs == "" {
		obs = "Trainer shared an image observation."
	}
	return domain.AgentSessionDecision{
		Action:      domain.ActionAddObservation,
		Reason:      "Image recorded for current session context",
		Observation: obs,
	}
}

func decideHealthUpdate(input DecisionInput, lower string) domain.AgentSessionDecision {
	name, ambiguous := InferPokemonName(input.UserMessage, input.RecentEvents, input.Team, input.BattleState)
	if ambiguous {
		return domain.AgentSessionDecision{
			Action:              domain.ActionAskClarification,
			Reason:              "Multiple Pokémon could match; need clarification",
			NeedsClarification:  true,
			ClarificationPrompt: "Which Pokémon fainted?",
		}
	}
	if name == "" && (mentionsFainted(lower) || mentionsHPChange(lower)) {
		return domain.AgentSessionDecision{
			Action:              domain.ActionAskClarification,
			Reason:              "Could not infer which Pokémon fainted",
			NeedsClarification:  true,
			ClarificationPrompt: "Which Pokémon fainted?",
		}
	}

	hp := injuredHP(lower)
	if explicit, ok := parseExplicitHP(lower); ok {
		hp = explicit
	}
	if mentionsFainted(lower) {
		hp = 0
	}

	obs := fmt.Sprintf("%s health updated to %d.", name, hp)
	if hp == 0 {
		obs = fmt.Sprintf("%s fainted. Health updated to 0.", name)
	} else if mentionsHurt(lower) {
		obs = fmt.Sprintf("%s looks injured. Health updated to %d.", name, hp)
	}

	return domain.AgentSessionDecision{
		Action:        domain.ActionUpdatePokemon,
		Reason:        "Trainer reported a Pokémon health change",
		Observation:   obs,
		PokemonName:   name,
		PokemonHealth: &hp,
	}
}

// TryDecideHealthUpdate returns a health-update decision when the message implies one.
func TryDecideHealthUpdate(input DecisionInput) (domain.AgentSessionDecision, bool) {
	lower := strings.ToLower(strings.TrimSpace(input.UserMessage))
	if !(mentionsFainted(lower) || mentionsHurt(lower) || mentionsHPChange(lower)) {
		return domain.AgentSessionDecision{}, false
	}
	return decideHealthUpdate(input, lower), true
}

func mentionsHPChange(lower string) bool {
	phrases := []string{
		"health updated", "hp updated", "health is 0", "hp is 0", "hp 0",
		"at 0 hp", "0/42", "knocked out", "defeated", "can't battle", "cannot battle",
		"no hp left", "out of hp",
	}
	for _, phrase := range phrases {
		if strings.Contains(lower, phrase) {
			return true
		}
	}
	return false
}

func parseExplicitHP(lower string) (int, bool) {
	patterns := []string{
		`health (?:updated |set |is |at )?to (\d+)`,
		`hp (?:updated |set |is |at )?to (\d+)`,
		`(\d+) hp`,
		`(\d+)/\d+ hp`,
	}
	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if match := re.FindStringSubmatch(lower); len(match) == 2 {
			var hp int
			if _, err := fmt.Sscanf(match[1], "%d", &hp); err == nil {
				return hp, true
			}
		}
	}
	return 0, false
}

func mentionsFainted(lower string) bool {
	return strings.Contains(lower, "fainted") || strings.Contains(lower, "faint")
}

func mentionsHurt(lower string) bool {
	return strings.Contains(lower, "injured") || strings.Contains(lower, "hurt")
}

func injuredHP(lower string) int {
	if mentionsFainted(lower) {
		return 0
	}
	if strings.Contains(lower, "tired") {
		return 5
	}
	return 10
}

func sessionMentionsBattle(events []domain.SessionEvent) bool {
	for _, event := range events {
		text := strings.ToLower(event.Content + " " + event.Category)
		if strings.Contains(text, "battle") || strings.Contains(text, "geodude") || strings.Contains(text, "gym") {
			return true
		}
	}
	return false
}

// InferPokemonName resolves which team Pokémon the trainer is referring to.
func InferPokemonName(
	message string,
	events []domain.SessionEvent,
	team []domain.TeamPokemon,
	battleState domain.BattleSessionState,
) (string, bool) {
	lower := strings.ToLower(message)
	for _, pokemon := range team {
		if strings.Contains(lower, strings.ToLower(pokemon.Name)) {
			return pokemon.Name, false
		}
	}

	if isImplicitPartyReference(lower) {
		if name := matchTeamPokemonName(battleState.ActivePokemon, team); name != "" {
			return name, false
		}
		if name := inferRecommendedPokemon(events, team); name != "" {
			return name, false
		}
	}

	candidates := pokemonMentionedInRecentEvents(events, team, 6)
	if isImplicitPartyReference(lower) {
		switch len(candidates) {
		case 1:
			return candidates[0], false
		case 0:
			return "", false
		default:
			return "", true
		}
	}

	if len(candidates) == 1 {
		return candidates[0], false
	}
	if len(candidates) > 1 {
		return "", true
	}
	return "", false
}

func isImplicitPartyReference(lower string) bool {
	phrases := []string{
		"my pokemon", "my pokémon", "my party pokemon", "my party pokémon",
		"it fainted", "it faint", "it got hurt", "it is hurt", "it's hurt",
	}
	for _, phrase := range phrases {
		if strings.Contains(lower, phrase) {
			return true
		}
	}
	return false
}

var recommendPokemonPattern = regexp.MustCompile(`(?i)\buse\s+([A-Za-z]+)`)

func inferRecommendedPokemon(events []domain.SessionEvent, team []domain.TeamPokemon) string {
	for i := len(events) - 1; i >= 0; i-- {
		if events[i].EventType != domain.EventAssistantMessage {
			continue
		}
		match := recommendPokemonPattern.FindStringSubmatch(events[i].Content)
		if len(match) != 2 {
			continue
		}
		if name := matchTeamPokemonName(match[1], team); name != "" {
			return name
		}
	}
	return ""
}

func matchTeamPokemonName(name string, team []domain.TeamPokemon) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	for _, pokemon := range team {
		if strings.EqualFold(pokemon.Name, name) {
			return pokemon.Name
		}
	}
	return ""
}

func pokemonMentionedInRecentEvents(events []domain.SessionEvent, team []domain.TeamPokemon, limit int) []string {
	if limit <= 0 {
		limit = len(events)
	}
	start := len(events) - limit
	if start < 0 {
		start = 0
	}

	seen := make(map[string]struct{})
	var names []string
	for i := len(events) - 1; i >= start; i-- {
		text := strings.ToLower(events[i].Content)
		for _, pokemon := range team {
			if pokemon.HP <= 0 {
				continue
			}
			pname := strings.ToLower(pokemon.Name)
			if strings.Contains(text, pname) {
				if _, ok := seen[pokemon.Name]; !ok {
					seen[pokemon.Name] = struct{}{}
					names = append(names, pokemon.Name)
				}
			}
		}
	}
	return names
}

func pokemonMentionedInEvents(events []domain.SessionEvent, team []domain.TeamPokemon) []string {
	return pokemonMentionedInRecentEvents(events, team, len(events))
}
