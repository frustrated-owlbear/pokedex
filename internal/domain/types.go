package domain

// TrainerProfile holds trainer sidebar data.
type TrainerProfile struct {
	TrainerID        string `json:"trainerId"`
	AvatarURL        string `json:"avatarUrl"`
	ConnectionStatus string `json:"connectionStatus"`
}

// CurrentSituation is advice based on current game state.
type CurrentSituation struct {
	Summary string   `json:"summary"`
	Advice  []string `json:"advice"`
}

// PokemonInfo is a searchable pokedex entry.
type PokemonInfo struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Types  string `json:"types"`
	Region string `json:"region"`
}

// PokemonDetail is full info for a party member.
type PokemonDetail struct {
	ID          int      `json:"id"`
	Name        string   `json:"name"`
	ImageURL    string   `json:"imageUrl"`
	Types       string   `json:"types"`
	Region      string   `json:"region"`
	Level       int      `json:"level"`
	HP          int      `json:"hp"`
	MaxHP       int      `json:"maxHp"`
	Ability     string   `json:"ability"`
	TypesList   []string `json:"typesList"`
	Strengths   []string `json:"strengths"`
	Weaknesses  []string `json:"weaknesses"`
	Resistances []string `json:"resistances"`
}

// TypeAnalysis covers matchups for a pokemon.
type TypeAnalysis struct {
	PokemonName string   `json:"pokemonName"`
	Types       []string `json:"types"`
	Strengths   []string `json:"strengths"`
	Weaknesses  []string `json:"weaknesses"`
	Resistances []string `json:"resistances"`
}

// KnowledgeArticle is a knowledge-base entry.
type KnowledgeArticle struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Summary string `json:"summary"`
}
