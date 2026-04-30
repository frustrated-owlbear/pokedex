package pokemonstore

import (
	"database/sql"

	"github.com/frustrated-owlbear/pokedex/internal/domain"
	_ "github.com/mattn/go-sqlite3"
)

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	store := &SQLiteStore{db: db}
	if err := store.init(); err != nil {
		_ = db.Close()
		return nil, err
	}

	return store, nil
}

func (s *SQLiteStore) Close() error {
	return s.db.Close()
}

func (s *SQLiteStore) ListPokemons() ([]domain.Pokemon, error) {
	rows, err := s.db.Query(`SELECT name FROM pokemons ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.Pokemon
	for rows.Next() {
		var pokemon domain.Pokemon
		if err := rows.Scan(&pokemon.Name); err != nil {
			return nil, err
		}
		result = append(result, pokemon)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *SQLiteStore) init() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS pokemons (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE
		)
	`)
	if err != nil {
		return err
	}

	for _, name := range []string{"Squirtle", "Bulbasaur", "Charmander"} {
		if _, err := s.db.Exec(`INSERT OR IGNORE INTO pokemons(name) VALUES(?)`, name); err != nil {
			return err
		}
	}

	return nil
}
