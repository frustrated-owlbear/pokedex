package rag

import (
	"math"
	"sort"
	"sync"
)

type Document struct {
	ID      string
	Source  string
	Content string
	Vector  []float32
}

type Store struct {
	mu   sync.RWMutex
	docs []Document
}

func NewStore() *Store {
	return &Store{}
}

func (s *Store) Add(doc Document) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.docs = append(s.docs, doc)
}

func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.docs)
}

type ScoredDocument struct {
	Document
	Score float64
}

func (s *Store) Search(vector []float32, topK int) []ScoredDocument {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if topK <= 0 {
		topK = 4
	}

	scored := make([]ScoredDocument, 0, len(s.docs))
	for _, doc := range s.docs {
		scored = append(scored, ScoredDocument{
			Document: doc,
			Score:    CosineSimilarity(vector, doc.Vector),
		})
	}

	sort.Slice(scored, func(i, j int) bool {
		return scored[i].Score > scored[j].Score
	})

	if len(scored) > topK {
		scored = scored[:topK]
	}
	return scored
}

func CosineSimilarity(a, b []float32) float64 {
	if len(a) == 0 || len(b) == 0 || len(a) != len(b) {
		return 0
	}
	var dot, normA, normB float64
	for i := range a {
		dot += float64(a[i]) * float64(b[i])
		normA += float64(a[i]) * float64(a[i])
		normB += float64(b[i]) * float64(b[i])
	}
	if normA == 0 || normB == 0 {
		return 0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}
