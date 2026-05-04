package main

import "time"

type Session struct {
	Username  string
	Role      string
	ExpiresAt time.Time
}

var sessions = make(map[string]Session)
