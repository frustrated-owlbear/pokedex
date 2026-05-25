package main

import (
	"context"
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

func promptQuestion(question string) string {
	q := strings.TrimSpace(question)
	if q == "" {
		q = "Who was the first pokemon discovered?"
	}
	return fmt.Sprintf("Human: %s\nAssistant:", q)
}

func streamCompletion(ctx context.Context, prompt string, onChunk func(string)) error {
	model, err := ollama.New(ollama.WithModel(ollamaModelName))
	if err != nil {
		return err
	}

	_, err = llms.GenerateFromSinglePrompt(
		ctx,
		model,
		prompt,
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
