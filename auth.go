package main

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"sync"
	"time"
)

type Session struct {
	Username  string
	Role      string
	ExpiresAt time.Time
}

var sessions = make(map[string]Session)
var sessionsMu sync.RWMutex

func newSessionID() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func createSession(w http.ResponseWriter, user User) error {
	sessionID, err := newSessionID()
	if err != nil {
		log.Printf("Error creating session ID for user %s: %v", user.Username, err)
		return err
	}

	expiresAt := time.Now().Add(8 * time.Hour)

	sessionsMu.Lock()
	sessions[sessionID] = Session{
		Username:  user.Username,
		Role:      user.Role,
		ExpiresAt: expiresAt,
	}
	sessionsMu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   8 * 60 * 60,
		Secure:   false, // Set to true in production with HTTPS
	})

	return nil
}

func getSession(r *http.Request) (*Session, error) { return nil, nil }

func clearSession(w http.ResponseWriter, r *http.Request) error { return nil }

func requireAuth(next http.HandlerFunc) http.HandlerFunc { return nil }

func requireRole(role string, next http.HandlerFunc) http.HandlerFunc { return nil }
