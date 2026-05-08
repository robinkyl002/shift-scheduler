package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func home(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"./templates/base.html",
		"./components/navbar.html",
		"./templates/home.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	templateData := buildTemplateData(r)
	err = ts.ExecuteTemplate(w, "base", templateData)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Print("Home page served")
}

func loginPage(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"./templates/base.html",
		"./components/navbar.html",
		"./templates/login.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	templateData := buildTemplateData(r)

	if r.Header.Get("HX-Request") == "true" {
		err = ts.ExecuteTemplate(w, "content", templateData)
	} else {
		err = ts.ExecuteTemplate(w, "base", templateData)
	}

	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

}

func login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	userFile, err := os.ReadFile("users.json")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var users UserFile
	err = json.Unmarshal(userFile, &users)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	for _, user := range users.Users {
		if user.Username == username {
			// Login successful
			err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
			if err != nil {
				log.Printf("Failed login attempt for user %s: %v", user.Username, err)
				http.Error(w, "Invalid username or password", http.StatusUnauthorized)
				return
			}

			log.Printf("Successfully logged in user %s", user.Username)
			if err := createSession(w, user); err != nil {
				log.Printf("Error creating session for user %s: %v", user.Username, err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			redirectPath := defaultLandingPath(user.Role)

			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Redirect", redirectPath)
				w.WriteHeader(http.StatusOK)
				return
			}
			http.Redirect(w, r, redirectPath, http.StatusSeeOther)
			return
		}
	}

	log.Printf("User %s does not exist", username)
	http.Error(w, "Invalid username or password", http.StatusUnauthorized)
}

func logout(w http.ResponseWriter, r *http.Request) {
	clearSession(w, r)
	if r.Header.Get("HX-Request") == "true" {
		w.Header().Set("HX-Redirect", "/login")
		w.WriteHeader(http.StatusOK)
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func getIndividualSchedule(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"./templates/base.html",
		"./components/navbar.html",
		"./templates/schedule.html",
		"./templates/week_view.html",
	}

	funcMap := template.FuncMap{
		"formatHour":    formatHour,
		"formatMinutes": formatMinutes,
	}

	ts, err := template.New("base").Funcs(funcMap).ParseFiles(files...)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	templateData := buildTemplateData(r)

	session, _, err := getSession(r)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	slots := []TimeSlot{}
	readOnly := false

	schedules, err := loadSchedules()
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	submission, found := findScheduleByUsername(schedules.Schedules, session.Username)

	if found {
		templateData.CurrentSubmission = submission
		templateData.CurrentScheduleStatus = submission.Status
		slots = submission.Slots
		switch submission.Status {
		case StatusApproved, StatusPending:
			readOnly = true
		default:
			readOnly = false
		}
	}

	templateData.WeekView = buildWeekViewData(slots, readOnly)

	if r.Header.Get("HX-Request") == "true" {
		err = ts.ExecuteTemplate(w, "content", templateData)
	} else {
		err = ts.ExecuteTemplate(w, "base", templateData)
	}

	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func submitSchedule(w http.ResponseWriter, r *http.Request) {
	session, _, err := getSession(r)

	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal ServerError", http.StatusInternalServerError)
		return
	}

	user := session.Username

	files := []string{
		"./templates/base.html",
		"./components/navbar.html",
		"./templates/schedule.html",
		"./templates/week_view.html",
	}

	funcMap := template.FuncMap{
		"formatHour":    formatHour,
		"formatMinutes": formatMinutes,
	}

	ts, err := template.New("base").Funcs(funcMap).ParseFiles(files...)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var slots []TimeSlot
	raw := r.FormValue("selected-slots")

	slots, err = parseSelectedSlots(raw)

	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Prevent race conditions
	scheduleMu.Lock()
	defer scheduleMu.Unlock()

	schedules, err := loadSchedules()
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Could not load stored schedules", http.StatusInternalServerError)
		return
	}

	existingIndex := findScheduleIndex(schedules.Schedules, user)

	var existing *ScheduleSubmission
	if existingIndex >= 0 {
		existing = &schedules.Schedules[existingIndex]
	}

	if existing != nil && (existing.Status == StatusPending || existing.Status == StatusApproved) {
		templateData := buildTemplateData(r)
		templateData.CurrentSubmission = existing
		templateData.CurrentScheduleStatus = existing.Status
		templateData.WeekView = buildWeekViewData(existing.Slots, true)
		templateData.ValidationErrors = []string{"This schedule can no longer be edited."}

		_ = ts.ExecuteTemplate(w, "content", templateData)
		return
	}

	scheduleSubmission := buildScheduleSubmission(user, slots, existing, time.Now())

	valid := validateSchedule(scheduleSubmission)

	if len(valid.Errors) != 0 {
		log.Print(valid.Errors)

		templateData := buildTemplateData(r)
		templateData.ValidationErrors = valid.Errors
		templateData.WeekView = buildWeekViewData(slots, false)

		if existing != nil {
			templateData.CurrentSubmission = existing
			templateData.CurrentScheduleStatus = existing.Status
		}

		err = ts.ExecuteTemplate(w, "content", templateData)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		return
	}

	upsertScheduleAtIndex(&schedules, existingIndex, scheduleSubmission)

	err = saveSchedules(schedules)

	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Could not write data to file", http.StatusInternalServerError)
		return
	}

	templateData := buildTemplateData(r)
	templateData.CurrentSubmission = &scheduleSubmission
	templateData.CurrentScheduleStatus = StatusPending
	templateData.WeekView = buildWeekViewData(scheduleSubmission.Slots, true)
	templateData.ValidationErrors = nil

	err = ts.ExecuteTemplate(w, "content", templateData)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func buildAdminReviewTemplateData(r *http.Request, selectedUsername string) (TemplateData, error) {
	templateData := buildTemplateData(r)

	schedules, err := loadSchedules()
	if err != nil {
		return TemplateData{}, err
	}

	pendingSchedules := make([]ScheduleSubmission, 0)

	for _, schedule := range schedules.Schedules {
		if schedule.Status == StatusPending {
			pendingSchedules = append(pendingSchedules, schedule)
		}
	}

	scheduleSummaries := make([]PendingScheduleSummary, 0)
	for _, schedule := range pendingSchedules {
		scheduleSummaries = append(scheduleSummaries, PendingScheduleSummary{
			Username:    schedule.Username,
			SubmittedAt: schedule.SubmittedAt,
			Status:      schedule.Status,
		})
	}

	templateData.PendingSchedules = scheduleSummaries
	var selected *ScheduleSubmission

	if selectedUsername != "" {
		for i := range pendingSchedules {
			if pendingSchedules[i].Username == selectedUsername {
				selected = &pendingSchedules[i]
				break
			}
		}
	}

	if selected == nil && len(pendingSchedules) > 0 {
		selected = &pendingSchedules[0]
	}

	templateData.SelectedPendingSchedule = selected
	if selected != nil {
		templateData.WeekView = buildWeekViewData(selected.Slots, true)
	}

	return templateData, nil
}

