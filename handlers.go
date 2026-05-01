package main

import (
	"html/template"
	"log"
	"net/http"
)

func home(w http.ResponseWriter, r *http.Request) {
	files := []string{
		"./templates/base.html",
	}

	ts, err := template.ParseFiles(files...)
	if err != nil {
		log.Print(err.Error())
	}

	err = ts.Execute(w, nil)
	if err != nil {
		log.Print(err.Error())
	}
}
