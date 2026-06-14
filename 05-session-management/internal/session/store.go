package session

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/domain"
	"github.com/frustrated-owlbear/pokedex/05-session-management/internal/rag"
	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	db          *sql.DB
	embedder    rag.Embedder
	summarizer  Summarizer
	vectors     []storedMemory
	activeID    string
	inMemoryDSN bool
}

type storedMemory struct {
	SessionID   string
	Kind        string // summary | observation
	Content     string
	SessionDate string
	Vector      []float32
}

type SessionView struct {
	ID         string `json:"id"`
	StartedAt  string `json:"startedAt"`
	EndedAt    string `json:"endedAt,omitempty"`
	Summary    string `json:"summary,omitempty"`
	EventCount int    `json:"eventCount"`
	Active     bool   `json:"active"`
}

func NewStore(embedder rag.Embedder, summarizer Summarizer) (*Store, error) {
	return NewStoreWithDSN("", embedder, summarizer)
}

// NewStoreWithDSN opens a session store. Empty dsn uses the user config directory.
// Tests may pass an in-memory DSN.
func NewStoreWithDSN(dsn string, embedder rag.Embedder, summarizer Summarizer) (*Store, error) {
	inMemory := false
	if dsn == "" {
		path, err := defaultDBPath()
		if err != nil {
			return nil, err
		}
		dsn = "file:" + path + "?cache=shared&_foreign_keys=on"
	} else {
		inMemory = strings.Contains(dsn, "mode=memory")
	}

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, err
	}
	s := &Store{
		db:          db,
		embedder:    embedder,
		summarizer:  summarizer,
		inMemoryDSN: inMemory,
	}
	if err := s.init(context.Background()); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func defaultDBPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	root := filepath.Join(dir, "pokedex")
	if err := os.MkdirAll(root, 0o755); err != nil {
		return "", err
	}
	return filepath.Join(root, "sessions.db"), nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Ready() bool {
	return s.db != nil
}

func (s *Store) init(ctx context.Context) error {
	if _, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			started_at TEXT NOT NULL,
			ended_at TEXT,
			summary TEXT
		);
		CREATE TABLE IF NOT EXISTS observations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL,
			timestamp TEXT NOT NULL,
			event_type TEXT NOT NULL DEFAULT 'observation',
			category TEXT NOT NULL,
			content TEXT NOT NULL,
			FOREIGN KEY(session_id) REFERENCES sessions(id)
		);
	`); err != nil {
		return err
	}

	if err := s.migrateSchema(); err != nil {
		return err
	}

	if err := s.seedIfEmpty(ctx); err != nil {
		return err
	}
	return s.rebuildIndex(ctx)
}

func (s *Store) migrateSchema() error {
	rows, err := s.db.Query(`PRAGMA table_info(observations)`)
	if err != nil {
		return err
	}
	defer rows.Close()

	hasEventType := false
	for rows.Next() {
		var cid int
		var name, colType string
		var notNull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &colType, &notNull, &dflt, &pk); err != nil {
			return err
		}
		if name == "event_type" {
			hasEventType = true
			break
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if hasEventType {
		return nil
	}
	_, err = s.db.Exec(`ALTER TABLE observations ADD COLUMN event_type TEXT NOT NULL DEFAULT 'observation'`)
	return err
}

func (s *Store) rebuildIndex(ctx context.Context) error {
	s.vectors = nil

	rows, err := s.db.Query(`
		SELECT s.id, s.started_at, s.summary, o.category, o.content
		FROM sessions s
		LEFT JOIN observations o ON o.session_id = s.id
		WHERE s.ended_at IS NOT NULL
		ORDER BY s.started_at ASC, o.timestamp ASC
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	type sessionSummary struct {
		id      string
		started string
		summary string
		added   bool
	}
	seen := make(map[string]*sessionSummary)

	for rows.Next() {
		var sessionID, startedAt, summary, category, content sql.NullString
		if err := rows.Scan(&sessionID, &startedAt, &summary, &category, &content); err != nil {
			return err
		}

		entry, ok := seen[sessionID.String]
		if !ok {
			entry = &sessionSummary{id: sessionID.String, started: startedAt.String, summary: summary.String}
			seen[sessionID.String] = entry
		}

		if summary.Valid && strings.TrimSpace(summary.String) != "" && !entry.added {
			text := fmt.Sprintf("Session %s: %s", formatSessionDate(startedAt.String), summary.String)
			if err := s.indexMemory(ctx, sessionID.String, "summary", text, startedAt.String); err != nil {
				return err
			}
			entry.added = true
		}

		if category.Valid && content.Valid && strings.TrimSpace(content.String) != "" {
			text := fmt.Sprintf("[%s] %s (session %s)", category.String, content.String, formatSessionDate(startedAt.String))
			if err := s.indexMemory(ctx, sessionID.String, "observation", text, startedAt.String); err != nil {
				return err
			}
		}
	}
	return rows.Err()
}

