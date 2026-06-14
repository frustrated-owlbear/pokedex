package agent

import (
	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/rag"
	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/session"
	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/simulation"
	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/tools"
	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/pokemonstore"
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
		tools.NewObservationTool(deps.Sessions),
		tools.NewSessionMemoryTool(deps.Sessions, 3),
		tools.NewKnowledgeSearchTool(deps.Retriever, topK),
	)
}
