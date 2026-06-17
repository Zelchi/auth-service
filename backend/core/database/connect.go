package database

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"os"

	_ "modernc.org/sqlite"
)

//go:embed migrations/001_init.sql
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

	if err = DB.Ping(); err != nil {
		log.Fatalf("erro ao conectar ao banco: %v", err)
	}

	migration, err := migrations.ReadFile("migrations/001_init.sql")
	if err != nil {
		log.Fatalf("erro ao ler migration: %v", err)
	}
	if _, err = DB.Exec(string(migration)); err != nil {
		log.Fatalf("erro ao executar migration: %v", err)
	}

	fmt.Println("✓ Banco de dados conectado")
}
