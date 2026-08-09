package observability

import (
	"log"
	"log/slog"
	"os"
	"sync/atomic"
)

var Logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
	Level: slog.LevelInfo,
}))

var verificationEmailFailures atomic.Uint64

func IncVerificationEmailFailure() {
	verificationEmailFailures.Add(1)
}

func VerificationEmailFailures() uint64 {
	return verificationEmailFailures.Load()
}

func init() {
	slog.SetDefault(Logger)
	log.SetFlags(0)
	log.SetOutput(slog.NewLogLogger(Logger.Handler(), slog.LevelInfo).Writer())
}
