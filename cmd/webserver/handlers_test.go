package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestRejectedScheduleResubmissionUpdatesExistingRecord(t *testing.T) {
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	tempDir := t.TempDir()
	firstPayload := buildSelectedSlotsJSON(t, "08", "12")
	secondPayload := buildSelectedSlotsJSON(t, "09", "13")
	firstScheduleSlots := marshalScheduleSlots(t, firstPayload)

	initialSchedules := `{"schedules":[{"username":"student1","slots":` + firstScheduleSlots + `,"status":"rejected","rejection_comment":"Needs more weekday coverage","submitted_at":"2026-05-01T09:00:00Z","updated_at":"2026-05-01T09:00:00Z"}]}`

	if err := os.WriteFile(filepath.Join(tempDir, "schedules.json"), []byte(initialSchedules), 0644); err != nil {
		t.Fatalf("seed schedules.json: %v", err)
	}
	seedScheduleTemplates(t, tempDir)

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

	if stored.Status != StatusPending {
		t.Fatalf("expected resubmission to return to pending, got %s", stored.Status)
	}

	if stored.RejectionComment != "" {
		t.Fatalf("expected rejection comment to be cleared, got %q", stored.RejectionComment)
	}

	if len(stored.Slots) != 120 {
		t.Fatalf("expected 120 stored slots, got %d", len(stored.Slots))
	}

	firstStoredSlot := stored.Slots[0]
	if firstStoredSlot.Day != "Mon" || firstStoredSlot.Hour != 9 || firstStoredSlot.Slice != 0 {
		t.Fatalf("expected second submission to replace slots, got %+v", firstStoredSlot)
	}
}

