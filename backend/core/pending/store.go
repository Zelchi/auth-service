package pending

import (
	"sync"
	"time"
)

const TTL = 15 * time.Minute

type Registration struct {
	Email        string
	PasswordHash string
	Code         string
	ExpiresAt    time.Time
}

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

func Put(email, passwordHash, code string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.byEmail[email] = Registration{
		Email:        email,
		PasswordHash: passwordHash,
		Code:         code,
		ExpiresAt:    time.Now().UTC().Add(TTL),
	}
}

func Get(email string) (Registration, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	reg, ok := s.byEmail[email]
	if !ok {
		return Registration{}, false
	}
	if time.Now().UTC().After(reg.ExpiresAt) {
		delete(s.byEmail, email)
		return Registration{}, false
	}
	return reg, true
}

func Delete(email string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.byEmail, email)
}

func (s *store) purgeExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now().UTC()
	for email, reg := range s.byEmail {
		if now.After(reg.ExpiresAt) {
			delete(s.byEmail, email)
		}
	}
}
