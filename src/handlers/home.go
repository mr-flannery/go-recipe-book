package handlers

import (
	"html/template"
	"net/http"

	"github.com/mr-flannery/go-recipe-book/src/auth"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	username, isLoggedIn := auth.GetUser(r) // Check if the user is logged in
	data := struct {
		IsLoggedIn bool
		Username   string
	}{
		IsLoggedIn: isLoggedIn,
		Username:   username,
	}

	templates := template.Must(template.ParseGlob("templates/*.gohtml"))
	err := templates.ExecuteTemplate(w, "home.gohtml", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
