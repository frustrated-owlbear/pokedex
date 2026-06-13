package agent

import (
	"github.com/frustrated-owlbear/pokedex/04-agentic-rag/internal/rag"
	"github.com/frustrated-owlbear/pokedex/04-agentic-rag/internal/session"
	"github.com/frustrated-owlbear/pokedex/04-agentic-rag/internal/simulation"
	"github.com/frustrated-owlbear/pokedex/04-agentic-rag/internal/tools"
	"github.com/frustrated-owlbear/pokedex/04-agentic-rag/internal/pokemonstore"
)

type Dependencies struct {
	TeamStore *pokemonstore.SQLiteStore
	Clock     *simulation.Clock
	GPS       *simulation.GPS
	Sessions  *session.Store
	Retriever *rag.Retriever
	RAGTopK   int
}

func NewDefaultRegistry(deps Dependencies) *Registry {
	topK := deps.RAGTopK
	return NewRegistry(
		tools.NewClockTool(deps.Clock),
		tools.NewGPSTool(deps.GPS),
		tools.NewPokemonDBTool(deps.TeamStore),
		tools.NewSessionMemoryTool(deps.Sessions, 3),
		tools.NewKnowledgeSearchTool(deps.Retriever, topK),
	)
}
