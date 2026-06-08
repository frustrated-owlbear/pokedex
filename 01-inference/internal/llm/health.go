package llm

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type ProbeResult struct {
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

// CheckHealth probes the configured Ollama instance for model availability.
func (c *Client) CheckHealth(ctx context.Context) ProbeResult {
	timeout := c.settings.HealthTimeout
	if timeout <= 0 {
		timeout = defaultHealthTimeout
	}

	checkCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return Probe(
		checkCtx,
		c.settings.BaseURL,
		c.settings.ModelName,
		&http.Client{},
	)
}

// Probe checks whether Ollama is reachable and the given model is installed.
func Probe(
	ctx context.Context,
	baseURL string,
	modelName string,
	client *http.Client,
) ProbeResult {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		baseURL+"/api/tags",
		nil,
	)
	if err != nil {
		return ProbeResult{Detail: "Invalid Ollama URL"}
	}

	resp, err := client.Do(req)
	if err != nil {
		return ProbeResult{Detail: formatOllamaProbeError(err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ProbeResult{
			Detail: fmt.Sprintf("Ollama returned HTTP %d", resp.StatusCode),
		}
	}

	var tags ollamaTagsResponse
	if err := json.NewDecoder(resp.Body).Decode(&tags); err != nil {
		return ProbeResult{Detail: "Invalid response from Ollama"}
	}

	for _, model := range tags.Models {
		if model.Name == modelName || model.Model == modelName {
			return ProbeResult{
				Reachable:      true,
				ModelAvailable: true,
			}
		}
	}

	return ProbeResult{
		Reachable: true,
		Detail:    modelName + " not installed",
	}
}

func formatOllamaProbeError(err error) string {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return "Timed out contacting Ollama"
	default:
		return err.Error()
	}
}
