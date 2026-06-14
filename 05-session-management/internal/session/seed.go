package session

import (
	"context"
	"database/sql"
	"time"

	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/domain"
)

type seedSession struct {
	id        string
	startedAt time.Time
	endedAt   time.Time
	summary   string
	events    []domain.Observation
}

func seedSessions() []seedSession {
	return []seedSession{
		{
			id:        "session-viridian-pewter",
			startedAt: time.Date(2024, 2, 15, 9, 0, 0, 0, time.UTC),
			endedAt:   time.Date(2024, 2, 20, 18, 0, 0, 0, time.UTC),
			summary:   "Early Kanto journey from Pallet Town through Viridian Forest. Bulbasaur is the go-to lead. Trainer is cautious about item use before gyms.",
			events: []domain.Observation{
				{Timestamp: time.Date(2024, 2, 15, 9, 30, 0, 0, time.UTC), Category: domain.ObservationCapture, Content: "Chose Bulbasaur as starter in Pallet Town"},
				{Timestamp: time.Date(2024, 2, 16, 11, 0, 0, 0, time.UTC), Category: domain.ObservationLocation, Content: "Traveled through Route 1 and Viridian Forest"},
				{Timestamp: time.Date(2024, 2, 17, 14, 0, 0, 0, time.UTC), Category: domain.ObservationBattle, Content: "Defeated wild Spearow with Bulbasaur's Vine Whip"},
				{Timestamp: time.Date(2024, 2, 18, 10, 0, 0, 0, time.UTC), Category: domain.ObservationPreference, Content: "Trainer prefers leading with Bulbasaur in early battles"},
				{Timestamp: time.Date(2024, 2, 19, 16, 0, 0, 0, time.UTC), Category: domain.ObservationNote, Content: "Stocked up on Potions before Pewter City"},
			},
		},
		{
			id:        "session-pewter-vermilion",
			startedAt: time.Date(2024, 3, 1, 9, 0, 0, 0, time.UTC),
			endedAt:   time.Date(2024, 3, 10, 20, 0, 0, 0, time.UTC),
			summary:   "Mid-Kanto arc through Pewter and Vermilion Gyms. Trainer habitually leads with Bulbasaur even vs Electric-types, favoring Leech Seed and status play. Earned Boulder and Thunder Badges.",
			events: []domain.Observation{
				{Timestamp: time.Date(2024, 3, 1, 10, 0, 0, 0, time.UTC), Category: domain.ObservationCapture, Content: "Caught Pidgey on Route 1"},
				{Timestamp: time.Date(2024, 3, 3, 15, 0, 0, 0, time.UTC), Category: domain.ObservationBadge, Content: "Earned Boulder Badge at Pewter Gym with Bulbasaur"},
				{Timestamp: time.Date(2024, 3, 6, 13, 0, 0, 0, time.UTC), Category: domain.ObservationBattle, Content: "Fought Lt. Surge's Voltorb; used Bulbasaur's Leech Seed despite Electric disadvantage"},
				{Timestamp: time.Date(2024, 3, 7, 11, 0, 0, 0, time.UTC), Category: domain.ObservationBattle, Content: "Repeated Bulbasaur leads against Electric-types at Vermilion Gym"},
				{Timestamp: time.Date(2024, 3, 7, 12, 0, 0, 0, time.UTC), Category: domain.ObservationPreference, Content: "Usually relies on Bulbasaur against Electric opponents, using status moves over raw damage"},
				{Timestamp: time.Date(2024, 3, 8, 9, 0, 0, 0, time.UTC), Category: domain.ObservationLocation, Content: "Visited Vermilion City and the SS Anne"},
				{Timestamp: time.Date(2024, 3, 10, 19, 0, 0, 0, time.UTC), Category: domain.ObservationBadge, Content: "Earned Thunder Badge"},
			},
		},
	}
}

func (s *Store) seedIfEmpty(ctx context.Context) error {
	var count int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM sessions WHERE ended_at IS NOT NULL`).Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	for _, seed := range seedSessions() {
		if err := s.insertSeedSession(seed); err != nil {
			return err
		}
	}
	return s.rebuildIndex(ctx)
}

func (s *Store) insertSeedSession(seed seedSession) error {
	_, err := s.db.Exec(
		`INSERT OR IGNORE INTO sessions (id, started_at, ended_at, summary) VALUES (?, ?, ?, ?)`,
		seed.id,
		seed.startedAt.Format(time.RFC3339),
		seed.endedAt.Format(time.RFC3339),
		seed.summary,
	)
	if err != nil {
		return err
	}

	for _, event := range seed.events {
		_, err := s.db.Exec(
			`INSERT INTO observations (session_id, timestamp, event_type, category, content) VALUES (?, ?, ?, ?, ?)`,
			seed.id,
			event.Timestamp.Format(time.RFC3339),
			string(domain.EventObservation),
			event.Category,
			event.Content,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func insertSeedSessionTx(tx *sql.Tx, seed seedSession) error {
	_, err := tx.Exec(
		`INSERT OR IGNORE INTO sessions (id, started_at, ended_at, summary) VALUES (?, ?, ?, ?)`,
		seed.id,
		seed.startedAt.Format(time.RFC3339),
		seed.endedAt.Format(time.RFC3339),
		seed.summary,
	)
	if err != nil {
		return err
	}
	for _, event := range seed.events {
		_, err := tx.Exec(
			`INSERT INTO observations (session_id, timestamp, event_type, category, content) VALUES (?, ?, ?, ?, ?)`,
			seed.id,
			event.Timestamp.Format(time.RFC3339),
			string(domain.EventObservation),
			event.Category,
			event.Content,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
