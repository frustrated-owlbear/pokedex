package llm

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func TestProbeHealthy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"models":[{"name":"gemma3:latest"}]}`))
	}))
	defer server.Close()

	result := Probe(context.Background(), server.URL, "gemma3:latest", server.Client())

	if !result.Reachable {
		t.Fatalf("expected Ollama to be reachable")
	}
	if !result.ModelAvailable {
		t.Fatalf("expected gemma3:latest to be available")
	}
	if result.Detail != "" {
		t.Fatalf("expected no detail for healthy probe, got %q", result.Detail)
	}
}

func TestProbeModelMissing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"models":[{"name":"llama3:latest"}]}`))
	}))
	defer server.Close()

	result := Probe(context.Background(), server.URL, "gemma3:latest", server.Client())

	if !result.Reachable {
		t.Fatalf("expected Ollama to be reachable")
	}
	if result.ModelAvailable {
		t.Fatalf("expected gemma3:latest to be missing")
	}
	if !strings.Contains(result.Detail, "not installed") {
		t.Fatalf("expected model-missing detail, got %q", result.Detail)
	}
}

func TestProbeTimeout(t *testing.T) {
	client := &http.Client{
		Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, context.DeadlineExceeded
		}),
	}

	result := Probe(context.Background(), "http://ollama.local", "gemma3:latest", client)

	if result.Reachable {
		t.Fatalf("expected timeout probe to be unreachable")
	}
	if result.ModelAvailable {
		t.Fatalf("expected no model availability on timeout")
	}
	if result.Detail != "Timed out contacting Ollama" {
		t.Fatalf("unexpected timeout detail: %q", result.Detail)
	}
}
