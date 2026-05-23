package main

import (
	"fmt"
	"strings"

	"github.com/frustrated-owlbear/pokedex/internal/domain"
	"github.com/frustrated-owlbear/pokedex/internal/llm"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// GetTrainerProfile returns trainer sidebar data.
func (a *App) GetTrainerProfile() domain.TrainerProfile {
	return domain.TrainerProfile{
		TrainerID:        "000123",
		AvatarURL:        "",
		ConnectionStatus: "ONLINE",
	}
}

// GetCurrentSituation returns placeholder analysis and advice.
func (a *App) GetCurrentSituation() domain.CurrentSituation {
	return domain.CurrentSituation{
		Summary: "Your party is balanced for early-route trainers.",
		Advice: []string{
			"Lead with Squirtle against fire-type gym leaders.",
			"Stock up on potions before the next badge battle.",
			"Consider teaching Bulbasaur a status move for utility.",
		},
	}
}

// GetMyTeam returns the trainer's current party with image URLs.
func (a *App) GetMyTeam() []domain.Pokemon {
	team := a.MyPokemons()
	result := make([]domain.Pokemon, len(team))
	for i, p := range team {
		id := pokemonIDByName(p.Name)
		result[i] = domain.Pokemon{
			Name:     p.Name,
			ID:       id,
			ImageURL: domain.ImageURL(id),
		}
	}
	return result
}

// GetPokemonList returns pokedex entries filtered by query.
func (a *App) GetPokemonList(query string) []domain.PokemonInfo {
	all := []domain.PokemonInfo{
		{ID: 1, Name: "Bulbasaur", Types: "Grass/Poison", Region: "Kanto"},
		{ID: 4, Name: "Charmander", Types: "Fire", Region: "Kanto"},
		{ID: 7, Name: "Squirtle", Types: "Water", Region: "Kanto"},
		{ID: 25, Name: "Pikachu", Types: "Electric", Region: "Kanto"},
		{ID: 133, Name: "Eevee", Types: "Normal", Region: "Kanto"},
	}
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return all
	}
	var filtered []domain.PokemonInfo
	for _, p := range all {
		if strings.Contains(strings.ToLower(p.Name), q) ||
			strings.Contains(strings.ToLower(p.Types), q) {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

// GetPokemonDetail returns detailed info for a pokemon by name.
func (a *App) GetPokemonDetail(name string) domain.PokemonDetail {
	name = strings.TrimSpace(name)
	info := domain.PokemonInfo{ID: 0, Name: name, Types: "Unknown", Region: "Kanto"}
	for _, p := range a.GetPokemonList(name) {
		if strings.EqualFold(p.Name, name) {
			info = p
			break
		}
	}
	analysis := a.GetTypeAnalysis(name)
	stats := pokemonPartyStats(strings.ToLower(name))
	return domain.PokemonDetail{
		ID:          info.ID,
		Name:        info.Name,
		ImageURL:    domain.ImageURL(info.ID),
		Types:       info.Types,
		Region:      info.Region,
		Level:       stats.level,
		HP:          stats.hp,
		MaxHP:       stats.maxHP,
		Ability:     stats.ability,
		TypesList:   analysis.Types,
		Strengths:   analysis.Strengths,
		Weaknesses:  analysis.Weaknesses,
		Resistances: analysis.Resistances,
	}
}

type partyStats struct {
	level   int
	hp      int
	maxHP   int
	ability string
}

func pokemonIDByName(name string) int {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "bulbasaur":
		return 1
	case "charmander":
		return 4
	case "squirtle":
		return 7
	case "pikachu":
		return 25
	case "eevee":
		return 133
	default:
		return 0
	}
}

func pokemonPartyStats(name string) partyStats {
	switch name {
	case "bulbasaur":
		return partyStats{level: 18, hp: 52, maxHP: 52, ability: "Overgrow"}
	case "charmander":
		return partyStats{level: 19, hp: 48, maxHP: 48, ability: "Blaze"}
	case "squirtle":
		return partyStats{level: 17, hp: 50, maxHP: 50, ability: "Torrent"}
	default:
		return partyStats{level: 15, hp: 40, maxHP: 45, ability: "—"}
	}
}

// GetTypeAnalysis returns type matchup data for a pokemon name.
func (a *App) GetTypeAnalysis(pokemonName string) domain.TypeAnalysis {
	name := strings.TrimSpace(pokemonName)
	if name == "" {
		name = "Squirtle"
	}
	switch strings.ToLower(name) {
	case "bulbasaur":
		return domain.TypeAnalysis{
			PokemonName: "Bulbasaur",
			Types:       []string{"Grass", "Poison"},
			Strengths:   []string{"Water", "Ground", "Rock"},
			Weaknesses:  []string{"Fire", "Ice", "Flying", "Psychic"},
			Resistances: []string{"Water", "Electric", "Grass", "Fighting"},
		}
	case "charmander":
		return domain.TypeAnalysis{
			PokemonName: "Charmander",
			Types:       []string{"Fire"},
			Strengths:   []string{"Grass", "Ice", "Bug", "Steel"},
			Weaknesses:  []string{"Water", "Ground", "Rock"},
			Resistances: []string{"Fire", "Grass", "Ice", "Bug", "Steel", "Fairy"},
		}
	default:
		return domain.TypeAnalysis{
			PokemonName: name,
			Types:       []string{"Water"},
			Strengths:   []string{"Fire", "Ground", "Rock"},
			Weaknesses:  []string{"Electric", "Grass"},
			Resistances: []string{"Fire", "Water", "Ice", "Steel"},
		}
	}
}

// GetKnowledgeBase returns reference articles.
func (a *App) GetKnowledgeBase() []domain.KnowledgeArticle {
	return []domain.KnowledgeArticle{
		{
			ID:      "type-chart",
			Title:   "Type effectiveness chart",
			Summary: "How attack types interact with defending types.",
		},
		{
			ID:      "evolution",
			Title:   "Evolution basics",
			Summary: "Level, stone, trade, and friendship evolutions explained.",
		},
		{
			ID:      "status",
			Title:   "Status conditions",
			Summary: "Burn, poison, paralysis, sleep, and freeze effects.",
		},
	}
}

// AskPokedex streams an LLM reply for a user prompt; chunks emit as "llm:chunk".
func (a *App) AskPokedex(prompt string) error {
	p := strings.TrimSpace(prompt)
	if p == "" {
		return fmt.Errorf("prompt is required")
	}
	return llm.StreamCompletion(a.ctx, llm.Prompt(p), func(chunk string) {
		runtime.EventsEmit(a.ctx, "llm:chunk", chunk)
	})
}
