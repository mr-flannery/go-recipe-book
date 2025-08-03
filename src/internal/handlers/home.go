package handlers

import (
	"html/template"
	"net/http"
)

var templates = template.Must(template.ParseGlob("templates/*.gohtml"))

func Home(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := templates.ExecuteTemplate(w, "home.gohtml", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
