package handler

import (
	"authentication/core/database"
	"authentication/core/middleware"
	"authentication/core/pending"
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
)

func testUsersDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`
		CREATE TABLE users (
			id TEXT PRIMARY KEY,
			email TEXT NOT NULL UNIQUE,
			password TEXT NOT NULL,
			created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
		)
	`); err != nil {
		_ = db.Close()
		t.Fatalf("create users table error = %v", err)
	}
	return db
}

func TestVerifyCreatesUserAndConsumesPendingRegistration(t *testing.T) {
	db := testUsersDB(t)
	t.Cleanup(func() { _ = db.Close() })
	previousDB := database.DB
	database.DB = db
	t.Cleanup(func() { database.DB = previousDB })
	ctx := context.Background()
	email := "verify@example.com"
	hash, err := bcrypt.GenerateFromPassword([]byte("correct-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword() error = %v", err)
	}
	if err := pending.Put(ctx, email, string(hash), digestVerificationCode("123456")); err != nil {
		t.Fatalf("pending.Put() error = %v", err)
	}
	t.Cleanup(func() { _ = pending.Delete(ctx, email) })

	req := httptest.NewRequest(http.MethodPost, "/api/verify", strings.NewReader(`{"email":"VERIFY@example.com","code":"123456"}`))
	recorder := httptest.NewRecorder()
	Verify(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM users WHERE email = ?`, email).Scan(&count); err != nil {
		t.Fatalf("users query error = %v", err)
	}
	if count != 1 {
		t.Fatalf("user count = %d, want 1", count)
	}
	_, _, exists, err := pending.Check(ctx, email, digestVerificationCode("123456"))
	if err != nil {
		t.Fatalf("pending.Check() error = %v", err)
	}
	if exists {
		t.Fatal("pending registration was not consumed")
	}
}

func TestMeReturnsAuthenticatedUser(t *testing.T) {
	db := testUsersDB(t)
	t.Cleanup(func() { _ = db.Close() })
	previousDB := database.DB
	database.DB = db
	t.Cleanup(func() { database.DB = previousDB })
	if _, err := db.Exec(`INSERT INTO users (id, email, password) VALUES (?, ?, ?)`, "user-123", "user@example.com", "hash"); err != nil {
		t.Fatalf("insert user error = %v", err)
	}

	ctx := context.WithValue(context.Background(), middleware.UserIDKey, "user-123")
	req := httptest.NewRequest(http.MethodGet, "/api/me", nil).WithContext(ctx)
	recorder := httptest.NewRecorder()
	Me(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
}

func TestRegisterRejectsUnknownJSONFields(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(`{"email":"user@example.com","password":"StrongPassword1","password_confirmation":"StrongPassword1","extra":true}`))
	recorder := httptest.NewRecorder()

	Register(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestRegisterRejectsWeakPassword(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(`{"email":"user@example.com","password":"weak-password1","password_confirmation":"weak-password1"}`))
	recorder := httptest.NewRecorder()

	Register(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestRegisterRejectsMismatchedPasswords(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(`{"email":"user@example.com","password":"StrongPassword1","password_confirmation":"DifferentPassword1"}`))
	recorder := httptest.NewRecorder()

	Register(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

func TestRegisterStoresPendingRegistrationAndSendsCode(t *testing.T) {
	db := testUsersDB(t)
	t.Cleanup(func() { _ = db.Close() })
	previousDB := database.DB
	database.DB = db
	t.Cleanup(func() { database.DB = previousDB })

	previousSender := sendVerificationCode
	var sentEmail, sentCode string
	sendVerificationCode = func(_ context.Context, to, code string) error {
		sentEmail = to
		sentCode = code
		return nil
	}
	t.Cleanup(func() { sendVerificationCode = previousSender })

	req := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(`{"email":"New@Example.com","password":"StrongPassword1","password_confirmation":"StrongPassword1"}`))
	recorder := httptest.NewRecorder()
	Register(recorder, req)

	if recorder.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusCreated)
	}
	if sentEmail != "new@example.com" || len(sentCode) != 6 {
		t.Fatalf("sent email/code = %q/%q", sentEmail, sentCode)
	}
	_, valid, exists, err := pending.Check(context.Background(), sentEmail, digestVerificationCode(sentCode))
	if err != nil {
		t.Fatalf("pending.Check() error = %v", err)
	}
	if !exists || !valid {
		t.Fatal("pending registration was not stored with the sent code")
	}
}

func TestRegisterDeletesPendingRegistrationWhenEmailFails(t *testing.T) {
	db := testUsersDB(t)
	t.Cleanup(func() { _ = db.Close() })
	previousDB := database.DB
	database.DB = db
	t.Cleanup(func() { database.DB = previousDB })

	previousSender := sendVerificationCode
	sendVerificationCode = func(context.Context, string, string) error {
		return errors.New("provider unavailable")
	}
	t.Cleanup(func() { sendVerificationCode = previousSender })

	email := "failed@example.com"
	req := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(`{"email":"failed@example.com","password":"StrongPassword1","password_confirmation":"StrongPassword1"}`))
	recorder := httptest.NewRecorder()
	Register(recorder, req)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusServiceUnavailable)
	}
	_, _, exists, err := pending.Check(context.Background(), email, "anything")
	if err != nil {
		t.Fatalf("pending.Check() error = %v", err)
	}
	if exists {
		t.Fatal("pending registration survived email failure")
	}
}

func TestLoginReturnsServerErrorWhenDatabaseIsUnavailable(t *testing.T) {
	db := testUsersDB(t)
	if err := db.Close(); err != nil {
		t.Fatalf("db.Close() error = %v", err)
	}
	previousDB := database.DB
	database.DB = db
	t.Cleanup(func() { database.DB = previousDB })

	req := httptest.NewRequest(http.MethodPost, "/api/login", strings.NewReader(`{"email":"user@example.com","password":"correct-password"}`))
	recorder := httptest.NewRecorder()

	Login(recorder, req)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusInternalServerError)
	}
}
