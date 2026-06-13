package agent

import "strings"

func toolHintForQuestion(question string) string {
	q := strings.ToLower(strings.TrimSpace(question))
	if q == "" {
		return ""
	}

	switch {
	case mentionsTrainerTeam(q):
		hint := "Tool routing: This question is about the trainer's current party or owned Pokémon. Call pokemon_db. Do not use session_memory for team or roster questions."
		if details := pokemonDBRoutingHint(q); details != "" {
			hint += " " + details
		}
		return hint
	case mentionsPastConversation(q):
		return "Tool routing: This question asks about earlier conversations. Call session_memory, not pokemon_db."
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

func mentionsTrainerTeam(q string) bool {
	teamPhrases := []string{
		"my team", "my pokemon", "my pokémon", "my party",
		"first pokemon", "first pokémon",
		"owned pokemon", "owned pokémon",
		"which pokemon do i", "which pokémon do i",
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
	return strings.Contains(q, "my ") && (strings.Contains(q, "pokemon") || strings.Contains(q, "pokémon"))
}

func mentionsPastConversation(q string) bool {
	phrases := []string{
		"last time", "earlier", "before", "previous session",
		"remember when", "we talked", "you said", "last conversation",
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
