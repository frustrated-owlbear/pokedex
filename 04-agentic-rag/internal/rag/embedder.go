package rag

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms/ollama"
)

type Embedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
	ModelName() string
}

type OllamaEmbedder struct {
	model *ollama.LLM
	name  string
}

func NewEmbedder(baseURL, modelName string) (*OllamaEmbedder, error) {
	opts := []ollama.Option{
		ollama.WithModel(modelName),
		ollama.WithHTTPClient(&http.Client{Timeout: 30 * time.Second}),
	}
	if baseURL != "" {
		opts = append(opts, ollama.WithServerURL(baseURL))
	}

	model, err := ollama.New(opts...)
	if err != nil {
		return nil, err
	}
	return &OllamaEmbedder{model: model, name: modelName}, nil
}

func (e *OllamaEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("empty embed input")
	}
	vectors, err := e.model.CreateEmbedding(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(vectors) == 0 {
		return nil, fmt.Errorf("no embedding returned")
	}
	return vectors[0], nil
}

func (e *OllamaEmbedder) ModelName() string {
	return e.name
}
