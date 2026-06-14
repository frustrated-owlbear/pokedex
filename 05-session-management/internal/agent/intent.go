package agent

import "strings"

func toolHintForQuestion(question string) string {
	q := strings.ToLower(strings.TrimSpace(question))
	if q == "" {
		return ""
	}

	switch {
	case mentionsPokemonRecommendation(q):
		return "Tool routing: The trainer wants a Pokémon recommendation for the current battle. Call pokemon_db to inspect their party, then recommend the best option against the opponent Pokémon in session state. Do not ask which Pokémon they will choose."
	case mentionsPokemonHealthUpdate(q):
		return "Tool routing: Pokémon health was updated in session state. Use pokemon_db only if you need current party HP. Give a brief next-step suggestion such as visiting a Pokémon Center or switching Pokémon. Do not call session_memory for current health."
	case mentionsPastHabits(q) || mentionsPastConversation(q):
		return "Tool routing: This question asks about past gameplay sessions or trainer habits. Call session_memory, not pokemon_db."
	case mentionsTrainerTeam(q):
		hint := "Tool routing: This question is about the trainer's current party or owned Pokémon. Call pokemon_db. Do not use session_memory for team or roster questions."
		if details := pokemonDBRoutingHint(q); details != "" {
			hint += " " + details
		}
		return hint
	case mentionsLocation(q):
		return "Tool routing: Call gps for the trainer's current location."
	case mentionsTimeOrWeather(q):
		return "Tool routing: Call clock for in-game time and weather."
	case mentionsPokemonFacts(q) && !mentionsTrainerTeam(q):
		return "Tool routing: Call knowledge_search for Pokédex facts and Kanto lore."
	default:
		return ""
	}
}

func mentionsPokemonRecommendation(q string) bool {
	q = strings.ToLower(strings.TrimSpace(q))
	phrases := []string{
		"what pokemon should i choose",
		"what pokémon should i choose",
		"what pokemon should i chose",
		"what pokémon should i chose",
		"which pokemon should i choose",
		"which pokémon should i choose",
		"which pokemon should i use",
		"which pokémon should i use",
		"who should i use against",
		"recommend a pokemon",
		"recommend a pokémon",
	}
	for _, phrase := range phrases {
		if strings.Contains(q, phrase) {
			return true
		}
	}
	return (strings.Contains(q, "should i choose") || strings.Contains(q, "should i chose")) &&
		(strings.Contains(q, "pokemon") || strings.Contains(q, "pokémon"))
}

func mentionsPokemonHealthUpdate(q string) bool {
	q = strings.ToLower(strings.TrimSpace(q))
	phrases := []string{
		"fainted", "faint", "injured", "hurt", "knocked out", "defeated",
		"health updated", "hp updated", "health is 0", "hp is 0", "0 hp", "no hp",
	}
	for _, phrase := range phrases {
		if strings.Contains(q, phrase) {
			return true
		}
	}
	return false
}

func battleToolHintForQuestion(question, sessionContext string) string {
	q := strings.ToLower(strings.TrimSpace(question))
	if !mentionsPokemonRecommendation(q) {
		return ""
	}
	if !strings.Contains(strings.ToLower(sessionContext), "opponent pokémon:") {
		return ""
	}
	return "Battle routing: Opponent Pokémon is known from session state. Recommend from the trainer's party using pokemon_db. Do not ask the trainer which Pokémon they will pick."
}

func healthAdviceHint(question, sessionContext string) string {
	freshFaint := strings.Contains(sessionContext, "Recent state change:")
	if !freshFaint && !mentionsPokemonHealthUpdate(question) {
		return ""
	}
	return "Response guidance: A party Pokémon has 0 HP. Confirm its status, then briefly advise visiting a Pokémon Center or switching to another healthy Pokémon."
}

var staleHealthContextMarkers = []string{
	"health updated", "hp updated", "fainted", "knocked out", "hp 0", "0 hp",
}

func filterSessionContextForQuestion(question, sessionContext string) string {
	if mentionsPokemonHealthUpdate(question) {
		return sessionContext
	}

	prefix, freshSuffix, hasFresh := strings.Cut(sessionContext, "\n\nRecent state change:")
	filtered := filterStaleHealthEvents(prefix)
	if hasFresh {
		return strings.TrimSpace(filtered + "\n\nRecent state change:" + freshSuffix)
	}
	return filtered
}

func filterStaleHealthEvents(sessionContext string) string {
	lines := strings.Split(sessionContext, "\n")
	filtered := make([]string, 0, len(lines))
	for _, line := range lines {
		lower := strings.ToLower(line)
		skip := false
		for _, marker := range staleHealthContextMarkers {
			if strings.Contains(lower, marker) {
				skip = true
				break
			}
		}
		if !skip {
			filtered = append(filtered, line)
		}
	}
	return strings.TrimSpace(strings.Join(filtered, "\n"))
}

