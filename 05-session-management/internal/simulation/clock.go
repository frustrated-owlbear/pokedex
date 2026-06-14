package simulation

import (
	"sync"
	"time"
)

type Clock struct {
	mu      sync.RWMutex
	started time.Time
	offset  time.Duration
}

func NewClock() *Clock {
	return &Clock{
		started: time.Now(),
		offset:  6 * time.Hour, // start at 06:00 in-game
	}
}

type ClockSnapshot struct {
	Time       string `json:"time"`
	Period     string `json:"period"`
	Weather    string `json:"weather"`
	InGameHour int    `json:"inGameHour"`
}

func (c *Clock) Snapshot() ClockSnapshot {
	c.mu.RLock()
	defer c.mu.RUnlock()

	elapsed := time.Since(c.started) + c.offset
	minutes := int(elapsed.Minutes()) % (24 * 60)
	hour := minutes / 60
	minute := minutes % 60

	period := "Morning"
	switch {
	case hour >= 5 && hour < 12:
		period = "Morning"
	case hour >= 12 && hour < 17:
		period = "Afternoon"
	case hour >= 17 && hour < 21:
		period = "Evening"
	default:
		period = "Night"
	}

	weather := "Clear"
	if hour >= 18 || hour < 6 {
		weather = "Clear"
	}

	return ClockSnapshot{
		Time:       formatGameTime(hour, minute),
		Period:     period,
		Weather:    weather,
		InGameHour: hour,
	}
}

func formatGameTime(hour, minute int) string {
	return time.Date(1997, 4, 1, hour, minute, 0, 0, time.UTC).Format("15:04")
}
