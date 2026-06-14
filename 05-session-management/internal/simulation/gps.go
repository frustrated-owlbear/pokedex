package simulation

import "sync"

type GPS struct {
	mu       sync.RWMutex
	location string
	region   string
}

func NewGPS() *GPS {
	return &GPS{
		location: "Viridian Forest",
		region:   "Kanto",
	}
}

type GPSSnapshot struct {
	Location string `json:"location"`
	Region   string `json:"region"`
}

func (g *GPS) Snapshot() GPSSnapshot {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return GPSSnapshot{
		Location: g.location,
		Region:   g.region,
	}
}

func (g *GPS) SetLocation(location string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.location = location
}
