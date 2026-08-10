package database

import (
	"context"
	"database/sql"
	"testing"
)

func TestInitializeSchemaIsIdempotent(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	db.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = db.Close() })

	previousDB := DB
	DB = db
	t.Cleanup(func() { DB = previousDB })

	if err := initializeSchema(context.Background()); err != nil {
		t.Fatalf("first initializeSchema() error = %v", err)
	}
	if err := initializeSchema(context.Background()); err != nil {
		t.Fatalf("second initializeSchema() error = %v", err)
	}

	var tableName string
	if err := DB.QueryRow(
		`SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'users'`,
	).Scan(&tableName); err != nil {
		t.Fatalf("users table was not created: %v", err)
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

	for _, column := range []string{"name", "image", "name_normalized"} {
		var columnCount int
		if err := DB.QueryRow(
			`SELECT COUNT(*) FROM pragma_table_info('users') WHERE name = ?`,
			column,
		).Scan(&columnCount); err != nil {
			t.Fatalf("column %q query error: %v", column, err)
		}
		if columnCount != 1 {
			t.Fatalf("column %q was not created", column)
		}
	}
}
