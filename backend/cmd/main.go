package main

import (
	"authentication/core/database"
	"authentication/core/handler"
	"authentication/core/jwt"
	"authentication/core/middleware"
	_ "authentication/core/observability"
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"log/slog"
	"net/http"
	"net/mail"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

//go:embed dist
var frontend embed.FS

type responseWriter struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (rw *responseWriter) WriteHeader(status int) {
	if rw.wroteHeader {
		return
	}

	rw.wroteHeader = true
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWriter) Write(body []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	return rw.ResponseWriter.Write(body)
}

func securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		frameAncestors := strings.TrimSpace(os.Getenv("FRAME_ANCESTORS"))
		if frameAncestors == "" {
			frameAncestors = "'self'"
		}

		w.Header().Set("Content-Security-Policy", "frame-ancestors "+frameAncestors)
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

		next.ServeHTTP(w, r)
	})
}

func limitRequestBody(next http.Handler) http.Handler {
	const maxRequestBody = 1 << 20

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxRequestBody)
		next.ServeHTTP(w, r)
	})
}

func validateEnvironment() {
	if err := jwt.ValidateSecret(); err != nil {
		log.Fatalf("configuração inválida: %v", err)
	}

	for _, key := range []string{"DB_URL", "RESEND_API_KEY", "RESEND_FROM"} {
		if strings.TrimSpace(os.Getenv(key)) == "" {
			log.Fatalf("configuração inválida: %s não definida", key)
		}
	}
	if _, err := mail.ParseAddress(strings.TrimSpace(os.Getenv("RESEND_FROM"))); err != nil {
		log.Fatalf("configuração inválida: RESEND_FROM deve ser um endereço válido")
	}
}

func console(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		now := time.Now()
		requestID := r.Header.Get("X-Request-ID")
		if !validRequestID(requestID) {
			requestID = uuid.NewString()
		}
		w.Header().Set("X-Request-ID", requestID)
		slog.Info("http_request_started", "request_id", requestID, "method", r.Method, "path", r.URL.Path)

		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)

		slog.Info("http_request_finished", "request_id", requestID, "method", r.Method, "path", r.URL.Path, "status", rw.status, "duration", time.Since(now).String())
	})
}

func validRequestID(value string) bool {
	if len(value) == 0 || len(value) > 64 {
		return false
	}
	for _, char := range value {
		if (char < 'a' || char > 'z') && (char < 'A' || char > 'Z') &&
			(char < '0' || char > '9') && char != '-' && char != '_' {
			return false
		}
	}
	return true
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("[AVISO]: .env não encontrado!")
	}

	validateEnvironment()
	database.Connect()
	defer database.DB.Close()

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", handler.Health)
	mux.HandleFunc("/readyz", handler.Ready)
	mux.HandleFunc("/metrics", handler.Metrics)

	registerLimiter := middleware.NewRateLimiter(5, 15*time.Minute)
	registerEmailLimiter := middleware.NewRateLimiter(5, 15*time.Minute)
	registerComboLimiter := middleware.NewRateLimiter(5, 15*time.Minute)
	verifyLimiter := middleware.NewRateLimiter(10, 15*time.Minute)
	verifyEmailLimiter := middleware.NewRateLimiter(10, 15*time.Minute)
	verifyComboLimiter := middleware.NewRateLimiter(10, 15*time.Minute)
	resendLimiter := middleware.NewRateLimiter(5, 15*time.Minute)
	resendEmailLimiter := middleware.NewRateLimiter(5, 15*time.Minute)
	resendComboLimiter := middleware.NewRateLimiter(5, 15*time.Minute)
	loginLimiter := middleware.NewRateLimiter(10, 15*time.Minute)
	loginEmailLimiter := middleware.NewRateLimiter(10, 15*time.Minute)
	loginComboLimiter := middleware.NewRateLimiter(10, 15*time.Minute)
	bridgeLimiter := middleware.NewRateLimiter(20, 15*time.Minute)

	mux.Handle("/api/register", middleware.LimitMany([]middleware.RateLimitRule{
		{Limiter: registerLimiter, Key: middleware.ClientIP},
		{Limiter: registerEmailLimiter, Key: middleware.Email},
		{Limiter: registerComboLimiter, Key: middleware.ClientIPEmail},
	}, http.HandlerFunc(handler.Register)))
	mux.Handle("/api/verify", middleware.LimitMany([]middleware.RateLimitRule{
		{Limiter: verifyLimiter, Key: middleware.ClientIP},
		{Limiter: verifyEmailLimiter, Key: middleware.Email},
		{Limiter: verifyComboLimiter, Key: middleware.ClientIPEmail},
	}, http.HandlerFunc(handler.Verify)))
	mux.Handle("/api/resend", middleware.LimitMany([]middleware.RateLimitRule{
		{Limiter: resendLimiter, Key: middleware.ClientIP},
		{Limiter: resendEmailLimiter, Key: middleware.Email},
		{Limiter: resendComboLimiter, Key: middleware.ClientIPEmail},
	}, http.HandlerFunc(handler.Resend)))
	mux.Handle("/api/login", middleware.LimitMany([]middleware.RateLimitRule{
		{Limiter: loginLimiter, Key: middleware.ClientIP},
		{Limiter: loginEmailLimiter, Key: middleware.Email},
		{Limiter: loginComboLimiter, Key: middleware.ClientIPEmail},
	}, http.HandlerFunc(handler.Login)))
	mux.HandleFunc("/api/logout", handler.Logout)
	mux.Handle("/api/bridge/token", middleware.Limit(bridgeLimiter, middleware.ClientIP, middleware.Auth(handler.BridgeToken)))
	mux.HandleFunc("/api/me", middleware.Auth(handler.Me))

	stripped, err := fs.Sub(frontend, "dist")
	if err != nil {
		log.Fatalf("erro ao preparar assets: %v", err)
	}
	fileServer := http.FileServer(http.FS(stripped))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, statErr := fs.Stat(stripped, r.URL.Path[1:])
		if r.URL.Path != "/" && statErr != nil {
			r.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, r)
	})

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8888"
	}

	server := &http.Server{
		Addr:              ":" + port,
		Handler:           console(securityHeaders(limitRequestBody(mux))),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
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
