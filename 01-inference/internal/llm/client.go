package llm

import (
	"context"
	"fmt"

	"github.com/tmc/langchaingo/llms"
)

// Prompt formats a user question for the assistant turn.
func Prompt(question string) string {
	return fmt.Sprintf("Human: %s\nAssistant:", question)
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
