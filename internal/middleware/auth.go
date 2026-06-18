// Package middleware provides HTTP middleware for the web server.
package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"time"

	"alfaunit1/internal/db"
)

const (
	sessionCookieName = "alfa_session"
	sessionDuration   = 12 * time.Hour
)

// RequireAuth is an HTTP middleware that ensures the request carries a valid
// admin session cookie. Unauthenticated requests are redirected to /admin/login.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookieName)
		if err != nil || !isValidSession(cookie.Value) {
			http.Redirect(w, r, "/admin/login", http.StatusFound)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// isValidSession checks whether a session token exists in the database and has
// not expired.
func isValidSession(token string) bool {
	if token == "" {
		return false
	}
	var expiresAt string
	err := db.DB.QueryRow(
		`SELECT expires_at FROM sessions WHERE token = ?`, token,
	).Scan(&expiresAt)
	if err != nil {
		return false
	}

	t, err := time.Parse(time.RFC3339, expiresAt)
	if err != nil {
		return false
	}
	return time.Now().Before(t)
}

// CreateSession generates a new session token, persists it to the database,
// and sets the session cookie on the response.
func CreateSession(w http.ResponseWriter) error {
	token, err := generateToken()
	if err != nil {
		return fmt.Errorf("middleware: CreateSession: %w", err)
	}

	expiresAt := time.Now().Add(sessionDuration).Format(time.RFC3339)
	if _, err := db.DB.Exec(
		`INSERT INTO sessions (token, expires_at) VALUES (?, ?)`, token, expiresAt,
	); err != nil {
		return fmt.Errorf("middleware: persist session: %w", err)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/admin",
		Expires:  time.Now().Add(sessionDuration),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

// DestroySession deletes the session from the database and clears the cookie.
func DestroySession(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(sessionCookieName)
	if err == nil && cookie.Value != "" {
		_, _ = db.DB.Exec(`DELETE FROM sessions WHERE token = ?`, cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/admin",
		MaxAge:   -1,
		HttpOnly: true,
	})
}

// generateToken returns a cryptographically random 32-byte hex string.
func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
