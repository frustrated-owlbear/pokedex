package llm

import (
	"encoding/base64"
	"errors"
	"strings"
	"testing"

	"github.com/frustrated-owlbear/pokedex/04-agentic-rag/internal/domain"
	"github.com/tmc/langchaingo/llms"
)

func TestBuildMessagesEmpty(t *testing.T) {
	_, err := BuildMessages("   ", "", "", nil)
	if !errors.Is(err, ErrEmptyInput) {
		t.Fatalf("expected ErrEmptyInput, got %v", err)
	}
}

func TestBuildMessagesTextOnly(t *testing.T) {
	messages, err := BuildMessages("What is Pikachu?", "", "", nil)
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

	messages, err := BuildMessages("", encoded, "image/png", nil)
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

	messages, err := BuildMessages("What is this?", encoded, "image/png", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(messages[1].Parts) != 2 {
		t.Fatalf("expected image and text parts, got %d", len(messages[1].Parts))
	}
}

func TestBuildMessagesIncludesTeamInSystemPrompt(t *testing.T) {
	team := []domain.TeamPokemon{
		{
			Name:        "Bulbasaur",
			Level:       16,
			PrimaryType: "GRASS",
			HP:          42,
			MaxHP:       42,
			CaughtDate:  "2024-02-15",
			Birthday:    "2024-02-14",
		},
		{
			Name:        "Pidgey",
			Level:       12,
			PrimaryType: "NORMAL",
			HP:          31,
			MaxHP:       31,
			CaughtDate:  "2024-03-01",
		},
	}

	messages, err := BuildMessages("Who is on my team?", "", "", team)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	systemText, ok := messages[0].Parts[0].(llms.TextContent)
	if !ok {
		t.Fatalf("expected text system part")
	}
	if !strings.Contains(systemText.Text, "Bulbasaur") {
		t.Fatalf("system prompt missing Bulbasaur: %q", systemText.Text)
	}
	if !strings.Contains(systemText.Text, "Pidgey") {
		t.Fatalf("system prompt missing Pidgey: %q", systemText.Text)
	}
	if !strings.Contains(systemText.Text, "caught 2024-02-15") {
		t.Fatalf("system prompt missing caught date: %q", systemText.Text)
	}
}
