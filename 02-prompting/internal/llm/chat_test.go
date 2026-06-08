package llm

import (
	"encoding/base64"
	"errors"
	"testing"
)

func TestBuildMessagesEmpty(t *testing.T) {
	_, err := BuildMessages("   ", "", "")
	if !errors.Is(err, ErrEmptyInput) {
		t.Fatalf("expected ErrEmptyInput, got %v", err)
	}
}

func TestBuildMessagesTextOnly(t *testing.T) {
	messages, err := BuildMessages("What is Pikachu?", "", "")
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

	messages, err := BuildMessages("", encoded, "image/png")
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

	messages, err := BuildMessages("What is this?", encoded, "image/png")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(messages[1].Parts) != 2 {
		t.Fatalf("expected image and text parts, got %d", len(messages[1].Parts))
	}
}
