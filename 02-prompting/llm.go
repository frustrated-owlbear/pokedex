package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/ollama"
)

const (
	ollamaModelName      = "gemma3:latest"
	ollamaDefaultBaseURL = "http://127.0.0.1:11434"
	ollamaHealthTimeout  = 2 * time.Second
	maxImageBytes        = 8 << 20
)

type ollamaProbeResult struct {
	Reachable      bool
	ModelAvailable bool
	Detail         string
}

type ollamaTagsResponse struct {
	Models []ollamaModelInfo `json:"models"`
}

type ollamaModelInfo struct {
	Name  string `json:"name"`
	Model string `json:"model"`
}

const pokedexSystemPrompt = `You are a Pokédex from the Kanto region, as in Season 1 of the Pokémon anime.

Answer like the handheld device: brief, clear, and matter-of-fact.
- Keep answers simple and concise (usually one to three sentences).
- Use only knowledge that fits Kanto and Season 1. Do not mention later regions, seasons, or games unless they would plausibly exist in Kanto at that time.
- Do not ask follow-up questions or offer extra help.
- When the trainer provides an image, describe what you see in Kanto Season 1 terms.
- Give the answer and stop.`

var (
	errEmptyInput   = errors.New("prompt or image is required")
	errInvalidImage = errors.New("invalid image data")
)

func buildMessages(question, imageBase64, imageMIME string) ([]llms.MessageContent, error) {
	q := strings.TrimSpace(question)
	imageData, mime, err := decodeImageInput(imageBase64, imageMIME)
	if err != nil {
		return nil, err
	}
	if q == "" && len(imageData) == 0 {
		return nil, errEmptyInput
	}

	messages := []llms.MessageContent{
		{
			Role:  llms.ChatMessageTypeSystem,
			Parts: []llms.ContentPart{llms.TextPart(pokedexSystemPrompt)},
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
		return nil, "", fmt.Errorf("%w: %w", errInvalidImage, err)
	}
	if len(data) == 0 {
		return nil, "", fmt.Errorf("%w: empty payload", errInvalidImage)
	}
	if len(data) > maxImageBytes {
		return nil, "", fmt.Errorf("%w: image exceeds 8 MB", errInvalidImage)
	}

	mime := strings.TrimSpace(imageMIME)
	if mime == "" {
		mime = http.DetectContentType(data)
	}
	if !strings.HasPrefix(mime, "image/") {
		return nil, "", fmt.Errorf("%w: unsupported type %s", errInvalidImage, mime)
	}

	return data, mime, nil
}

func streamChat(ctx context.Context, messages []llms.MessageContent, onChunk func(string)) error {
	model, err := ollama.New(ollama.WithModel(ollamaModelName))
	if err != nil {
		return err
	}

	_, err = model.GenerateContent(
		ctx,
		messages,
		llms.WithTemperature(0.8),
		llms.WithStreamingFunc(func(_ context.Context, chunk []byte) error {
			onChunk(string(chunk))
			return nil
		}),
	)

	return err
}

func checkOllamaHealth(ctx context.Context) ollamaProbeResult {
	checkCtx, cancel := context.WithTimeout(ctx, ollamaHealthTimeout)
	defer cancel()

	return probeOllama(
		checkCtx,
		resolveOllamaBaseURL(),
		ollamaModelName,
		&http.Client{},
	)
}

func probeOllama(
	ctx context.Context,
	baseURL string,
	modelName string,
	client *http.Client,
) ollamaProbeResult {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		baseURL+"/api/tags",
		nil,
	)
	if err != nil {
		return ollamaProbeResult{Detail: "Invalid Ollama URL"}
	}

	resp, err := client.Do(req)
	if err != nil {
		return ollamaProbeResult{Detail: formatOllamaProbeError(err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ollamaProbeResult{
			Detail: fmt.Sprintf("Ollama returned HTTP %d", resp.StatusCode),
		}
	}

	var tags ollamaTagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return ollamaProbeResult{Detail: "Invalid response from Ollama"}
	}

	for _, model := range tags.Models {
		if model.Name == modelName || model.Model == modelName {
			return ollamaProbeResult{
				Reachable:      true,
				ModelAvailable: true,
			}
		}
	}

	return ollamaProbeResult{
		Reachable: true,
		Detail:    modelName + " not installed",
	}
}

func resolveOllamaBaseURL() string {
	host := strings.TrimSpace(os.Getenv("OLLAMA_HOST"))
	if host == "" {
		return ollamaDefaultBaseURL
	}
	if !strings.Contains(host, "://") {
		host = "http://" + host
	}
	return strings.TrimRight(host, "/")
}

func formatOllamaProbeError(err error) string {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return "Timed out contacting Ollama"
	default:
		return err.Error()
	}
}
