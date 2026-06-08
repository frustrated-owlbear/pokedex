package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

// Prompt formats a user question for the assistant turn.
func Prompt(question string) string {
	q := strings.TrimSpace(question)
	if q == "" {
		q = "Who was the first pokemon discovered?"
	}
	return fmt.Sprintf("Human: %s\nAssistant:", q)
}

// StreamCompletion streams tokens from Ollama and invokes onChunk for each piece.
func (c *Client) StreamCompletion(ctx context.Context, prompt string, onChunk func(string)) error {
	opts := []ollama.Option{
		ollama.WithModel(c.settings.ModelName),
	}
	if c.settings.BaseURL != "" {
		opts = append(opts, ollama.WithServerURL(c.settings.BaseURL))
	}

	model, err := ollama.New(opts...)
	if err != nil {
		return err
	}

	_, err = llms.GenerateFromSinglePrompt(
		ctx,
		model,
		prompt,
		llms.WithTemperature(c.settings.Temperature),
		llms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
			onChunk(string(chunk))
			return nil
		}),
	)

	return err
}
