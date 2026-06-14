package session

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/domain"
)

var (
	trainWithPattern    = regexp.MustCompile(`(?i)(?:going to |gonna )?train(?:ing)? with ([a-zA-Z][a-zA-Z\s'-]{1,20})`)
	namedChosePattern   = regexp.MustCompile(`(?i)([a-zA-Z][a-zA-Z\s'-]{1,20})\s+chose\s+([a-zA-Z][a-zA-Z0-9\s'-]{1,20})`)
	pronounChosePattern = regexp.MustCompile(`(?i)\b(?:he|she|they)\s+chose\s+([a-zA-Z][a-zA-Z0-9\s'-]{1,20})`)
	pokemonNamePattern  = regexp.MustCompile(`\b([A-Z][a-z]+(?:[ '-][A-Z][a-z]+)*)\b`)
)

var trainerAliases = map[string]string{
	"brok":  "Brock",
	"brock": "Brock",
	"misty": "Misty",
	"gary":  "Gary",
	"oak":   "Professor Oak",
}

var knownPokemonNames = []string{
	"Onix", "Geodude", "Pikachu", "Bulbasaur", "Squirtle", "Charmander", "Pidgey",
	"Staryu", "Starmie", "Vulpix", "Oddish", "Sandshrew", "Mankey", "Growlithe",
}

// LoadBattleState returns the latest structured battle state from session events.
func LoadBattleState(events []domain.SessionEvent) domain.BattleSessionState {
	for i := len(events) - 1; i >= 0; i-- {
		if state, ok := parseBattleStateEvent(events[i]); ok {
			return state
		}
	}
	return domain.BattleSessionState{}
}

func (s *Store) LoadBattleState(sessionID string) (domain.BattleSessionState, error) {
	var content string
	err := s.db.QueryRow(`
		SELECT content FROM observations
		WHERE session_id = ? AND event_type = ? AND category = ?
		ORDER BY id DESC LIMIT 1
	`, sessionID, string(domain.EventStateUpdate), domain.BattleStateCategory).Scan(&content)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.BattleSessionState{}, nil
		}
		return domain.BattleSessionState{}, err
	}
	var state domain.BattleSessionState
	if err := json.Unmarshal([]byte(content), &state); err != nil {
		return domain.BattleSessionState{}, err
	}
	return state, nil
}

func parseBattleStateEvent(event domain.SessionEvent) (domain.BattleSessionState, bool) {
	if event.EventType != domain.EventStateUpdate || event.Category != domain.BattleStateCategory {
		return domain.BattleSessionState{}, false
	}
	var state domain.BattleSessionState
	if err := json.Unmarshal([]byte(event.Content), &state); err != nil {
		return domain.BattleSessionState{}, false
	}
	return state, true
}

// ExtractBattleState merges durable facts from the current message and prior state.
func ExtractBattleState(message string, prior domain.BattleSessionState, events []domain.SessionEvent) domain.BattleSessionState {
	state := prior
	msg := strings.TrimSpace(message)
	lower := strings.ToLower(msg)

	if match := trainWithPattern.FindStringSubmatch(msg); len(match) == 2 {
		state.Activity = "training"
		state.OpponentTrainer = NormalizeTrainerName(strings.TrimSpace(match[1]))
		state.CurrentTopic = "training battle"
		state.TrainerGoal = fmt.Sprintf("train with %s", state.OpponentTrainer)
	}

	if match := pronounChosePattern.FindStringSubmatch(msg); len(match) == 2 {
		pokemon := NormalizePokemonName(strings.TrimSpace(match[1]))
		if pokemon != "" {
			state.OpponentPokemon = pokemon
		}
		if state.OpponentTrainer == "" {
			state.OpponentTrainer = inferOpponentTrainerFromEvents(events)
		}
		if state.OpponentTrainer != "" && pokemon != "" {
			state.CurrentTopic = "opponent team selection"
			if state.Activity == "" {
				state.Activity = "training"
			}
		}
	} else if match := namedChosePattern.FindStringSubmatch(msg); len(match) == 3 {
		trainerRaw := strings.TrimSpace(match[1])
		pokemon := NormalizePokemonName(strings.TrimSpace(match[2]))
		if !isPronoun(trainerRaw) {
			if trainer := NormalizeTrainerName(trainerRaw); trainer != "" {
				state.OpponentTrainer = trainer
			}
		} else if state.OpponentTrainer == "" {
			state.OpponentTrainer = inferOpponentTrainerFromEvents(events)
		}
		if pokemon != "" {
			state.OpponentPokemon = pokemon
		}
		if state.Activity == "" {
			state.Activity = "training"
		}
		state.CurrentTopic = "opponent team selection"
	}

	if IsPokemonRecommendationRequest(lower) {
		state.TrainerGoal = "get a Pokémon recommendation for the current battle"
		state.CurrentTopic = "team selection advice"
		if state.Activity == "" && state.OpponentTrainer != "" {
			state.Activity = "training"
		}
	}

	if state.Activity == "" && (state.OpponentTrainer != "" || state.OpponentPokemon != "") {
		state.Activity = "battle"
	}

	return mergeBattleState(state, prior)
}

