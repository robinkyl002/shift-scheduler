package main

import (
	"crypto/rand"
	"encoding/hex"
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

func createSession(username, role string) (string, error) { return "", nil }

func getSession(r *http.Request) (*Session, error) { return nil, nil }

func clearSession(w http.ResponseWriter, r *http.Request) error { return nil }

func requireAuth(next http.HandlerFunc) http.HandlerFunc { return nil }

func requireRole(role string, next http.HandlerFunc) http.HandlerFunc { return nil }
