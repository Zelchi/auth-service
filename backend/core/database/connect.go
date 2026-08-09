package database

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"strings"
	"time"

	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrations embed.FS

var DB *sql.DB

func Connect() {
	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		log.Fatal("DB_URL não definida")
	}

	var err error
	DB, err = sql.Open("sqlite", dsn)
	if err != nil {
		log.Fatalf("erro ao abrir conexão com banco: %v", err)
	}
	DB.SetMaxOpenConns(1)
	DB.SetMaxIdleConns(1)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = DB.PingContext(ctx); err != nil {
		log.Fatalf("erro ao conectar ao banco: %v", err)
	}
	if _, err = DB.ExecContext(ctx, `PRAGMA busy_timeout = 5000`); err != nil {
		log.Fatalf("erro ao configurar busy_timeout: %v", err)
	}
	if _, err = DB.ExecContext(ctx, `PRAGMA journal_mode = WAL`); err != nil {
		log.Fatalf("erro ao configurar WAL: %v", err)
	}

	if err := applyMigrations(); err != nil {
		log.Fatalf("erro ao executar migrations: %v", err)
	}

	fmt.Println("✓ Banco de dados conectado")
}

func applyMigrations() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if _, err := DB.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    TEXT PRIMARY KEY,
			applied_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		return err
	}

	entries, err := migrations.ReadDir("migrations")
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}

		var applied int
		err := DB.QueryRowContext(ctx,
			`SELECT 1 FROM schema_migrations WHERE version = ?`,
			entry.Name(),
		).Scan(&applied)
		if err == nil {
			continue
		}
		if !errors.Is(err, sql.ErrNoRows) {
			return err
		}

		script, err := fs.ReadFile(migrations, "migrations/"+entry.Name())
		if err != nil {
			return err
		}

		tx, err := DB.BeginTx(ctx, nil)
		if err != nil {
			return err
		}

		if _, err = tx.ExecContext(ctx, string(script)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("%s: %w", entry.Name(), err)
		}
		if _, err = tx.ExecContext(ctx,
			`INSERT INTO schema_migrations (version) VALUES (?)`,
			entry.Name(),
		); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("registrar %s: %w", entry.Name(), err)
		}
		if err = tx.Commit(); err != nil {
			return fmt.Errorf("confirmar %s: %w", entry.Name(), err)
		}
	}

	return nil
}
