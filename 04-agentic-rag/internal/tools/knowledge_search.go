package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/frustrated-owlbear/pokedex/04-agentic-rag/internal/rag"
)

type KnowledgeSearchTool struct {
	retriever *rag.Retriever
	topK      int
}

func NewKnowledgeSearchTool(retriever *rag.Retriever, topK int) *KnowledgeSearchTool {
	if topK <= 0 {
		topK = 4
	}
	return &KnowledgeSearchTool{retriever: retriever, topK: topK}
}

func (t *KnowledgeSearchTool) Name() string { return "knowledge_search" }

func (t *KnowledgeSearchTool) Description() string {
	return "Searches the Pokédex knowledge base for Pokémon facts, locations, and Kanto Season 1 lore."
}

func (t *KnowledgeSearchTool) Parameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"query": map[string]any{
				"type":        "string",
				"description": "Search query for Pokémon knowledge retrieval",
			},
		},
		"required": []string{"query"},
	}
}

type knowledgeSearchArgs struct {
	Query string `json:"query"`
}

func (t *KnowledgeSearchTool) Execute(ctx context.Context, arguments json.RawMessage) (string, error) {
	var args knowledgeSearchArgs
	if err := json.Unmarshal(arguments, &args); err != nil {
		return "", fmt.Errorf("invalid arguments: %w", err)
	}
	results, err := t.retriever.Search(ctx, args.Query, t.topK)
	if err != nil {
		return "", err
	}
	return rag.FormatResults(results), nil
}
