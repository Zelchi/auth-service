package pending

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestCheckAcceptsMatchingDigest(t *testing.T) {
	ctx := context.Background()
	email := "matching@example.com"
	if err := Put(ctx, email, "hash", "digest"); err != nil {
		t.Fatalf("Put() error = %v", err)
	}
	t.Cleanup(func() { _ = Delete(ctx, email) })

	reg, valid, exists, err := Check(ctx, email, "digest")
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !exists || !valid {
		t.Fatal("Check() rejected a matching digest")
	}
	if reg.Attempts != 0 {
		t.Fatalf("attempts = %d, want 0", reg.Attempts)
	}
}

func TestCheckInvalidatesAfterMaximumAttempts(t *testing.T) {
	ctx := context.Background()
	email := "attempts@example.com"
	if err := Put(ctx, email, "hash", "digest"); err != nil {
		t.Fatalf("Put() error = %v", err)
	}
	t.Cleanup(func() { _ = Delete(ctx, email) })

	for attempt := 1; attempt <= MaxAttempts; attempt++ {
		reg, valid, exists, err := Check(ctx, email, "wrong")
		if err != nil {
			t.Fatalf("Check() error = %v", err)
		}
		if !exists || valid {
			t.Fatalf("attempt %d returned exists=%v valid=%v", attempt, exists, valid)
		}
		if reg.Attempts != attempt {
			t.Fatalf("attempts = %d, want %d", reg.Attempts, attempt)
		}
	}

	_, _, exists, err := Check(ctx, email, "digest")
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if exists {
		t.Fatal("registration still exists after maximum attempts")
	}
}

func TestCheckRemovesExpiredRegistration(t *testing.T) {
	ctx := context.Background()
	email := "expired@example.com"
	if err := Put(ctx, email, "hash", "digest"); err != nil {
		t.Fatalf("Put() error = %v", err)
	}
	t.Cleanup(func() { _ = Delete(ctx, email) })

	s.mu.Lock()
	reg := s.byEmail[email]
	reg.ExpiresAt = time.Now().UTC().Add(-time.Second)
	s.byEmail[email] = reg
	s.mu.Unlock()

	_, valid, exists, err := Check(ctx, email, "digest")
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if valid || exists {
		t.Fatalf("expired registration returned valid=%v exists=%v", valid, exists)
	}
}

func TestReplaceCodeResetsAttempts(t *testing.T) {
	ctx := context.Background()
	email := "replace@example.com"
	if err := Put(ctx, email, "hash", "old"); err != nil {
		t.Fatalf("Put() error = %v", err)
	}
	t.Cleanup(func() { _ = Delete(ctx, email) })
	if _, _, _, err := Check(ctx, email, "wrong"); err != nil {
		t.Fatalf("Check() error = %v", err)
	}

	updated, err := ReplaceCode(ctx, email, "new")
	if err != nil {
		t.Fatalf("ReplaceCode() error = %v", err)
	}
	if !updated {
		t.Fatal("ReplaceCode() reported no updated registration")
	}
	reg, valid, exists, err := Check(ctx, email, "new")
	if err != nil {
		t.Fatalf("Check() error = %v", err)
	}
	if !exists || !valid || reg.Attempts != 0 {
		t.Fatalf("registration after replacement = %+v, valid=%v exists=%v", reg, valid, exists)
	}
}

func TestReplaceCodeDoesNotReactivateExpiredRegistration(t *testing.T) {
	ctx := context.Background()
	email := "expired-replace@example.com"
	if err := Put(ctx, email, "hash", "old"); err != nil {
		t.Fatalf("Put() error = %v", err)
	}
	t.Cleanup(func() { _ = Delete(ctx, email) })

	s.mu.Lock()
	reg := s.byEmail[email]
	reg.ExpiresAt = time.Now().UTC().Add(-time.Second)
	s.byEmail[email] = reg
	s.mu.Unlock()

	updated, err := ReplaceCode(ctx, email, "new")
	if err != nil {
		t.Fatalf("ReplaceCode() error = %v", err)
	}
	if updated {
		t.Fatal("ReplaceCode() reactivated an expired registration")
	}
}

func TestConcurrentChecksRespectAttemptLimit(t *testing.T) {
	ctx := context.Background()
	email := "concurrent@example.com"
	if err := Put(ctx, email, "hash", "digest"); err != nil {
		t.Fatalf("Put() error = %v", err)
	}
	t.Cleanup(func() { _ = Delete(ctx, email) })

	var group sync.WaitGroup
	for i := 0; i < MaxAttempts+3; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			_, _, _, _ = Check(ctx, email, "wrong")
		}()
	}
	group.Wait()

	_, _, exists, err := Check(ctx, email, "digest")
	if err != nil {
		t.Fatalf("final Check() error = %v", err)
	}
	if exists {
		t.Fatal("registration survived concurrent maximum attempts")
	}
}
