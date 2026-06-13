package rag

import (
	"context"
	"fmt"
	"strings"
)

type Retriever struct {
	embedder Embedder
	store    *Store
	topK     int
	ready    bool
}

func NewRetriever(embedder Embedder, store *Store, topK int) *Retriever {
	if topK <= 0 {
		topK = 4
	}
	return &Retriever{
		embedder: embedder,
		store:    store,
		topK:     topK,
	}
}

func (r *Retriever) Bootstrap(ctx context.Context) error {
	for _, entry := range KantoCorpus() {
		vector, err := r.embedder.Embed(ctx, entry.Content)
		if err != nil {
			return fmt.Errorf("embed %s: %w", entry.Source, err)
		}
		r.store.Add(Document{
			ID:      entry.Source,
			Source:  entry.Source,
			Content: entry.Content,
			Vector:  vector,
		})
	}
	r.ready = true
	return nil
}

func (r *Retriever) Ready() bool {
	return r.ready && r.store.Len() > 0
}

func (r *Retriever) Search(ctx context.Context, query string, topK int) ([]ScoredDocument, error) {
	if !r.ready {
		return nil, fmt.Errorf("retriever not ready")
	}
	if topK <= 0 {
		topK = r.topK
	}
	vector, err := r.embedder.Embed(ctx, query)
	if err != nil {
		return nil, err
	}
	return r.store.Search(vector, topK), nil
}

func FormatResults(results []ScoredDocument) string {
	if len(results) == 0 {
		return "No matching knowledge entries found."
	}
	var b strings.Builder
	for i, doc := range results {
		fmt.Fprintf(&b, "%d. [%s] (score %.2f) %s\n", i+1, doc.Source, doc.Score, doc.Content)
	}
	return strings.TrimSpace(b.String())
}
