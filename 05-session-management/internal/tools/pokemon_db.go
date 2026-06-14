package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/domain"
	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/pokemonstore"
)

type PokemonDBTool struct {
	store *pokemonstore.SQLiteStore
}

func NewPokemonDBTool(store *pokemonstore.SQLiteStore) *PokemonDBTool {
	return &PokemonDBTool{store: store}
}

func (t *PokemonDBTool) Name() string { return "pokemon_db" }

func (t *PokemonDBTool) Description() string {
	return "Search the trainer's current party in the local Pokémon database. Filter by species name, Pokédex number, elemental type, level, and caught date. Sort with sort_by and sort_order before using limit/offset. For first/earliest caught Pokémon use sort_by=caught_date, sort_order=asc, limit=1. For most recently caught use sort_by=caught_date, sort_order=desc, limit=1. For first party slot / lead Pokémon use sort_by=slot, limit=1. Do not use limit/offset alone for catch-order questions. Do not use session_memory for team or roster questions."
}

func (t *PokemonDBTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{
				"type":        "string",
				"description": "Filter by Pokémon species name (partial match, e.g. \"Bulba\" or \"Pidgey\").",
			},
			"dex_id": map[string]any{
				"type":        "integer",
				"description": "Filter by National Pokédex number.",
			},
			"primary_type": map[string]any{
				"type":        "string",
				"description": "Filter by elemental type (e.g. GRASS, FIRE, NORMAL).",
			},
			"type": map[string]any{
				"type":        "string",
				"description": "Alias for primary_type when filtering by elemental type.",
			},
			"level": map[string]any{
				"type":        "integer",
				"description": "Filter by exact level.",
			},
			"min_level": map[string]any{
				"type":        "integer",
				"description": "Minimum level (inclusive). Ignored when level is set.",
			},
			"max_level": map[string]any{
				"type":        "integer",
				"description": "Maximum level (inclusive). Ignored when level is set.",
			},
			"caught_date": map[string]any{
				"type":        "string",
				"description": "Filter by exact caught date in YYYY-MM-DD format (e.g. \"2024-02-15\").",
			},
			"caught_after": map[string]any{
				"type":        "string",
				"description": "Earliest caught date (inclusive, YYYY-MM-DD). Ignored when caught_date is set.",
			},
			"caught_before": map[string]any{
				"type":        "string",
				"description": "Latest caught date (inclusive, YYYY-MM-DD). Ignored when caught_date is set.",
			},
			"limit": map[string]any{
				"type":        "integer",
				"description": "Maximum number of matching party members to return. Omit to return all matches.",
			},
			"offset": map[string]any{
				"type":        "integer",
				"description": "Number of matching party members to skip after sorting. Defaults to 0.",
			},
			"sort_by": map[string]any{
				"type":        "string",
				"description": "Sort key: slot (party order, default), caught_date, level, or name.",
			},
			"sort_order": map[string]any{
				"type":        "string",
				"description": "Sort direction: asc (default) or desc.",
			},
		},
	}
}

type pokemonDBArgs struct {
	Name         string
	DexID        *int
	PrimaryType  string
	Level        *int
	MinLevel     *int
	MaxLevel     *int
	CaughtDate   string
	CaughtAfter  string
	CaughtBefore string
	SortBy       string
	SortOrder    string
	Limit        *int
	Offset       *int
}

