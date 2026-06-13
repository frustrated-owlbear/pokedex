package pokemonstore

import (
	"database/sql"
	"fmt"
	"strings"
	"sync/atomic"

	"github.com/frustrated-owlbear/pokedex/04-agentic-rag/internal/domain"
	_ "github.com/mattn/go-sqlite3"
)

var sqliteStoreID atomic.Int64

const teamSelectColumns = `
	id, dex_id, name, level, primary_type, hp, max_hp, caught_date, birthday
`

func newInMemoryDSN() string {
	id := sqliteStoreID.Add(1)
	// Shared cache lets every connection in the pool see the same in-memory DB.
	return fmt.Sprintf("file:pokedex%d?mode=memory&cache=shared", id)
}

type SQLiteStore struct {
	db *sql.DB
}

func NewSQLiteStore() (*SQLiteStore, error) {
	db, err := sql.Open("sqlite3", newInMemoryDSN())
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

func (s *SQLiteStore) ListTeam() ([]domain.TeamPokemon, error) {
	return s.SearchTeam(TeamFilter{})
}

func (s *SQLiteStore) SearchTeam(filter TeamFilter) ([]domain.TeamPokemon, error) {
	query := `
		SELECT ` + teamSelectColumns + `
		FROM team_pokemon
	`
	var (
		clauses []string
		args    []any
	)

	if filter.PrimaryType != "" {
		clauses = append(clauses, "LOWER(primary_type) = LOWER(?)")
		args = append(args, filter.PrimaryType)
	}
	if filter.Name != "" {
		clauses = append(clauses, "LOWER(name) LIKE '%' || LOWER(?) || '%'")
		args = append(args, filter.Name)
	}
	if filter.DexID != nil {
		clauses = append(clauses, "dex_id = ?")
		args = append(args, *filter.DexID)
	}
	if filter.Level != nil {
		clauses = append(clauses, "level = ?")
		args = append(args, *filter.Level)
	} else {
		if filter.MinLevel != nil {
			clauses = append(clauses, "level >= ?")
			args = append(args, *filter.MinLevel)
		}
		if filter.MaxLevel != nil {
			clauses = append(clauses, "level <= ?")
			args = append(args, *filter.MaxLevel)
		}
	}
	if filter.CaughtDate != "" {
		clauses = append(clauses, "caught_date = ?")
		args = append(args, filter.CaughtDate)
	} else {
		if filter.CaughtAfter != "" {
			clauses = append(clauses, "caught_date >= ?")
			args = append(args, filter.CaughtAfter)
		}
		if filter.CaughtBefore != "" {
			clauses = append(clauses, "caught_date <= ?")
			args = append(args, filter.CaughtBefore)
		}
	}

	if len(clauses) > 0 {
		query += " WHERE " + strings.Join(clauses, " AND ")
	}

	sortBy := filter.SortBy
	if sortBy == "" {
		sortBy = SortBySlot
	}
	sortColumn, err := sortBy.SQLColumn()
	if err != nil {
		return nil, err
	}
	order := "ASC"
	if filter.SortDesc {
		order = "DESC"
	}
	query += fmt.Sprintf(" ORDER BY %s %s, slot_order ASC", sortColumn, order)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("search team: %w", err)
	}
	defer rows.Close()

	return scanTeamRows(rows)
}

func (s *SQLiteStore) GetTeamMember(id int) (domain.TeamPokemon, error) {
	row := s.db.QueryRow(`
		SELECT `+teamSelectColumns+`
		FROM team_pokemon
		WHERE id = ?
	`, id)

	pokemon, err := scanTeamPokemon(row)
	if err == sql.ErrNoRows {
		return domain.TeamPokemon{}, err
	}
	return pokemon, err
}

func (s *SQLiteStore) init() error {
	if _, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS team_pokemon (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			dex_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			level INTEGER NOT NULL,
			primary_type TEXT NOT NULL,
			hp INTEGER NOT NULL,
			max_hp INTEGER NOT NULL,
			caught_date TEXT NOT NULL,
			birthday TEXT,
			slot_order INTEGER NOT NULL DEFAULT 0,
			UNIQUE(slot_order)
		)
	`); err != nil {
		return err
	}

	seeds := []struct {
		dexID       int
		name        string
		level       int
		primaryType string
		hp          int
		maxHP       int
		caughtDate  string
		birthday    string
		slot        int
	}{
		{1, "Bulbasaur", 16, "GRASS", 42, 42, "2024-02-15", "2024-02-14", 1},
		{16, "Pidgey", 12, "NORMAL", 31, 31, "2024-03-01", "", 2},
	}

	for _, seed := range seeds {
		var birthday any
		if seed.birthday != "" {
			birthday = seed.birthday
		}

		if _, err := s.db.Exec(`
			INSERT OR IGNORE INTO team_pokemon
				(dex_id, name, level, primary_type, hp, max_hp, caught_date, birthday, slot_order)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, seed.dexID, seed.name, seed.level, seed.primaryType, seed.hp, seed.maxHP, seed.caughtDate, birthday, seed.slot); err != nil {
			return err
		}
	}

	return nil
}

func scanTeamRows(rows *sql.Rows) ([]domain.TeamPokemon, error) {
	var result []domain.TeamPokemon
	for rows.Next() {
		pokemon, err := scanTeamPokemon(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, pokemon)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

type teamScanner interface {
	Scan(dest ...any) error
}

func scanTeamPokemon(row teamScanner) (domain.TeamPokemon, error) {
	var pokemon domain.TeamPokemon
	var birthday sql.NullString
	if err := row.Scan(
		&pokemon.ID,
		&pokemon.DexID,
		&pokemon.Name,
		&pokemon.Level,
		&pokemon.PrimaryType,
		&pokemon.HP,
		&pokemon.MaxHP,
		&pokemon.CaughtDate,
		&birthday,
	); err != nil {
		return domain.TeamPokemon{}, err
	}
	if birthday.Valid {
		pokemon.Birthday = birthday.String
	}
	pokemon.ImageURL = domain.ImageURL(pokemon.DexID)
	return pokemon, nil
}
