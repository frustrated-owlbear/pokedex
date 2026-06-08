package llm

import (
	"context"
	"fmt"
	"strings"

	"github.com/tmc/langchaingo/llms"
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
	_, err := llms.GenerateFromSinglePrompt(
		ctx,
		c.model,
		prompt,
		llms.WithTemperature(c.settings.Temperature),
		llms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
			onChunk(string(chunk))
			return nil
		}),
	)

	return err
}
