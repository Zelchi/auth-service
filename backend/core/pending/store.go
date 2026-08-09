package pending

import (
	"context"
	"crypto/subtle"
	"sync"
	"time"
)

const (
	TTL         = 15 * time.Minute
	MaxAttempts = 5
)

type Registration struct {
	Email        string
	PasswordHash string
	CodeDigest   string
	Attempts     int
	ExpiresAt    time.Time
}

// store keeps pending registrations local to this process. The mutex makes
// Check's attempt counter atomic when multiple requests verify the same email.
type store struct {
	mu      sync.Mutex
	byEmail map[string]Registration
}

var s = &store{
	byEmail: make(map[string]Registration),
}

func init() {
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			s.purgeExpired()
		}
	}()
}

func Put(_ context.Context, email, passwordHash, codeDigest string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.byEmail[email] = Registration{
		Email:        email,
		PasswordHash: passwordHash,
		CodeDigest:   codeDigest,
		ExpiresAt:    time.Now().UTC().Add(TTL),
	}
	return nil
}

// ReplaceCode troca o código de um cadastro pendente e zera as tentativas.
// O registro expirado não pode ser reativado pelo reenvio.
func ReplaceCode(_ context.Context, email, codeDigest string) (bool, error) {
	now := time.Now().UTC()

	s.mu.Lock()
	defer s.mu.Unlock()

	reg, ok := s.byEmail[email]
	if !ok || !now.Before(reg.ExpiresAt) {
		delete(s.byEmail, email)
		return false, nil
	}

	reg.CodeDigest = codeDigest
	reg.Attempts = 0
	reg.ExpiresAt = now.Add(TTL)
	s.byEmail[email] = reg
	return true, nil
}

// Check compara o digest e registra tentativas inválidas atomicamente.
// O terceiro retorno diferencia código inválido de cadastro inexistente.
func Check(_ context.Context, email, codeDigest string) (Registration, bool, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	reg, ok := s.byEmail[email]
	if !ok {
		return Registration{}, false, false, nil
	}
	if !time.Now().UTC().Before(reg.ExpiresAt) {
		delete(s.byEmail, email)
		return Registration{}, false, false, nil
	}

	if subtle.ConstantTimeCompare([]byte(reg.CodeDigest), []byte(codeDigest)) == 1 {
		return reg, true, true, nil
	}

	reg.Attempts++
	if reg.Attempts >= MaxAttempts {
		delete(s.byEmail, email)
	} else {
		s.byEmail[email] = reg
	}

	return reg, false, true, nil
}

func Delete(_ context.Context, email string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.byEmail, email)
	return nil
}

func (s *store) purgeExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	for email, reg := range s.byEmail {
		if !now.Before(reg.ExpiresAt) {
			delete(s.byEmail, email)
		}
	}
}
