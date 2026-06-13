package llm

import (
	"errors"
	"testing"
)

func TestIsToolsNotSupported(t *testing.T) {
	t.Parallel()

	if IsToolsNotSupported(nil) {
		t.Fatal("nil error should be false")
	}
	if !IsToolsNotSupported(ErrToolsNotSupported) {
		t.Fatal("expected wrapped sentinel")
	}
	if !IsToolsNotSupported(errors.New(`{"error":"gemma3:latest does not support tools"}`)) {
		t.Fatal("expected substring match")
	}
}
