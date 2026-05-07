package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSubmitScheduleTwiceForSameUserKeepsOneRecord(t *testing.T) {
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	tempDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(tempDir, "schedules.json"), []byte(`{"schedules":[]}`), 0644); err != nil {
		t.Fatalf("seed schedules.json: %v", err)
	}

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("change working directory: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})

	sessionID := "test-session-id"
	expiresAt := time.Now().Add(time.Hour)

	sessionsMu.Lock()
	previousSessions := sessions
	sessions = map[string]Session{
		sessionID: {
			Username:  "student1",
			Role:      "user",
			ExpiresAt: expiresAt,
		},
	}
	sessionsMu.Unlock()

	t.Cleanup(func() {
		sessionsMu.Lock()
		sessions = previousSessions
		sessionsMu.Unlock()
	})

	firstPayload := buildSelectedSlotsJSON(t, "08", "12")
	secondPayload := buildSelectedSlotsJSON(t, "09", "13")

	submitScheduleRequest(t, sessionID, firstPayload)
	submitScheduleRequest(t, sessionID, secondPayload)

	storedSchedules, err := loadSchedules()
	if err != nil {
		t.Fatalf("load stored schedules: %v", err)
	}

	if len(storedSchedules.Schedules) != 1 {
		t.Fatalf("expected 1 stored schedule, got %d", len(storedSchedules.Schedules))
	}

	stored := storedSchedules.Schedules[0]
	if stored.Username != "student1" {
		t.Fatalf("expected stored schedule for student1, got %s", stored.Username)
	}

	if stored.SubmittedAt == "" {
		t.Fatal("expected SubmittedAt to be preserved")
	}

	if stored.UpdatedAt == "" {
		t.Fatal("expected UpdatedAt to be set")
	}

	if len(stored.Slots) != 120 {
		t.Fatalf("expected 120 stored slots, got %d", len(stored.Slots))
	}

	firstStoredSlot := stored.Slots[0]
	if firstStoredSlot.Day != "Mon" || firstStoredSlot.Hour != 9 || firstStoredSlot.Slice != 0 {
		t.Fatalf("expected second submission to replace slots, got %+v", firstStoredSlot)
	}
}

func submitScheduleRequest(t *testing.T, sessionID, selectedSlots string) {
	t.Helper()

	form := url.Values{}
	form.Set("selected-slots", selectedSlots)

	req := httptest.NewRequest(http.MethodPost, "/schedule", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{
		Name:  SESSION_COOKIE_NAME,
		Value: sessionID,
	})

	recorder := httptest.NewRecorder()
	submitSchedule(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d with body %q", recorder.Code, recorder.Body.String())
	}
}

func buildSelectedSlotsJSON(t *testing.T, startHour, endHour string) string {
	t.Helper()

	days := []string{"Mon", "Tue", "Wed", "Thu", "Fri"}
	keys := make([]string, 0, 120)

	for _, day := range days {
		for hour := startHour; hour < endHour; hour = incrementHour(hour) {
			for slice := 0; slice < 6; slice++ {
				keys = append(keys, `"`+day+`|`+hour+`|`+string(rune('0'+slice))+`"`)
			}
		}
	}

	return "[" + strings.Join(keys, ",") + "]"
}

func incrementHour(hour string) string {
	switch hour {
	case "08":
		return "09"
	case "09":
		return "10"
	case "10":
		return "11"
	case "11":
		return "12"
	case "12":
		return "13"
	default:
		return ""
	}
}
