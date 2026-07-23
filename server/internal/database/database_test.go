package database

import (
	"testing"
	"testing/fstest"
)

func TestOpenRunsMigrations(t *testing.T) {
	migrationFS := fstest.MapFS{
		"001_test.sql": {
			Data: []byte("CREATE TABLE example (id INTEGER PRIMARY KEY);"),
		},
	}

	db, err := Open(":memory:", migrationFS)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	var name string
	err = db.QueryRow(
		"SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'example'",
	).Scan(&name)
	if err != nil {
		t.Fatalf("query migrated table: %v", err)
	}
}