func parseAdminTemplates() (*template.Template, error) {
	files := []string{
		"./templates/approval.html",
		"./templates/base.html",
		"./components/navbar.html",
		"./templates/week_view.html",
	}

	log.Print("Attempting to parse files")

	funcMap := template.FuncMap{
		"formatHour":    formatHour,
		"formatMinutes": formatMinutes,
	}

	return template.New("base").Funcs(funcMap).ParseFiles(files...)
}

func adminPage(w http.ResponseWriter, r *http.Request) {
	ts, err := parseAdminTemplates()
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	log.Print("Files parsed, building template data and trying to build template")

	templateData, err := buildAdminReviewTemplateData(r, r.URL.Query().Get("username"))
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if r.Header.Get("HX-Request") == "true" {
		err = ts.ExecuteTemplate(w, "admin_review_panel", templateData)
	} else {
		err = ts.ExecuteTemplate(w, "base", templateData)
	}

	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Print("Successfully built the page")
}

func invalidReviewFields(username string, status string, comment string) []string {
	errors := make([]string, 0)
	if status != "approved" && status != "rejected" {
		errors = append(errors, "Status must be approved or rejected")
	}

	if status == "rejected" && comment == "" {
		errors = append(errors, "A rejection comment is required.")
	}

	if username == "" {
		errors = append(errors, "Username may not be empty")
	}

	return errors
}

func submitScheduleReview(w http.ResponseWriter, r *http.Request) {
	ts, err := parseAdminTemplates()
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	username := r.FormValue("username")
	status := r.FormValue("status")
	comment := strings.TrimSpace(r.FormValue("rejection_comment"))

	reviewErrors := invalidReviewFields(username, status, comment)
	if len(reviewErrors) > 0 {
		templateData, err := buildAdminReviewTemplateData(r, username)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		templateData.AdminReviewErrors = reviewErrors
		templateData.AdminReviewComment = comment
		templateData.ShowRejectFields = true

		err = ts.ExecuteTemplate(w, "admin_review_panel", templateData)
		if err != nil {
			log.Print(err.Error())
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		return

	}

	scheduleMu.Lock()
	defer scheduleMu.Unlock()
	schedules, err := loadSchedules()
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	log.Print(schedules)

	// schedule, found := findScheduleByUsername(schedules.Schedules, username)
}
