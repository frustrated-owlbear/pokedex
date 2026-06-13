package pokemonstore

import "fmt"

type SortBy string

const (
	SortBySlot       SortBy = "slot"
	SortByCaughtDate SortBy = "caught_date"
	SortByLevel      SortBy = "level"
	SortByName       SortBy = "name"
)

func ParseSortBy(raw string) (SortBy, error) {
	switch SortBy(raw) {
	case "", SortBySlot, "slot_order", "party":
		return SortBySlot, nil
	case SortByCaughtDate, "catch_date", "caught":
		return SortByCaughtDate, nil
	case SortByLevel:
		return SortByLevel, nil
	case SortByName:
		return SortByName, nil
	default:
		return "", fmt.Errorf("unsupported sort_by %q", raw)
	}
}

func (s SortBy) SQLColumn() (string, error) {
	switch s {
	case "", SortBySlot:
		return "slot_order", nil
	case SortByCaughtDate:
		return "caught_date", nil
	case SortByLevel:
		return "level", nil
	case SortByName:
		return "name COLLATE NOCASE", nil
	default:
		return "", fmt.Errorf("unsupported sort_by %q", s)
	}
}

func ParseSortOrder(raw string) (desc bool, err error) {
	switch raw {
	case "", "asc", "ascending":
		return false, nil
	case "desc", "descending":
		return true, nil
	default:
		return false, fmt.Errorf("unsupported sort_order %q", raw)
	}
}