func mergeBattleState(current, prior domain.BattleSessionState) domain.BattleSessionState {
	if current.Activity == "" {
		current.Activity = prior.Activity
	}
	if current.OpponentTrainer == "" {
		current.OpponentTrainer = prior.OpponentTrainer
	}
	if current.OpponentPokemon == "" {
		current.OpponentPokemon = prior.OpponentPokemon
	}
	if current.TrainerGoal == "" {
		current.TrainerGoal = prior.TrainerGoal
	}
	if current.CurrentTopic == "" {
		current.CurrentTopic = prior.CurrentTopic
	}
	if current.ActivePokemon == "" {
		current.ActivePokemon = prior.ActivePokemon
	}
	return current
}

// NormalizeTrainerName maps common misspellings to Kanto trainer names.
func NormalizeTrainerName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	lower := strings.ToLower(name)
	if canonical, ok := trainerAliases[lower]; ok {
		return canonical
	}
	return canonicalName(lower)
}

func canonicalName(name string) string {
	if name == "" {
		return ""
	}
	parts := strings.Fields(strings.ToLower(name))
	for i, part := range parts {
		if len(part) == 0 {
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}

func isPronoun(word string) bool {
	switch strings.ToLower(strings.TrimSpace(word)) {
	case "he", "she", "they", "him", "her", "them":
		return true
	default:
		return false
	}
}

// NormalizePokemonName extracts a known Pokémon species from free text.
func NormalizePokemonName(text string) string {
	text = strings.Trim(strings.TrimSpace(text), "!?.")
	for _, name := range knownPokemonNames {
		if strings.EqualFold(text, name) || strings.Contains(strings.ToLower(text), strings.ToLower(name)) {
			return name
		}
	}
	if match := pokemonNamePattern.FindString(text); match != "" {
		return match
	}
	return strings.TrimSpace(text)
}

func inferOpponentTrainerFromEvents(events []domain.SessionEvent) string {
	for i := len(events) - 1; i >= 0; i-- {
		text := events[i].Content
		if match := trainWithPattern.FindStringSubmatch(text); len(match) == 2 {
			return NormalizeTrainerName(match[1])
		}
		if match := namedChosePattern.FindStringSubmatch(text); len(match) == 3 {
			return NormalizeTrainerName(match[1])
		}
		lower := strings.ToLower(text)
		for alias, canonical := range trainerAliases {
			if strings.Contains(lower, alias) {
				return canonical
			}
		}
	}
	return ""
}

// IsPokemonRecommendationRequest detects when the trainer wants advice, not to name their pick.
func IsPokemonRecommendationRequest(lower string) bool {
	phrases := []string{
		"what pokemon should i choose",
		"what pokémon should i choose",
		"what pokemon should i chose",
		"what pokémon should i chose",
		"which pokemon should i choose",
		"which pokémon should i choose",
		"which pokemon should i use",
		"which pokémon should i use",
		"what should i use",
		"who should i use",
		"recommend a pokemon",
		"recommend a pokémon",
		"best pokemon against",
		"best pokémon against",
	}
	for _, phrase := range phrases {
		if strings.Contains(lower, phrase) {
			return true
		}
	}
	return strings.Contains(lower, "should i choose") &&
		(strings.Contains(lower, "pokemon") || strings.Contains(lower, "pokémon"))
}

// RecommendPokemon picks a party member with type advantage against the opponent.
func RecommendPokemon(team []domain.TeamPokemon, opponentPokemon string) (string, bool) {
	_, answer, ok := RecommendPokemonChoice(team, opponentPokemon)
	return answer, ok
}

func RecommendPokemonChoice(team []domain.TeamPokemon, opponentPokemon string) (domain.TeamPokemon, string, bool) {
	if len(team) == 0 {
		return domain.TeamPokemon{}, "", false
	}
	preferredTypes := preferredTypesAgainst(opponentPokemon)
	for _, preferred := range preferredTypes {
		for _, pokemon := range team {
			if pokemon.HP <= 0 {
				continue
			}
			if strings.EqualFold(pokemon.PrimaryType, preferred) {
				return pokemon, formatRecommendation(pokemon, opponentPokemon), true
			}
		}
	}
	for _, pokemon := range team {
		if pokemon.HP > 0 {
			return pokemon, formatRecommendation(pokemon, opponentPokemon), true
		}
	}
	return domain.TeamPokemon{}, "", false
}

func preferredTypesAgainst(opponent string) []string {
	switch strings.ToLower(opponent) {
	case "onix", "geodude", "graveler", "golem", "sandshrew", "sandslash", "diglett", "dugtrio":
		return []string{"WATER", "GRASS", "ICE"}
	case "voltorb", "electrode", "pikachu", "raichu", "magnemite", "magneton":
		return []string{"GROUND"}
	case "squirtle", "wartortle", "blastoise", "psyduck", "golduck", "poliwag", "poliwhirl", "poliwrath":
		return []string{"GRASS", "ELECTRIC"}
	default:
		return []string{"WATER", "GRASS", "FIRE", "ELECTRIC", "NORMAL"}
	}
}

func formatRecommendation(pokemon domain.TeamPokemon, opponent string) string {
	reason := typeAdvantageReason(pokemon.PrimaryType, opponent)
	return fmt.Sprintf(
		"Use %s — %s moves are strong against %s.",
		pokemon.Name,
		strings.ToUpper(pokemon.PrimaryType[:1])+strings.ToLower(pokemon.PrimaryType[1:]),
		opponent,
	) + " " + reason
}

func typeAdvantageReason(primaryType, opponent string) string {
	switch {
	case strings.EqualFold(opponent, "Onix") && strings.EqualFold(primaryType, "WATER"):
		return "Water-type moves hit Rock and Ground Pokémon hard."
	case strings.EqualFold(opponent, "Onix") && strings.EqualFold(primaryType, "GRASS"):
		return "Grass-type moves are effective against Rock and Ground Pokémon."
	default:
		return fmt.Sprintf("%s-type coverage looks solid here.", strings.ToUpper(primaryType[:1])+strings.ToLower(primaryType[1:]))
	}
}

func (m *Manager) UpdateBattleState(ctx context.Context, sessionID, userMessage string) (domain.BattleSessionState, error) {
	events, err := m.store.RecentEvents(sessionID, 30)
	if err != nil {
		return domain.BattleSessionState{}, err
	}

	prior, err := m.store.LoadBattleState(sessionID)
	if err != nil {
		return domain.BattleSessionState{}, err
	}
	state := ExtractBattleState(userMessage, prior, events)
	if state.IsEmpty() {
		return state, nil
	}

	payload, err := json.Marshal(state)
	if err != nil {
		return domain.BattleSessionState{}, err
	}
	if err := m.store.AddEvent(ctx, sessionID, domain.EventStateUpdate, domain.BattleStateCategory, string(payload)); err != nil {
		return domain.BattleSessionState{}, err
	}
	return state, nil
}

func (m *Manager) BattleStateForSession(sessionID string) (domain.BattleSessionState, error) {
	return m.store.LoadBattleState(sessionID)
}

func (m *Manager) TryBattleRecommendation(ctx context.Context, sessionID, userMessage string) (string, bool, error) {
	if !IsPokemonRecommendationRequest(strings.ToLower(userMessage)) {
		return "", false, nil
	}

	state, err := m.BattleStateForSession(sessionID)
	if err != nil {
		return "", false, err
	}
	if state.OpponentPokemon == "" {
		return "", false, nil
	}

	team, err := m.team.ListTeam()
	if err != nil {
		return "", false, err
	}
	pokemon, answer, ok := RecommendPokemonChoice(team, state.OpponentPokemon)
	if !ok {
		return "", false, nil
	}
	if err := m.RecordActivePokemon(ctx, sessionID, pokemon.Name); err != nil {
		return "", false, err
	}
	return answer, true, nil
}

func (m *Manager) RecordActivePokemon(ctx context.Context, sessionID, pokemonName string) error {
	state, err := m.store.LoadBattleState(sessionID)
	if err != nil {
		return err
	}
	state.ActivePokemon = strings.TrimSpace(pokemonName)
	if state.Activity == "" {
		state.Activity = "battle"
	}
	payload, err := json.Marshal(state)
	if err != nil {
		return err
	}
	return m.store.AddEvent(ctx, sessionID, domain.EventStateUpdate, domain.BattleStateCategory, string(payload))
}
