package domain

// TeamPokemon is a party member entry loaded from SQLite.
type TeamPokemon struct {
	ID          int    `json:"id"`
	DexID       int    `json:"dexId"`
	Name        string `json:"name"`
	Level       int    `json:"level"`
	PrimaryType string `json:"primaryType"`
	HP          int    `json:"hp"`
	MaxHP       int    `json:"maxHp"`
	CaughtDate  string `json:"caughtDate"`
	Birthday    string `json:"birthday,omitempty"`
	ImageURL    string `json:"imageUrl"`
}
