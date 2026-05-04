package main

import (
	"log"
	"net/http"
)

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", home)
	mux.HandleFunc("GET /login", loginPage)
	mux.HandleFunc("POST /login", login)
	mux.HandleFunc("GET /schedule", getIndividualSchedule)
	mux.HandleFunc("POST /schedule", submitSchedule)

	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	log.Print("Starting server on :8081")

	err := http.ListenAndServe(":8081", mux)
	if err != nil {
		log.Fatal(err)
	}
}
