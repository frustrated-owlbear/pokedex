package main

import (
	"context"
	"encoding/base64"
	"errors"
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

func TestBuildMessagesEmpty(t *testing.T) {
	_, err := buildMessages("   ", "", "")
	if !errors.Is(err, errEmptyInput) {
		t.Fatalf("expected errEmptyInput, got %v", err)
	}
}

func TestBuildMessagesTextOnly(t *testing.T) {
	messages, err := buildMessages("What is Pikachu?", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(messages) != 2 {
		t.Fatalf("expected system and human messages, got %d", len(messages))
	}
	if len(messages[1].Parts) != 1 {
		t.Fatalf("expected one human part, got %d", len(messages[1].Parts))
	}
}

func TestBuildMessagesImageOnly(t *testing.T) {
	png := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	encoded := base64.StdEncoding.EncodeToString(png)

	messages, err := buildMessages("", encoded, "image/png")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(messages[1].Parts) != 1 {
		t.Fatalf("expected one human part for image-only input, got %d", len(messages[1].Parts))
	}
}

func TestBuildMessagesTextAndImage(t *testing.T) {
	png := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	encoded := base64.StdEncoding.EncodeToString(png)

	messages, err := buildMessages("What is this?", encoded, "image/png")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(messages[1].Parts) != 2 {
		t.Fatalf("expected image and text parts, got %d", len(messages[1].Parts))
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
