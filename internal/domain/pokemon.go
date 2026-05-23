package domain

// Pokemon is a domain model for a pokemon entry.
type Pokemon struct {
	Name     string `json:"name"`
	ID       int    `json:"id"`
	ImageURL string `json:"imageUrl"`
}