func (s *Store) indexMemory(ctx context.Context, sessionID, kind, content, sessionDate string) error {
	vector, err := s.embedder.Embed(ctx, content)
	if err != nil {
		return err
	}
	s.vectors = append(s.vectors, storedMemory{
		SessionID:   sessionID,
		Kind:        kind,
		Content:     content,
		SessionDate: sessionDate,
		Vector:      vector,
	})
	return nil
}

func (s *Store) EnsureActiveSession(ctx context.Context) (string, error) {
	if s.activeID != "" {
		return s.activeID, nil
	}

	var id string
	err := s.db.QueryRow(`SELECT id FROM sessions WHERE ended_at IS NULL ORDER BY started_at DESC LIMIT 1`).Scan(&id)
	if err == nil {
		s.activeID = id
		return id, nil
	}
	if err != sql.ErrNoRows {
		return "", err
	}

	id = fmt.Sprintf("session-%d", time.Now().UTC().UnixNano())
	now := time.Now().UTC()
	if _, err := s.db.Exec(
		`INSERT INTO sessions (id, started_at) VALUES (?, ?)`,
		id, now.Format(time.RFC3339),
	); err != nil {
		return "", err
	}
	s.activeID = id
	return id, nil
}

func (s *Store) ActiveSessionID() string {
	return s.activeID
}

func (s *Store) AddObservation(ctx context.Context, sessionID, category, content string) error {
	return s.AddEvent(ctx, sessionID, domain.EventObservation, category, content)
}

func (s *Store) AddEvent(ctx context.Context, sessionID string, eventType domain.SessionEventType, category, content string) error {
	content = strings.TrimSpace(content)
	if content == "" {
		return fmt.Errorf("event content is required")
	}
	if !domain.IsValidSessionEventType(eventType) {
		return fmt.Errorf("invalid event type %q", eventType)
	}
	if domain.EventTypeNeedsCategory(eventType) {
		if !domain.IsValidObservationCategory(category) {
			return fmt.Errorf("invalid observation category %q", category)
		}
	} else if category == "" {
		category = string(eventType)
	}
	if sessionID == "" {
		var err error
		sessionID, err = s.EnsureActiveSession(ctx)
		if err != nil {
			return err
		}
	}

	now := time.Now().UTC()
	_, err := s.db.Exec(
		`INSERT INTO observations (session_id, timestamp, event_type, category, content) VALUES (?, ?, ?, ?, ?)`,
		sessionID, now.Format(time.RFC3339), string(eventType), category, content,
	)
	return err
}

