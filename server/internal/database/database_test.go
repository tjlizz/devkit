package database

import (
	"io/fs"
	"testing"
	"testing/fstest"

	"devkit/server/migrations"
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

func TestVerificationMigration(t *testing.T) {
	migrationFS, err := fs.Sub(migrations.Files, ".")
	if err != nil {
		t.Fatalf("open embedded migrations: %v", err)
	}
	db, err := Open(":memory:", migrationFS)
	if err != nil {
		t.Fatalf("Open() error = %v", err)
	}
	defer db.Close()

	for _, column := range []string{"verified_at", "avatar_url", "display_name"} {
		var count int
		if err := db.QueryRow(
			"SELECT COUNT(*) FROM pragma_table_info('users') WHERE name = ?",
			column,
		).Scan(&count); err != nil {
			t.Fatalf("query users column %s: %v", column, err)
		}
		if count != 1 {
			t.Errorf("users column %s count = %d, want 1", column, count)
		}
	}

	var table string
	if err := db.QueryRow(
		"SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'user_verifications'",
	).Scan(&table); err != nil {
		t.Fatalf("query user_verifications table: %v", err)
	}
}