func (a *pokemonDBArgs) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		return nil
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	if v, ok := raw["name"]; ok {
		var name string
		if err := json.Unmarshal(v, &name); err != nil {
			return fmt.Errorf("name: %w", err)
		}
		a.Name = strings.TrimSpace(name)
	}
	if v, ok := raw["dex_id"]; ok {
		dexID, err := parseOptionalInt(v)
		if err != nil {
			return fmt.Errorf("dex_id: %w", err)
		}
		a.DexID = dexID
	}
	if v, ok := raw["primary_type"]; ok {
		var primaryType string
		if err := json.Unmarshal(v, &primaryType); err != nil {
			return fmt.Errorf("primary_type: %w", err)
		}
		a.PrimaryType = strings.TrimSpace(primaryType)
	}
	if v, ok := raw["type"]; ok {
		var elementType string
		if err := json.Unmarshal(v, &elementType); err != nil {
			return fmt.Errorf("type: %w", err)
		}
		elementType = strings.TrimSpace(elementType)
		if a.PrimaryType == "" {
			a.PrimaryType = elementType
		} else if !strings.EqualFold(a.PrimaryType, elementType) {
			return fmt.Errorf("type and primary_type must match when both are provided")
		}
	}
	if v, ok := raw["level"]; ok {
		level, err := parseOptionalInt(v)
		if err != nil {
			return fmt.Errorf("level: %w", err)
		}
		a.Level = level
	}
	if v, ok := raw["min_level"]; ok {
		minLevel, err := parseOptionalInt(v)
		if err != nil {
			return fmt.Errorf("min_level: %w", err)
		}
		a.MinLevel = minLevel
	}
	if v, ok := raw["max_level"]; ok {
		maxLevel, err := parseOptionalInt(v)
		if err != nil {
			return fmt.Errorf("max_level: %w", err)
		}
		a.MaxLevel = maxLevel
	}
	if v, ok := raw["caught_date"]; ok {
		var caughtDate string
		if err := json.Unmarshal(v, &caughtDate); err != nil {
			return fmt.Errorf("caught_date: %w", err)
		}
		a.CaughtDate = strings.TrimSpace(caughtDate)
	}
	if v, ok := raw["caught_after"]; ok {
		var caughtAfter string
		if err := json.Unmarshal(v, &caughtAfter); err != nil {
			return fmt.Errorf("caught_after: %w", err)
		}
		a.CaughtAfter = strings.TrimSpace(caughtAfter)
	}
	if v, ok := raw["caught_before"]; ok {
		var caughtBefore string
		if err := json.Unmarshal(v, &caughtBefore); err != nil {
			return fmt.Errorf("caught_before: %w", err)
		}
		a.CaughtBefore = strings.TrimSpace(caughtBefore)
	}
	if v, ok := raw["limit"]; ok {
		limit, err := parseOptionalInt(v)
		if err != nil {
			return fmt.Errorf("limit: %w", err)
		}
		a.Limit = limit
	}
	if v, ok := raw["offset"]; ok {
		offset, err := parseOptionalInt(v)
		if err != nil {
			return fmt.Errorf("offset: %w", err)
		}
		a.Offset = offset
	}
	if v, ok := raw["sort_by"]; ok {
		var sortBy string
		if err := json.Unmarshal(v, &sortBy); err != nil {
			return fmt.Errorf("sort_by: %w", err)
		}
		a.SortBy = strings.TrimSpace(sortBy)
	}
	if v, ok := raw["sort_order"]; ok {
		var sortOrder string
		if err := json.Unmarshal(v, &sortOrder); err != nil {
			return fmt.Errorf("sort_order: %w", err)
		}
		a.SortOrder = strings.TrimSpace(sortOrder)
	}
	return nil
}

func parseOptionalInt(raw json.RawMessage) (*int, error) {
	var n int
	if err := json.Unmarshal(raw, &n); err == nil {
		return &n, nil
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return nil, err
	}
	parsed, err := strconv.Atoi(s)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func (t *PokemonDBTool) Execute(ctx context.Context, arguments json.RawMessage) (string, error) {
	_ = ctx

	var args pokemonDBArgs
	if len(arguments) > 0 {
		if err := json.Unmarshal(arguments, &args); err != nil {
			return "", fmt.Errorf("invalid arguments: %w", err)
		}
	}

	sortBy, err := pokemonstore.ParseSortBy(args.SortBy)
	if err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}
	sortDesc, err := pokemonstore.ParseSortOrder(args.SortOrder)
	if err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}

	team, err := t.store.SearchTeam(pokemonstore.TeamFilter{
		PrimaryType:  args.PrimaryType,
		MinLevel:     args.MinLevel,
		MaxLevel:     args.MaxLevel,
		Level:        args.Level,
		Name:         args.Name,
		DexID:        args.DexID,
		CaughtDate:   args.CaughtDate,
		CaughtAfter:  args.CaughtAfter,
		CaughtBefore: args.CaughtBefore,
		SortBy:       sortBy,
		SortDesc:     sortDesc,
	})
	if err != nil {
		return "", err
	}

	offset := 0
	if args.Offset != nil {
		offset = *args.Offset
	}
	if offset < 0 {
		offset = 0
	}

	limit := len(team) - offset
	if args.Limit != nil {
		limit = *args.Limit
	}
	if limit < 0 {
		limit = 0
	}

	sliced := sliceTeam(team, offset, limit)
	data, err := json.Marshal(map[string]any{
		"team":   sliced,
		"total":  len(team),
		"offset": offset,
		"limit":  limit,
		"count":  len(sliced),
		"filters": map[string]any{
			"name":         args.Name,
			"dexId":        args.DexID,
			"primaryType":  args.PrimaryType,
			"level":        args.Level,
			"minLevel":     args.MinLevel,
			"maxLevel":     args.MaxLevel,
			"caughtDate":   args.CaughtDate,
			"caughtAfter":  args.CaughtAfter,
			"caughtBefore": args.CaughtBefore,
			"sortBy":       sortBy,
			"sortOrder":    sortOrderLabel(sortDesc),
		},
	})
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func sortOrderLabel(desc bool) string {
	if desc {
		return "desc"
	}
	return "asc"
}

func sliceTeam(team []domain.TeamPokemon, offset, limit int) []domain.TeamPokemon {
	if offset >= len(team) {
		return []domain.TeamPokemon{}
	}
	team = team[offset:]
	if limit > len(team) {
		limit = len(team)
	}
	if limit <= 0 {
		return []domain.TeamPokemon{}
	}
	return append([]domain.TeamPokemon(nil), team[:limit]...)
}
