package main

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"

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

	err = ts.ExecuteTemplate(w, "base", nil)
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

	if r.Header.Get("HX-Request") == "true" {
		err = ts.ExecuteTemplate(w, "content", nil)
	} else {
		err = ts.ExecuteTemplate(w, "base", nil)
	}

	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Print("Login page served")

}

func login(w http.ResponseWriter, r *http.Request) {
	// TODO: Finish implementing login process
	// log.Print("Login POST request received. Process is not yet implemented.")

	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	username := r.FormValue("username")
	password := r.FormValue("password")

	// log.Printf("Login attempt for user: %s. Used password: %s", username, password)

	userFile, err := os.ReadFile("users.json")
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Print("Read users.json successfully")

	// log.Printf("Current contents of users.json: %s", string(userFile))

	var users UserFile
	err = json.Unmarshal(userFile, &users)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	log.Print("Unmarshaled users successfully")

	log.Print("Checking credentials against users")
	for _, user := range users.Users {
		if user.Username == username {
			// Login successful
			err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
			if err != nil {
				log.Printf("Failed login attempt for user %s: %v", user.Username, err)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			log.Printf("Successfully logged in user %s", user.Username)
			if err := createSession(w, user); err != nil {
				log.Printf("Error creating session for user %s: %v", user.Username, err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			return
		}
	}

	log.Printf("User %s does not exist", username)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)
}

func getIndividualSchedule(w http.ResponseWriter, r *http.Request) {
	log.Print("GET /schedule request received. Process is not yet implemented.")
}

func submitSchedule(w http.ResponseWriter, r *http.Request) {
	log.Print("POST /schedule request received. Process is not yet implemented.")
}
