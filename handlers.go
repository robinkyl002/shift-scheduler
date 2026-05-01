package main

import (
	"html/template"
	"log"
	"net/http"
)

func home(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"./templates/base.html",
		"./components/navbar.html",
		"./templates/home.html",
		"./static/css/home.css",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Print(err.Error())
		return
	}

	err = ts.Execute(w, nil)
	if err != nil {
		log.Print(err.Error())
		return
	}
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

	err = ts.ExecuteTemplate(w, "base", nil)
	if err != nil {
		log.Print(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

}

func login(w http.ResponseWriter, r *http.Request) {
	// TODO: Finish implementing login process
	log.Print("Login POST request received. Process is not yet implemented.")
}