func (s *Store) RecentEvents(sessionID string, limit int) ([]domain.SessionEvent, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := s.db.Query(
		`SELECT timestamp, event_type, category, content FROM observations
		 WHERE session_id = ? ORDER BY timestamp DESC LIMIT ?`,
		sessionID, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.SessionEvent
	for rows.Next() {
		event, err := scanSessionEvent(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, event)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Return chronological order (oldest first).
	for i, j := 0, len(result)-1; i < j; i, j = i+1, j-1 {
		result[i], result[j] = result[j], result[i]
	}
	return result, nil
}

func (s *Store) ActiveSessionContext(sessionID string) (string, []domain.SessionEvent, error) {
	events, err := s.RecentEvents(sessionID, 20)
	if err != nil {
		return "", nil, err
	}
	return formatActiveSessionSummary(events), events, nil
}

func formatActiveSessionSummary(events []domain.SessionEvent) string {
	if len(events) == 0 {
		return "Active session just started."
	}
	var b strings.Builder
	b.WriteString("Current session events:\n")
	for i, event := range events {
		fmt.Fprintf(&b, "%d. [%s] %s\n", i+1, event.EventType, event.Content)
	}
	return strings.TrimSpace(b.String())
}

func (s *Store) EndSession(ctx context.Context, sessionID string) (domain.Session, error) {
	if sessionID == "" {
		sessionID = s.activeID
	}
	if sessionID == "" {
		return domain.Session{}, fmt.Errorf("no active session to end")
	}

	observations, err := s.listObservations(sessionID)
	if err != nil {
		return domain.Session{}, err
	}

	var summary string
	if len(observations) > 0 && s.summarizer != nil {
		summary, err = s.summarizer.Summarize(ctx, observations)
		if err != nil {
			summary = FallbackSummarize(observations)
		}
	}

	now := time.Now().UTC()
	_, err = s.db.Exec(
		`UPDATE sessions SET ended_at = ?, summary = ? WHERE id = ? AND ended_at IS NULL`,
		now.Format(time.RFC3339), nullIfEmpty(summary), sessionID,
	)
	if err != nil {
		return domain.Session{}, err
	}

	if s.activeID == sessionID {
		s.activeID = ""
	}

	session, err := s.getSession(sessionID)
	if err != nil {
		return domain.Session{}, err
	}

	if summary != "" {
		text := fmt.Sprintf("Session %s: %s", formatSessionDate(session.StartTime.Format(time.RFC3339)), summary)
		if err := s.indexMemory(ctx, sessionID, "summary", text, session.StartTime.Format(time.RFC3339)); err != nil {
			return domain.Session{}, err
		}
	}
	for _, obs := range observations {
		text := fmt.Sprintf("[%s] %s (session %s)", obs.Category, obs.Content, formatSessionDate(session.StartTime.Format(time.RFC3339)))
		if err := s.indexMemory(ctx, sessionID, "observation", text, session.StartTime.Format(time.RFC3339)); err != nil {
			return domain.Session{}, err
		}
	}

	return session, nil
}

func (s *Store) EndActiveSessionIfNeeded(ctx context.Context) error {
	if s.activeID == "" {
		var id string
		err := s.db.QueryRow(`SELECT id FROM sessions WHERE ended_at IS NULL ORDER BY started_at DESC LIMIT 1`).Scan(&id)
		if err == sql.ErrNoRows {
			return nil
		}
		if err != nil {
			return err
		}
		s.activeID = id
	}

	count, err := s.countObservations(s.activeID)
	if err != nil {
		return err
	}
	if count == 0 {
		return nil
	}
	_, err = s.EndSession(ctx, s.activeID)
	return err
}

func (s *Store) ListSessions() ([]SessionView, error) {
	rows, err := s.db.Query(`
		SELECT s.id, s.started_at, s.ended_at, COALESCE(s.summary, ''),
			(SELECT COUNT(*) FROM observations o WHERE o.session_id = s.id) AS event_count
		FROM sessions s
		ORDER BY (s.ended_at IS NULL) DESC, s.started_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []SessionView
	for rows.Next() {
		var view SessionView
		var endedAt sql.NullString
		if err := rows.Scan(&view.ID, &view.StartedAt, &endedAt, &view.Summary, &view.EventCount); err != nil {
			return nil, err
		}
		if endedAt.Valid {
			view.EndedAt = endedAt.String
		} else {
			view.Active = true
		}
		result = append(result, view)
	}
	return result, rows.Err()
}

func (s *Store) getSession(sessionID string) (domain.Session, error) {
	var startedAt string
	var endedAt, summary sql.NullString
	err := s.db.QueryRow(
		`SELECT started_at, ended_at, summary FROM sessions WHERE id = ?`,
		sessionID,
	).Scan(&startedAt, &endedAt, &summary)
	if err != nil {
		return domain.Session{}, err
	}

	start, err := time.Parse(time.RFC3339, startedAt)
	if err != nil {
		return domain.Session{}, err
	}

	session := domain.Session{
		ID:        sessionID,
		StartTime: start,
		Summary:   summary.String,
	}
	if endedAt.Valid {
		end, err := time.Parse(time.RFC3339, endedAt.String)
		if err != nil {
			return domain.Session{}, err
		}
		session.EndTime = &end
	}

	events, err := s.listEvents(sessionID)
	if err != nil {
		return domain.Session{}, err
	}
	session.Events = eventsToObservations(events)
	return session, nil
}

func eventsToObservations(events []domain.SessionEvent) []domain.Observation {
	result := make([]domain.Observation, 0, len(events))
	for _, event := range events {
		result = append(result, domain.Observation{
			Timestamp: event.Timestamp,
			Category:  summaryCategoryForEvent(event),
			Content:   event.Content,
		})
	}
	return result
}

func (s *Store) listEvents(sessionID string) ([]domain.SessionEvent, error) {
	rows, err := s.db.Query(
		`SELECT timestamp, event_type, category, content FROM observations WHERE session_id = ? ORDER BY timestamp ASC`,
		sessionID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.SessionEvent
	for rows.Next() {
		event, err := scanSessionEvent(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, event)
	}
	return result, rows.Err()
}

func (s *Store) listObservations(sessionID string) ([]domain.Observation, error) {
	events, err := s.listEvents(sessionID)
	if err != nil {
		return nil, err
	}
	result := make([]domain.Observation, 0, len(events))
	for _, event := range events {
		if !eventTypeIncludedInSummary(event.EventType) {
			continue
		}
		result = append(result, domain.Observation{
			Timestamp: event.Timestamp,
			Category:  summaryCategoryForEvent(event),
			Content:   event.Content,
		})
	}
	return result, nil
}

func eventTypeIncludedInSummary(eventType domain.SessionEventType) bool {
	switch eventType {
	case domain.EventUserMessage, domain.EventAssistantMessage, domain.EventToolCall:
		return false
	default:
		return true
	}
}

func summaryCategoryForEvent(event domain.SessionEvent) string {
	if event.Category != "" && event.Category != string(event.EventType) {
		return event.Category
	}
	return string(event.EventType)
}

type eventScanner interface {
	Scan(dest ...any) error
}

func scanSessionEvent(row eventScanner) (domain.SessionEvent, error) {
	var ts, eventType, category, content string
	if err := row.Scan(&ts, &eventType, &category, &content); err != nil {
		return domain.SessionEvent{}, err
	}
	timestamp, err := time.Parse(time.RFC3339, ts)
	if err != nil {
		return domain.SessionEvent{}, err
	}
	if eventType == "" {
		eventType = string(domain.EventObservation)
	}
	return domain.SessionEvent{
		Timestamp: timestamp,
		EventType: domain.SessionEventType(eventType),
		Category:  category,
		Content:   content,
	}, nil
}

func (s *Store) countObservations(sessionID string) (int, error) {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM observations WHERE session_id = ?`, sessionID).Scan(&count)
	return count, err
}

func (s *Store) LastSummary() string {
	var summary sql.NullString
	err := s.db.QueryRow(`
		SELECT summary FROM sessions
		WHERE ended_at IS NOT NULL AND summary IS NOT NULL AND summary != ''
		ORDER BY ended_at DESC LIMIT 1
	`).Scan(&summary)
	if err == nil && summary.Valid {
		return summary.String
	}
	return "No previous sessions recorded yet."
}

type SearchResult struct {
	Kind        string  `json:"kind"`
	Content     string  `json:"content"`
	Score       float64 `json:"score"`
	SessionDate string  `json:"sessionDate"`
}

func (s *Store) Search(ctx context.Context, query string, topK int) ([]SearchResult, error) {
	if topK <= 0 {
		topK = 3
	}
	vector, err := s.embedder.Embed(ctx, query)
	if err != nil {
		return nil, err
	}

	scored := make([]SearchResult, 0, len(s.vectors))
	for _, mem := range s.vectors {
		scored = append(scored, SearchResult{
			Kind:        mem.Kind,
			Content:     mem.Content,
			Score:       rag.CosineSimilarity(vector, mem.Vector),
			SessionDate: mem.SessionDate,
		})
	}

	sortResults(scored)
	if len(scored) > topK {
		scored = scored[:topK]
	}
	return scored, nil
}

func sortResults(results []SearchResult) {
	for i := 0; i < len(results); i++ {
		for j := i + 1; j < len(results); j++ {
			if results[j].Score > results[i].Score {
				results[i], results[j] = results[j], results[i]
			}
		}
	}
}

func FormatSearchResults(results []SearchResult) string {
	if len(results) == 0 {
		return "No matching past session memories found."
	}
	var b strings.Builder
	for i, r := range results {
		fmt.Fprintf(&b, "%d. [%s score %.2f] %s\n", i+1, r.Kind, r.Score, r.Content)
	}
	return strings.TrimSpace(b.String())
}

func formatSessionDate(startedAt string) string {
	t, err := time.Parse(time.RFC3339, startedAt)
	if err != nil {
		return startedAt
	}
	return t.Format("2006-01-02")
}

func nullIfEmpty(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

// SaveTurn records the assistant reply (and tool trace) as session events.
// The user message is recorded earlier by session lifecycle ApplyDecision.
func (s *Store) SaveTurn(ctx context.Context, sessionID, _, finalAnswer, traceSummary string) error {
	if sessionID == "" {
		var err error
		sessionID, err = s.EnsureActiveSession(ctx)
		if err != nil {
			return err
		}
	}
	finalAnswer = strings.TrimSpace(finalAnswer)
	if finalAnswer != "" {
		if err := s.AddEvent(ctx, sessionID, domain.EventAssistantMessage, domain.ObservationNote, finalAnswer); err != nil {
			return err
		}
	}
	if traceSummary = strings.TrimSpace(traceSummary); traceSummary != "" {
		if err := s.AddEvent(ctx, sessionID, domain.EventToolCall, domain.ObservationNote, traceSummary); err != nil {
			return err
		}
	}
	return nil
}
