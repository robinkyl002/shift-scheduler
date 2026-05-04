package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
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

const SESSION_COOKIE_NAME = "session_id"

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
		Name:     SESSION_COOKIE_NAME,
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

func getSession(r *http.Request) (*Session, string, error) {
	cookie, err := r.Cookie(SESSION_COOKIE_NAME)
	if err != nil {
		return nil, "", err
	}

	sessionsMu.RLock()
	session, ok := sessions[cookie.Value]
	sessionsMu.RUnlock()

	if !ok {
		return nil, cookie.Value, errors.New("session not found")
	}

	if time.Now().After(session.ExpiresAt) {
		return nil, cookie.Value, errors.New("session expired")
	}

	return &session, cookie.Value, nil
}

func clearSession(w http.ResponseWriter, r *http.Request) error {
	cookie, err := r.Cookie(SESSION_COOKIE_NAME)
	if err != nil {
		return err
	}

	sessionsMu.Lock()
	delete(sessions, cookie.Value)
	sessionsMu.Unlock()

	http.SetCookie(w, &http.Cookie{
		Name:     SESSION_COOKIE_NAME,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
	})
	return nil
}

func requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, sessionID, err := getSession(r)
		if err != nil {
			if sessionID != "" {
				sessionsMu.Lock()
				delete(sessions, sessionID)
				sessionsMu.Unlock()
			}

			http.SetCookie(w, &http.Cookie{
				Name:     SESSION_COOKIE_NAME,
				Value:    "",
				Path:     "/",
				MaxAge:   -1,
				HttpOnly: true,
			})
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			// log.Printf("Unauthorized access attempt: %v", err)
			// http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next(w, r)

	}
}

func requireRole(role string, next http.HandlerFunc) http.HandlerFunc { return nil }
