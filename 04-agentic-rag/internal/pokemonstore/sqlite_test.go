package pokemonstore

import (
	"testing"
)

func TestSearchTeamSortByCaughtDate(t *testing.T) {
	t.Parallel()

	store, err := NewSQLiteStore()
	if err != nil {
		t.Fatalf("NewSQLiteStore: %v", err)
	}
	defer store.Close()

	if _, err := store.db.Exec(`
		INSERT INTO team_pokemon
			(dex_id, name, level, primary_type, hp, max_hp, caught_date, slot_order)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, 4, "Charmander", 10, "FIRE", 28, 28, "2024-01-10", 3); err != nil {
		t.Fatalf("insert charmander: %v", err)
	}

	bySlot, err := store.SearchTeam(TeamFilter{SortBy: SortBySlot})
	if err != nil {
		t.Fatalf("SearchTeam slot: %v", err)
	}
	if len(bySlot) < 3 || bySlot[0].Name != "Bulbasaur" {
		t.Fatalf("expected Bulbasaur first by slot, got %#v", bySlot)
	}

	byCaught, err := store.SearchTeam(TeamFilter{SortBy: SortByCaughtDate})
	if err != nil {
		t.Fatalf("SearchTeam caught_date: %v", err)
	}
	if len(byCaught) < 3 || byCaught[0].Name != "Charmander" {
		t.Fatalf("expected Charmander first by caught date, got %#v", byCaught)
	}

	latest, err := store.SearchTeam(TeamFilter{SortBy: SortByCaughtDate, SortDesc: true})
	if err != nil {
		t.Fatalf("SearchTeam caught_date desc: %v", err)
	}
	if len(latest) < 3 || latest[0].Name != "Pidgey" {
		t.Fatalf("expected Pidgey most recent catch, got %#v", latest)
	}
}
