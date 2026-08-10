package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "modernc.org/sqlite"
)

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

	if err := initializeSchema(ctx); err != nil {
		log.Fatalf("erro ao inicializar schema: %v", err)
	}

	fmt.Println("✓ Banco de dados conectado")
}

func initializeSchema(ctx context.Context) error {
	if _, err := DB.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS users (
			id              TEXT PRIMARY KEY,
			email           TEXT NOT NULL UNIQUE,
			password        TEXT NOT NULL,
			name            TEXT NOT NULL DEFAULT '',
			image           TEXT NOT NULL DEFAULT '',
			name_normalized TEXT NOT NULL DEFAULT '',
			created_at      DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		return err
	}

	_, err := DB.ExecContext(ctx, `
		CREATE UNIQUE INDEX IF NOT EXISTS users_name_normalized_unique
		ON users(name_normalized)
		WHERE name_normalized <> ''
	`)
	return err
}
