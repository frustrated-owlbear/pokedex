package main

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

func TestProbeOllamaHealthy(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"models":[{"name":"gemma3:latest"}]}`))
	}))
	defer server.Close()

	result := probeOllama(context.Background(), server.URL, ollamaModelName, server.Client())

	if !result.Reachable {
		t.Fatalf("expected Ollama to be reachable")
	}
	if !result.ModelAvailable {
		t.Fatalf("expected %s to be available", ollamaModelName)
	}
	if result.Detail != "" {
		t.Fatalf("expected no detail for healthy probe, got %q", result.Detail)
	}
}

func TestProbeOllamaModelMissing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"models":[{"name":"llama3:latest"}]}`))
	}))
	defer server.Close()

	result := probeOllama(context.Background(), server.URL, ollamaModelName, server.Client())

	if !result.Reachable {
		t.Fatalf("expected Ollama to be reachable")
	}
	if result.ModelAvailable {
		t.Fatalf("expected %s to be missing", ollamaModelName)
	}
	if !strings.Contains(result.Detail, "not installed") {
		t.Fatalf("expected model-missing detail, got %q", result.Detail)
	}
}

func TestProbeOllamaTimeout(t *testing.T) {
	client := &http.Client{
		Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return nil, context.DeadlineExceeded
		}),
	}

	result := probeOllama(context.Background(), "http://ollama.local", ollamaModelName, client)

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
