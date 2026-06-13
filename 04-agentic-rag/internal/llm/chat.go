package llm

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/frustrated-owlbear/pokedex/04-agentic-rag/internal/domain"
	"github.com/tmc/langchaingo/llms"
)

const maxImageBytes = 8 << 20

const pokedexSystemPrompt = `You are a Pokédex from the Kanto region, as in Season 1 of the Pokémon anime.

Answer like the handheld device: brief, clear, and matter-of-fact.
- Keep answers simple and concise (usually one to three sentences).
- Use only knowledge that fits Kanto and Season 1. Do not mention later regions, seasons, or games unless they would plausibly exist in Kanto at that time.
- Do not ask follow-up questions or offer extra help.
- When the trainer provides an image, describe what you see in Kanto Season 1 terms.
- Give the answer and stop.`

var (
	ErrEmptyInput   = errors.New("prompt or image is required")
	ErrInvalidImage = errors.New("invalid image data")
)

// BuildMessages assembles system and human messages for a text and/or image prompt.
// team is the trainer's owned Pokémon, included in the system prompt when non-empty.
func BuildMessages(question, imageBase64, imageMIME string, team []domain.TeamPokemon) ([]llms.MessageContent, error) {
	q := strings.TrimSpace(question)
	imageData, mime, err := decodeImageInput(imageBase64, imageMIME)
	if err != nil {
		return nil, err
	}
	if q == "" && len(imageData) == 0 {
		return nil, ErrEmptyInput
	}

	systemPrompt := pokedexSystemPrompt
	if teamContext := formatTeamContext(team); teamContext != "" {
		systemPrompt += "\n\n" + teamContext
	}

	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextPart(systemPrompt)},
		},
	}

	var humanParts []llms.ContentPart
	if len(imageData) > 0 {
		humanParts = append(humanParts, llms.BinaryPart(mime, imageData))
	}
	if q != "" {
		humanParts = append(humanParts, llms.TextPart(q))
	}

	messages = append(messages, llms.MessageContent{
		Role:  llms.ChatMessageTypeHuman,
		Parts: humanParts,
	})

	return messages, nil
}

// DecodeImageInput validates optional base64 image input for agent and chat paths.
func DecodeImageInput(imageBase64, imageMIME string) ([]byte, string, error) {
	return decodeImageInput(imageBase64, imageMIME)
}

func decodeImageInput(imageBase64, imageMIME string) ([]byte, string, error) {
	encoded := strings.TrimSpace(imageBase64)
	if encoded == "" {
		return nil, "", nil
	}

	if idx := strings.Index(encoded, ","); idx >= 0 {
		encoded = encoded[idx+1:]
	}

	data, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, "", fmt.Errorf("%w: %w", ErrInvalidImage, err)
	}
	if len(data) == 0 {
		return nil, "", fmt.Errorf("%w: empty payload", ErrInvalidImage)
	}
	if len(data) > maxImageBytes {
		return nil, "", fmt.Errorf("%w: image exceeds 8 MB", ErrInvalidImage)
	}

	mime := strings.TrimSpace(imageMIME)
	if mime == "" {
		mime = http.DetectContentType(data)
	}
	if !strings.HasPrefix(mime, "image/") {
		return nil, "", fmt.Errorf("%w: unsupported type %s", ErrInvalidImage, mime)
	}

	return data, mime, nil
}

func formatTeamContext(team []domain.TeamPokemon) string {
	if len(team) == 0 {
		return "The trainer currently owns no Pokémon."
	}

	var b strings.Builder
	b.WriteString("The trainer's owned Pokémon:")
	for i, pokemon := range team {
		b.WriteString(fmt.Sprintf(
			"\n%d. %s (Lv. %d, %s, %d/%d HP, caught %s",
			i+1,
			pokemon.Name,
			pokemon.Level,
			pokemon.PrimaryType,
			pokemon.HP,
			pokemon.MaxHP,
			pokemon.CaughtDate,
		))
		if pokemon.Birthday != "" {
			b.WriteString(fmt.Sprintf(", birthday %s", pokemon.Birthday))
		}
		b.WriteByte(')')
	}
	b.WriteString("\nUse this list when the trainer asks about their team or party.")

	return b.String()
}

// StreamChat streams tokens from Ollama for a chat message list.
func (c *Client) StreamChat(ctx context.Context, messages []llms.MessageContent, onChunk func(string)) error {
	_, err := c.model.GenerateContent(
		ctx,
		messages,
		llms.WithTemperature(c.settings.Temperature),
		llms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
			onChunk(string(chunk))
			return nil
		}),
	)

	return err
}
