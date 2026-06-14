package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/tmc/langchaingo/llms"
)

type Tool interface {
	Name() string
	Description() string
	Parameters() map[string]any
	Execute(ctx context.Context, arguments json.RawMessage) (string, error)
}

type Registry struct {
	tools map[string]Tool
	order []string
}

func NewRegistry(tools ...Tool) *Registry {
	r := &Registry{
		tools: make(map[string]Tool, len(tools)),
		order: make([]string, 0, len(tools)),
	}
	for _, tool := range tools {
		r.Register(tool)
	}
	return r
}

func (r *Registry) Register(tool Tool) {
	if tool == nil {
		return
	}
	name := tool.Name()
	if _, exists := r.tools[name]; !exists {
		r.order = append(r.order, name)
	}
	r.tools[name] = tool
}

func (r *Registry) Get(name string) (Tool, bool) {
	tool, ok := r.tools[name]
	return tool, ok
}

func (r *Registry) Names() []string {
	names := append([]string(nil), r.order...)
	sort.Strings(names)
	return names
}

func (r *Registry) Definitions() []llms.Tool {
	defs := make([]llms.Tool, 0, len(r.order))
	for _, name := range r.order {
		tool := r.tools[name]
		defs = append(defs, llms.Tool{
			Type: "function",
			Function: &llms.FunctionDefinition{
				Name:        tool.Name(),
				Description: tool.Description(),
				Parameters:  tool.Parameters(),
			},
		})
	}
	return defs
}

func (r *Registry) Execute(ctx context.Context, name, arguments string) (string, error) {
	tool, ok := r.Get(name)
	if !ok {
		return "", fmt.Errorf("unknown tool %q", name)
	}
	return tool.Execute(ctx, json.RawMessage(arguments))
}
