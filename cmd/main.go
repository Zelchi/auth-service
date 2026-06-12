package main

import (
	"authentication/core/database"
	"authentication/core/handler"
	"authentication/core/middleware"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("[AVISO]: .env não encontrado!")
	}

	database.Connect()
	defer database.DB.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("/register", handler.Register)
	mux.HandleFunc("/verify", handler.Verify)
	mux.HandleFunc("/login", handler.Login)

	mux.HandleFunc("/me", middleware.Auth(handler.Me))

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8000"
	}

	addr := ":" + port
	fmt.Printf("Auth service rodando na porta %s\n", port)

	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatalf("erro ao iniciar servidor: %v", err)
	}
}
