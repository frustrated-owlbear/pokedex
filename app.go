package main

import (
	"context"
	"path/filepath"

	"github.com/frustrated-owlbear/pokedex/internal/domain"
	"github.com/frustrated-owlbear/pokedex/internal/llm"
	"github.com/frustrated-owlbear/pokedex/internal/pokemonstore"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx          context.Context
	pokemonStore *pokemonstore.SQLiteStore
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	store, err := pokemonstore.NewSQLiteStore(filepath.Join(".", "pokedex.db"))
	if err != nil {
		panic(err)
	}
	a.pokemonStore = store
}

func (a *App) shutdown(ctx context.Context) {
	if a.pokemonStore == nil {
		return
	}
	if err := a.pokemonStore.Close(); err != nil {
		runtime.LogError(ctx, err.Error())
	}
}

// AnswerQuestion streams an LLM reply; chunks are emitted as "llm:chunk" events.
func (a *App) AnswerQuestion(question string) error {
	prompt := llm.Prompt(question)
	return llm.StreamCompletion(a.ctx, prompt, func(chunk string) {
		runtime.EventsEmit(a.ctx, "llm:chunk", chunk)
	})
}

// MyPokemons returns a hardcoded list of favourite pokemons.
func (a *App) MyPokemons() []domain.Pokemon {
	if a.pokemonStore == nil {
		return []domain.Pokemon{}
	}

	pokemons, err := a.pokemonStore.ListPokemons()
	if err != nil {
		runtime.LogError(a.ctx, err.Error())
		return []domain.Pokemon{}
	}

	return pokemons
}
