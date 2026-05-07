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
	}

	funcMap := template.FuncMap{
		"formatHour":   formatHour,
		"sliceIndices": sliceIndices,
	}

	ts, err := template.New("base").Funcs(funcMap).ParseFiles(files...)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	templateData := buildTemplateData(r)
	templateData.Schedule = buildSchedulePageData()

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

	var slots []TimeSlot
	raw := r.FormValue("selected-slots")

	slots, err = parseSelectedSlots(raw)

	currDate := time.Now()

	scheduleSubmission := ScheduleSubmission{
		Username:         user,
		Slots:            slots,
		Status:           StatusPending,
		RejectionComment: "",
		SubmittedAt:      currDate.String(),
		UpdatedAt:        currDate.String(),
	}

	scheduleFile, err := os.ReadFile("schedules.json")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var schedules ScheduleFile
	err = json.Unmarshal(scheduleFile, &schedules)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	valid := validateSchedule(scheduleSubmission)

	if len(valid.Errors) != 0 {
		log.Print(valid.Errors)
		allErrors := strings.Join(valid.Errors, "\n")
		http.Error(w, allErrors, http.StatusBadRequest)

		return
	}

	if len(schedules.Schedules) == 0 {
		schedules.Schedules = append(schedules.Schedules, scheduleSubmission)
	} else {
		for i, schedule := range schedules.Schedules {
			if schedule.Username == scheduleSubmission.Username {
				schedules.Schedules[i] = scheduleSubmission
				break
			}
		}
		schedules.Schedules = append(schedules.Schedules, scheduleSubmission)
	}

	updated, err := json.MarshalIndent(schedules, "", "  ")
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "error marshalling JSON", http.StatusInternalServerError)
	}

	err = os.WriteFile("schedules.json", updated, 0644)

	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Coult not write data to file", http.StatusInternalServerError)
		return
	}
}

func adminPage(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"./templates/approval.html",
		"./templates/base.html",
		"./components/navbar.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Print(err.Error())
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