func mentionsPastHabits(q string) bool {
	phrases := []string{
		"usually", "typically", "have used", "past battles", "against electric",
		"my habit", "i tend to", "in past sessions", "previous battles",
	}
	for _, phrase := range phrases {
		if strings.Contains(q, phrase) {
			return true
		}
	}
	return false
}

func mentionsTrainerTeam(q string) bool {
	if mentionsPastHabits(q) {
		return false
	}
	teamPhrases := []string{
		"my team", "my pokemon", "my pokémon", "my party",
		"first pokemon", "first pokémon",
		"owned pokemon", "owned pokémon",
		"what pokemon do i have", "what pokémon do i have",
		"what pokemon am i", "what pokémon am i",
		"do i own", "my roster", "party member",
		"pokemon i have", "pokémon i have",
		"pokemon i own", "pokémon i own",
	}
	for _, phrase := range teamPhrases {
		if strings.Contains(q, phrase) {
			return true
		}
	}
	if strings.Contains(q, "which pokemon do i") || strings.Contains(q, "which pokémon do i") {
		return !mentionsPastHabits(q)
	}
	return strings.Contains(q, "my ") && (strings.Contains(q, "pokemon") || strings.Contains(q, "pokémon"))
}

func mentionsPastConversation(q string) bool {
	phrases := []string{
		"last time", "earlier", "before", "previous session",
		"remember when", "we talked", "you said", "last conversation",
		"past session", "earlier session",
	}
	for _, phrase := range phrases {
		if strings.Contains(q, phrase) {
			return true
		}
	}
	return false
}

func mentionsLocation(q string) bool {
	phrases := []string{"where am i", "current location", "where are we", "what route", "what area"}
	for _, phrase := range phrases {
		if strings.Contains(q, phrase) {
			return true
		}
	}
	return false
}

func mentionsTimeOrWeather(q string) bool {
	phrases := []string{"what time", "what's the weather", "what is the weather", "morning or night", "period of day"}
	for _, phrase := range phrases {
		if strings.Contains(q, phrase) {
			return true
		}
	}
	return false
}

func mentionsPokemonFacts(q string) bool {
	return strings.Contains(q, "pokemon") || strings.Contains(q, "pokémon")
}

func pokemonDBRoutingHint(q string) string {
	switch {
	case mentionsRecentlyCaught(q):
		return `Use pokemon_db with {"sort_by":"caught_date","sort_order":"desc","limit":1} for the most recently caught Pokémon.`
	case mentionsCaughtOrder(q):
		return `Use pokemon_db with {"sort_by":"caught_date","sort_order":"asc","limit":1} for the first/earliest caught Pokémon. Do not rely on limit/offset alone.`
	case mentionsPartyOrder(q):
		return `Use pokemon_db with {"sort_by":"slot","limit":1} for the first party slot / lead Pokémon.`
	default:
		return ""
	}
}

func pokemonDBToolArgsForQuestion(question string) string {
	q := strings.ToLower(strings.TrimSpace(question))
	switch {
	case mentionsRecentlyCaught(q):
		return `{"sort_by":"caught_date","sort_order":"desc","limit":1}`
	case mentionsCaughtOrder(q):
		return `{"sort_by":"caught_date","sort_order":"asc","limit":1}`
	case mentionsPartyOrder(q):
		return `{"sort_by":"slot","limit":1}`
	case strings.Contains(q, "first"):
		return `{"sort_by":"caught_date","sort_order":"asc","limit":1}`
	default:
		return `{}`
	}
}

func mentionsCaughtOrder(q string) bool {
	phrases := []string{
		"first caught", "first one i caught", "first one i've caught",
		"earliest caught", "earliest catch", "first catch",
		"oldest pokemon", "oldest pokémon", "oldest one i caught",
		"when did i catch my first", "first pokemon i caught", "first pokémon i caught",
	}
	for _, phrase := range phrases {
		if strings.Contains(q, phrase) {
			return true
		}
	}
	return strings.Contains(q, "caught") && strings.Contains(q, "first")
}

func mentionsRecentlyCaught(q string) bool {
	phrases := []string{
		"most recently caught", "recently caught", "latest catch",
		"last caught", "last pokemon i caught", "last pokémon i caught",
		"newest pokemon", "newest pokémon", "newest catch",
	}
	for _, phrase := range phrases {
		if strings.Contains(q, phrase) {
			return true
		}
	}
	return false
}

func mentionsPartyOrder(q string) bool {
	phrases := []string{
		"first party", "first in my team", "first in my party",
		"lead pokemon", "lead pokémon", "first slot", "party leader",
	}
	for _, phrase := range phrases {
		if strings.Contains(q, phrase) {
			return true
		}
	}
	return false
}
