package main

import (
	"authentication/core/database"
	"authentication/core/handler"
	"authentication/core/middleware"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
)

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func console(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()

		log.Printf("→ %s %s", r.Method, r.URL.Path)

		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)

		duration := time.Since(now)
		log.Printf("← %s %s %d (%s)",
			r.Method, r.URL.Path, rw.status, duration,
		)
	})
}

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
		port = "8888"
	}

	server := &http.Server{
		Addr:    ":" + port,
		Handler: console(mux),
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		fmt.Printf("Auth service rodando na porta %s\n", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("erro ao iniciar servidor: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("Recebido sinal de shutdown. Encerrando servidor...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("erro no graceful shutdown: %v", err)
	}

	log.Println("Servidor encerrado graciosamente.")
}