func TestAdminPageShowsOnlyPendingSchedulesAndSelectsFirstByDefault(t *testing.T) {
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	tempDir := t.TempDir()
	morningSlots := marshalScheduleSlots(t, buildSelectedSlotsJSON(t, "08", "12"))
	afternoonSlots := marshalScheduleSlots(t, buildSelectedSlotsJSON(t, "09", "13"))

	initialSchedules := `{"schedules":[` +
		`{"username":"student1","slots":` + morningSlots + `,"status":"pending","submitted_at":"2026-05-01T09:00:00Z","updated_at":"2026-05-01T09:00:00Z"},` +
		`{"username":"student2","slots":` + afternoonSlots + `,"status":"pending","submitted_at":"2026-05-02T09:00:00Z","updated_at":"2026-05-02T09:00:00Z"},` +
		`{"username":"student3","slots":` + morningSlots + `,"status":"approved","submitted_at":"2026-05-03T09:00:00Z","updated_at":"2026-05-03T09:00:00Z"},` +
		`{"username":"student4","slots":` + afternoonSlots + `,"status":"rejected","submitted_at":"2026-05-04T09:00:00Z","updated_at":"2026-05-04T09:00:00Z"}` +
		`]}`

	if err := os.WriteFile(filepath.Join(tempDir, "schedules.json"), []byte(initialSchedules), 0644); err != nil {
		t.Fatalf("seed schedules.json: %v", err)
	}
	seedScheduleTemplates(t, tempDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("change working directory: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})

	recorder := adminPageRequest(t, "/admin", false)
	body := recorder.Body.String()

	if !strings.Contains(body, `pending-user="student1"`) {
		t.Fatalf("expected first pending user in list, got body %q", body)
	}

	if !strings.Contains(body, `pending-user="student2"`) {
		t.Fatalf("expected second pending user in list, got body %q", body)
	}

	if strings.Contains(body, `pending-user="student3"`) || strings.Contains(body, `pending-user="student4"`) {
		t.Fatalf("expected only pending users in list, got body %q", body)
	}

	if !strings.Contains(body, `selected: student1`) {
		t.Fatalf("expected first pending schedule to be selected by default, got body %q", body)
	}
}

func TestAdminPageSelectedScheduleRendersReadOnly(t *testing.T) {
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	tempDir := t.TempDir()
	morningSlots := marshalScheduleSlots(t, buildSelectedSlotsJSON(t, "08", "12"))
	afternoonSlots := marshalScheduleSlots(t, buildSelectedSlotsJSON(t, "09", "13"))

	initialSchedules := `{"schedules":[` +
		`{"username":"student1","slots":` + morningSlots + `,"status":"pending","submitted_at":"2026-05-01T09:00:00Z","updated_at":"2026-05-01T09:00:00Z"},` +
		`{"username":"student2","slots":` + afternoonSlots + `,"status":"pending","submitted_at":"2026-05-02T09:00:00Z","updated_at":"2026-05-02T09:00:00Z"}` +
		`]}`

	if err := os.WriteFile(filepath.Join(tempDir, "schedules.json"), []byte(initialSchedules), 0644); err != nil {
		t.Fatalf("seed schedules.json: %v", err)
	}
	seedScheduleTemplates(t, tempDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("change working directory: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})

	recorder := adminPageRequest(t, "/admin?username=student2", false)
	body := recorder.Body.String()

	if !strings.Contains(body, `selected: student2`) {
		t.Fatalf("expected selected pending schedule details for student2, got body %q", body)
	}

	if !strings.Contains(body, `readOnly=true`) {
		t.Fatalf("expected admin week view to render in read-only mode, got body %q", body)
	}
}

func TestAdminPageShowsEmptyStateWhenNoPendingSchedulesExist(t *testing.T) {
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	tempDir := t.TempDir()
	morningSlots := marshalScheduleSlots(t, buildSelectedSlotsJSON(t, "08", "12"))

	initialSchedules := `{"schedules":[` +
		`{"username":"student3","slots":` + morningSlots + `,"status":"approved","submitted_at":"2026-05-03T09:00:00Z","updated_at":"2026-05-03T09:00:00Z"},` +
		`{"username":"student4","slots":` + morningSlots + `,"status":"rejected","submitted_at":"2026-05-04T09:00:00Z","updated_at":"2026-05-04T09:00:00Z"}` +
		`]}`

	if err := os.WriteFile(filepath.Join(tempDir, "schedules.json"), []byte(initialSchedules), 0644); err != nil {
		t.Fatalf("seed schedules.json: %v", err)
	}
	seedScheduleTemplates(t, tempDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("change working directory: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})

	recorder := adminPageRequest(t, "/admin", false)
	body := recorder.Body.String()

	if !strings.Contains(body, `No pending schedules to review.`) {
		t.Fatalf("expected empty state when no pending schedules exist, got body %q", body)
	}
}

func TestAdminPageHTMXReturnsOnlyDetailFragment(t *testing.T) {
	originalWD, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}

	tempDir := t.TempDir()
	morningSlots := marshalScheduleSlots(t, buildSelectedSlotsJSON(t, "08", "12"))
	afternoonSlots := marshalScheduleSlots(t, buildSelectedSlotsJSON(t, "09", "13"))

	initialSchedules := `{"schedules":[` +
		`{"username":"student1","slots":` + morningSlots + `,"status":"pending","submitted_at":"2026-05-01T09:00:00Z","updated_at":"2026-05-01T09:00:00Z"},` +
		`{"username":"student2","slots":` + afternoonSlots + `,"status":"pending","submitted_at":"2026-05-02T09:00:00Z","updated_at":"2026-05-02T09:00:00Z"}` +
		`]}`

	if err := os.WriteFile(filepath.Join(tempDir, "schedules.json"), []byte(initialSchedules), 0644); err != nil {
		t.Fatalf("seed schedules.json: %v", err)
	}
	seedScheduleTemplates(t, tempDir)

	if err := os.Chdir(tempDir); err != nil {
		t.Fatalf("change working directory: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(originalWD)
	})

	recorder := adminPageRequest(t, "/admin?username=student2", true)
	body := recorder.Body.String()

	if !strings.Contains(body, `id="review-detail"`) {
		t.Fatalf("expected HTMX response to include detail fragment, got body %q", body)
	}

	if strings.Contains(body, `id="approval-container"`) {
		t.Fatalf("expected HTMX response to exclude full page shell, got body %q", body)
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

func adminPageRequest(t *testing.T, path string, htmx bool) *httptest.ResponseRecorder {
	t.Helper()

	req := httptest.NewRequest(http.MethodGet, path, nil)
	if htmx {
		req.Header.Set("HX-Request", "true")
	}

	recorder := httptest.NewRecorder()
	adminPage(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d with body %q", recorder.Code, recorder.Body.String())
	}

	return recorder
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

func marshalScheduleSlots(t *testing.T, selectedSlots string) string {
	t.Helper()

	slots, err := parseSelectedSlots(selectedSlots)
	if err != nil {
		t.Fatalf("parse selected slots for seed data: %v", err)
	}

	encoded, err := json.Marshal(slots)
	if err != nil {
		t.Fatalf("marshal seeded schedule slots: %v", err)
	}

	return string(encoded)
}

func seedScheduleTemplates(t *testing.T, root string) {
	t.Helper()

	dirs := []string{
		filepath.Join(root, "templates"),
		filepath.Join(root, "components"),
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("create test template dir %s: %v", dir, err)
		}
	}

	files := map[string]string{
		filepath.Join(root, "templates", "base.html"):       `{{define "base"}}{{template "content" .}}{{end}}`,
		filepath.Join(root, "components", "navbar.html"):    `{{define "nav"}}{{end}}`,
		filepath.Join(root, "templates", "week_view.html"):  `{{define "week_view"}}<div>readOnly={{.WeekView.ReadOnly}} total={{.WeekView.WeeklyTotalMinutes}}</div>{{end}}`,
		filepath.Join(root, "templates", "schedule.html"):   `{{define "content"}}{{template "week_view" .}}{{end}}`,
		filepath.Join(root, "templates", "approval.html"):   `{{define "content"}}<div id="approval-container">{{template "admin_review_panel" .}}</div>{{end}}{{define "admin_review_panel"}}{{if .PendingSchedules}}{{range .PendingSchedules}}<span pending-user="{{.Username}}">{{.Username}}</span>{{end}}{{end}}{{template "admin_review_detail" .}}{{end}}{{define "admin_review_detail"}}<div id="review-detail">{{if .SelectedPendingSchedule}}selected: {{.SelectedPendingSchedule.Username}} {{template "week_view" .}}{{else}}No pending schedules to review.{{end}}</div>{{end}}`,
	}

	for path, contents := range files {
		if err := os.WriteFile(path, []byte(contents), 0644); err != nil {
			t.Fatalf("seed template %s: %v", path, err)
		}
	}
}
