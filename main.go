package main

import (
	"log"
	"net/http"
)

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", home)

	log.Print("Starting server on :8081")

	err := http.ListenAndServe(":8081", mux)
	if err != nil {
		log.Fatal(err)
	}
}
