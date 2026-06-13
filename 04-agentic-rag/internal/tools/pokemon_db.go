package tools

import (
	"context"
	"encoding/json"

	"github.com/frustrated-owlbear/pokedex/04-agentic-rag/internal/pokemonstore"
)

type PokemonDBTool struct {
	store *pokemonstore.SQLiteStore
}

func NewPokemonDBTool(store *pokemonstore.SQLiteStore) *PokemonDBTool {
	return &PokemonDBTool{store: store}
}

func (t *PokemonDBTool) Name() string { return "pokemon_db" }

func (t *PokemonDBTool) Description() string {
	return "Returns Pokémon currently owned by the trainer, including levels, types, and HP."
}

func (t *PokemonDBTool) Parameters() map[string]any {
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
}

func (t *PokemonDBTool) Execute(ctx context.Context, _ json.RawMessage) (string, error) {
	_ = ctx
	team, err := t.store.ListTeam()
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(map[string]any{"team": team})
	if err != nil {
		return "", err
	}
	return string(data), nil
}
