package llm

import (
	"errors"
	"strings"
)

var ErrToolsNotSupported = errors.New("ollama model does not support native tool calling")

func IsToolsNotSupported(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, ErrToolsNotSupported) {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "does not support tools")
}
