package database

import (
	"database/sql"
	"testing"
)

func TestApplyMigrationsIsVersionedAndIdempotent(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })

	previousDB := DB
	DB = db
	t.Cleanup(func() { DB = previousDB })

	if err := applyMigrations(); err != nil {
		t.Fatalf("first applyMigrations() error = %v", err)
	}
	if err := applyMigrations(); err != nil {
		t.Fatalf("second applyMigrations() error = %v", err)
	}

	var migrationCount int
	if err := DB.QueryRow(`SELECT COUNT(*) FROM schema_migrations`).Scan(&migrationCount); err != nil {
		t.Fatalf("schema_migrations query error = %v", err)
	}
	if migrationCount != 1 {
		t.Fatalf("migration count = %d, want 1", migrationCount)
	}

	for _, table := range []string{"users"} {
		var name string
		if err := DB.QueryRow(
			`SELECT name FROM sqlite_master WHERE type = 'table' AND name = ?`,
			table,
		).Scan(&name); err != nil {
			t.Fatalf("table %q was not created: %v", table, err)
		}
	}

	var pendingTableCount int
	if err := DB.QueryRow(
		`SELECT COUNT(*) FROM sqlite_master WHERE type = 'table' AND name = 'pending_registrations'`,
	).Scan(&pendingTableCount); err != nil {
		t.Fatalf("pending table query error: %v", err)
	}
	if pendingTableCount != 0 {
		t.Fatal("pending registrations must be kept in memory")
	}
}
