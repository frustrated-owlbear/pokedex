package session

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/frustrated-owlbear/pokedex/04-agentic-rag/internal/rag"
	_ "github.com/mattn/go-sqlite3"
)

const sessionDSN = "file:pokedex-sessions?mode=memory&cache=shared"

type Store struct {
	db       *sql.DB
	embedder rag.Embedder
	vectors  []storedMessage
}

type storedMessage struct {
	ID        int64
	SessionID string
	Role      string
	Content   string
	CreatedAt time.Time
	Vector    []float32
}

func NewStore(embedder rag.Embedder) (*Store, error) {
	db, err := sql.Open("sqlite3", sessionDSN)
	if err != nil {
		return nil, err
	}
	s := &Store{db: db, embedder: embedder}
	if err := s.init(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) Ready() bool {
	return s.db != nil
}

func (s *Store) init() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS sessions (
			id TEXT PRIMARY KEY,
			created_at TEXT NOT NULL
		);
		CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			session_id TEXT NOT NULL,
			role TEXT NOT NULL,
			content TEXT NOT NULL,
			created_at TEXT NOT NULL,
			FOREIGN KEY(session_id) REFERENCES sessions(id)
		);
	`)
	return err
}

func (s *Store) SaveTurn(ctx context.Context, sessionID, userInput, finalAnswer, traceSummary string) error {
	if strings.TrimSpace(sessionID) == "" {
		sessionID = "default"
	}
	now := time.Now().UTC()
	if _, err := s.db.Exec(
		`INSERT OR IGNORE INTO sessions (id, created_at) VALUES (?, ?)`,
		sessionID, now.Format(time.RFC3339),
	); err != nil {
		return err
	}

	if err := s.insertMessage(ctx, sessionID, "user", userInput, now); err != nil {
		return err
	}
	if err := s.insertMessage(ctx, sessionID, "assistant", finalAnswer, now.Add(time.Millisecond)); err != nil {
		return err
	}
	if traceSummary != "" {
		return s.insertMessage(ctx, sessionID, "trace", traceSummary, now.Add(2*time.Millisecond))
	}
	return nil
}

func (s *Store) insertMessage(ctx context.Context, sessionID, role, content string, at time.Time) error {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil
	}
	res, err := s.db.Exec(
		`INSERT INTO messages (session_id, role, content, created_at) VALUES (?, ?, ?, ?)`,
		sessionID, role, content, at.Format(time.RFC3339),
	)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}

	vector, err := s.embedder.Embed(ctx, content)
	if err != nil {
		return err
	}
	s.vectors = append(s.vectors, storedMessage{
		ID:        id,
		SessionID: sessionID,
		Role:      role,
		Content:   content,
		CreatedAt: at,
		Vector:    vector,
	})
	return nil
}

func (s *Store) LastSummary() string {
	for i := len(s.vectors) - 1; i >= 0; i-- {
		if s.vectors[i].Role == "trace" {
			return s.vectors[i].Content
		}
	}
	for i := len(s.vectors) - 1; i >= 0; i-- {
		if s.vectors[i].Role == "assistant" {
			content := s.vectors[i].Content
			if len(content) > 120 {
				return content[:120] + "…"
			}
			return content
		}
	}
	return "No previous sessions recorded yet."
}

type SearchResult struct {
	Role      string  `json:"role"`
	Content   string  `json:"content"`
	Score     float64 `json:"score"`
	CreatedAt string  `json:"createdAt"`
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
	for _, msg := range s.vectors {
		scored = append(scored, SearchResult{
			Role:      msg.Role,
			Content:   msg.Content,
			Score:     rag.CosineSimilarity(vector, msg.Vector),
			CreatedAt: msg.CreatedAt.Format(time.RFC3339),
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
		return "No matching past conversations found."
	}
	var b strings.Builder
	for i, r := range results {
		fmt.Fprintf(&b, "%d. [%s score %.2f] %s\n", i+1, r.Role, r.Score, r.Content)
	}
	return strings.TrimSpace(b.String())
}
