package main

import (
	"context"

	"github.com/frustrated-owlbear/pokedex/internal/llm"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx context.Context
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// AnswerQuestion streams an LLM reply; chunks are emitted as "llm:chunk" events.
func (a *App) AnswerQuestion(question string) error {
	prompt := llm.Prompt(question)
	return llm.StreamCompletion(a.ctx, prompt, func(chunk string) {
		runtime.EventsEmit(a.ctx, "llm:chunk", chunk)
	})
}
