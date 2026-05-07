package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
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
		return
	}

	templateData := buildTemplateData(r)
	err = ts.ExecuteTemplate(w, "base", templateData)
	if err != nil {
		log.Print(err.Error())
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

	scheduleSubmission := buildScheduleSubmission(user, slots, existing, time.Now())

	valid := validateSchedule(scheduleSubmission)

	if len(valid.Errors) != 0 {
		log.Print(valid.Errors)
		// allErrors := strings.Join(valid.Errors, "\n")
		// http.Error(w, allErrors, http.StatusBadRequest)

		templateData := buildTemplateData(r)
		templateData.CurrentScheduleStatus = scheduleSubmission.Status
		templateData.ValidationErrors = valid.Errors
		templateData.WeekView = buildWeekViewData(slots, false)

		if existing != nil {
			templateData.CurrentSubmission = existing
			templateData.CurrentScheduleStatus = existing.Status
		}

		w.WriteHeader(http.StatusBadRequest)
		err = ts.ExecuteTemplate(w, "content", templateData)
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
	return

}

func adminPage(w http.ResponseWriter, r *http.Request) {
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

	ts, err := template.New("base").Funcs(funcMap).ParseFiles(files...)
	if err != nil {
		log.Print(err.Error())
		return
	}
	log.Print("Files parsed, building template data and trying to build template")

	templateData := buildTemplateData(r)
	templateData.WeekView = buildWeekViewData([]TimeSlot{}, true)

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

	log.Print("Successfully built the page")
}
