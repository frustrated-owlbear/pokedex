package agent

import (
	"time"

	"github.com/google/uuid"
)

type StepKind string

const (
	StepEvent        StepKind = "Event"
	StepThought      StepKind = "Thought"
	StepAction       StepKind = "Action"
	StepObservation  StepKind = "Observation"
	StepFinalAnswer  StepKind = "FinalAnswer"
)

type TraceStep struct {
	ID        string   `json:"id"`
	Kind      StepKind `json:"kind"`
	Timestamp string   `json:"timestamp"`
	Title     string   `json:"title"`
	Detail    string   `json:"detail,omitempty"`
	Tools     []string `json:"tools,omitempty"`
}

type TraceLogger interface {
	Log(step TraceStep)
}

func NewTraceStep(kind StepKind, title, detail string, tools ...string) TraceStep {
	return TraceStep{
		ID:        uuid.NewString(),
		Kind:      kind,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Title:     title,
		Detail:    detail,
		Tools:     tools,
	}
}

func truncateDetail(s string, max int) string {
	if max <= 0 || len(s) <= max {
		return s
	}
	return s[:max] + "…"
}
