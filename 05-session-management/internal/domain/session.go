package domain

import (
	"strings"
	"time"
)

const (
	ObservationCapture         = "capture"
	ObservationBadge           = "badge"
	ObservationBattle          = "battle"
	ObservationPreference      = "preference"
	ObservationFavoritePokemon = "favorite_pokemon"
	ObservationLocation        = "location"
	ObservationNote            = "note"
)

// ValidObservationCategories lists accepted observation categories.
var ValidObservationCategories = []string{
	ObservationCapture,
	ObservationBadge,
	ObservationBattle,
	ObservationPreference,
	ObservationFavoritePokemon,
	ObservationLocation,
	ObservationNote,
}

type Session struct {
	ID        string        `json:"id"`
	StartTime time.Time     `json:"startedAt"`
	EndTime   *time.Time    `json:"endedAt,omitempty"`
	Events    []Observation `json:"events,omitempty"`
	Summary   string        `json:"summary,omitempty"`
}

type Observation struct {
	Timestamp time.Time `json:"timestamp"`
	Category  string    `json:"category"`
	Content   string    `json:"content"`
}

type SessionEventType string

const (
	EventUserMessage      SessionEventType = "user_message"
	EventAssistantMessage SessionEventType = "assistant_message"
	EventObservation      SessionEventType = "observation"
	EventImageObservation SessionEventType = "image_observation"
	EventToolCall         SessionEventType = "tool_call"
	EventStateUpdate      SessionEventType = "state_update"
	EventSummary          SessionEventType = "summary"
)

type SessionEvent struct {
	Timestamp time.Time        `json:"timestamp"`
	EventType SessionEventType `json:"eventType"`
	Category  string           `json:"category"`
	Content   string           `json:"content"`
}

type AgentSessionDecision struct {
	Action              string `json:"action"`
	Reason              string `json:"reason"`
	Observation         string `json:"observation,omitempty"`
	ShouldCompact       bool   `json:"should_compact"`
	ShouldCloseSession  bool   `json:"should_close_session"`
	ShouldStartNew      bool   `json:"should_start_new"`
	PokemonName         string `json:"pokemon_name,omitempty"`
	PokemonHealth       *int   `json:"pokemon_health,omitempty"`
	NeedsClarification  bool   `json:"needs_clarification"`
	ClarificationPrompt string `json:"clarification_prompt,omitempty"`
}

const (
	ActionContinueSession   = "continue_session"
	ActionAddObservation    = "add_observation"
	ActionCloseSession      = "close_session"
	ActionCompactSession    = "compact_session"
	ActionStartNewSession   = "start_new_session"
	ActionUpdatePokemon     = "update_pokemon_state"
	ActionAskClarification  = "ask_clarification"
)

// BattleSessionState captures durable facts about the current gameplay session.
type BattleSessionState struct {
	Activity        string `json:"activity,omitempty"`
	OpponentTrainer string `json:"opponent_trainer,omitempty"`
	OpponentPokemon string `json:"opponent_pokemon,omitempty"`
	TrainerGoal     string `json:"trainer_goal,omitempty"`
	CurrentTopic    string `json:"current_topic,omitempty"`
	ActivePokemon   string `json:"active_pokemon,omitempty"`
}

const BattleStateCategory = "battle_state"

func (s BattleSessionState) IsEmpty() bool {
	return s.Activity == "" && s.OpponentTrainer == "" && s.OpponentPokemon == "" &&
		s.TrainerGoal == "" && s.CurrentTopic == "" && s.ActivePokemon == ""
}

func (s BattleSessionState) PromptSummary() string {
	if s.IsEmpty() {
		return ""
	}
	var parts []string
	if s.Activity != "" {
		parts = append(parts, "Activity: "+s.Activity)
	}
	if s.OpponentTrainer != "" {
		parts = append(parts, "Opponent trainer: "+s.OpponentTrainer)
	}
	if s.OpponentPokemon != "" {
		parts = append(parts, "Opponent Pokémon: "+s.OpponentPokemon)
	}
	if s.TrainerGoal != "" {
		parts = append(parts, "Trainer goal: "+s.TrainerGoal)
	}
	if s.CurrentTopic != "" {
		parts = append(parts, "Current topic: "+s.CurrentTopic)
	}
	if s.ActivePokemon != "" {
		parts = append(parts, "Active Pokémon: "+s.ActivePokemon)
	}
	return strings.Join(parts, "\n")
}

func IsValidSessionEventType(eventType SessionEventType) bool {
	switch eventType {
	case EventUserMessage, EventAssistantMessage, EventObservation, EventImageObservation,
		EventToolCall, EventStateUpdate, EventSummary:
		return true
	default:
		return false
	}
}

func EventTypeNeedsCategory(eventType SessionEventType) bool {
	switch eventType {
	case EventObservation, EventImageObservation:
		return true
	default:
		return false
	}
}

func IsValidObservationCategory(category string) bool {
	for _, c := range ValidObservationCategories {
		if c == category {
			return true
		}
	}
	return false
}

// NormalizeObservationCategory maps LLM-invented categories to a valid observation category.
func NormalizeObservationCategory(category string) string {
	category = strings.ToLower(strings.TrimSpace(category))
	category = strings.ReplaceAll(category, " ", "_")
	category = strings.ReplaceAll(category, "-", "_")

	if IsValidObservationCategory(category) {
		return category
	}

	switch category {
	case "trainer_goal", "goal", "objective", "task":
		return ObservationNote
	case "training", "gym", "fight", "combat":
		return ObservationBattle
	case "injury", "health", "state_update", "state", "status", "fainted":
		return ObservationNote
	case "pokemon", "team", "party":
		return ObservationPreference
	case "observation", "event", "memory":
		return ObservationNote
	default:
		return ObservationNote
	}
}
