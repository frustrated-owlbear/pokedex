package pokemonstore

// TeamFilter selects party members matching optional criteria.
// All set fields are combined with AND. Empty strings and nil pointers are ignored.
type TeamFilter struct {
	PrimaryType  string
	MinLevel     *int
	MaxLevel     *int
	Level        *int
	Name         string
	DexID        *int
	CaughtDate   string
	CaughtAfter  string
	CaughtBefore string
	SortBy       SortBy
	SortDesc     bool
}
