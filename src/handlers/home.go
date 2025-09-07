package handlers

import (
	"html/template"
	"net/http"

	"github.com/mr-flannery/go-recipe-book/src/auth"
	"github.com/mr-flannery/go-recipe-book/src/db"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Get database connection
	database, err := db.GetConnection()
	if err != nil {
		// If DB fails, show page without user info
		data := struct {
			IsLoggedIn bool
			Username   string
			IsAdmin    bool
		}{
			IsLoggedIn: false,
			Username:   "",
			IsAdmin:    false,
		}
		templates := template.Must(template.ParseGlob("templates/*.gohtml"))
		err := templates.ExecuteTemplate(w, "home.gohtml", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	defer database.Close()

	// Check if user is logged in
	user, err := auth.GetUserBySession(database, r)
	isLoggedIn := err == nil
	username := ""
	if isLoggedIn {
		username = user.Username
	}

	isAdmin := false
	if isLoggedIn {
		isAdmin = user.IsAdmin
	}

	data := struct {
		IsLoggedIn bool
		Username   string
		IsAdmin    bool
	}{
		IsLoggedIn: isLoggedIn,
		Username:   username,
		IsAdmin:    isAdmin,
	}

	templates := template.Must(template.ParseGlob("templates/*.gohtml"))
	err = templates.ExecuteTemplate(w, "home.gohtml", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func ImprintHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Get database connection
	database, err := db.GetConnection()
	if err != nil {
		// If DB fails, show page without user info
		data := struct {
			IsLoggedIn bool
			Username   string
			IsAdmin    bool
		}{
			IsLoggedIn: false,
			Username:   "",
			IsAdmin:    false,
		}
		templates := template.Must(template.ParseGlob("templates/*.gohtml"))
		err := templates.ExecuteTemplate(w, "imprint.gohtml", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	defer database.Close()

	// Check if user is logged in
	user, err := auth.GetUserBySession(database, r)
	isLoggedIn := err == nil
	username := ""
	if isLoggedIn {
		username = user.Username
	}

	isAdmin := false
	if isLoggedIn {
		isAdmin = user.IsAdmin
	}

	data := struct {
		IsLoggedIn bool
		Username   string
		IsAdmin    bool
	}{
		IsLoggedIn: isLoggedIn,
		Username:   username,
		IsAdmin:    isAdmin,
	}

	templates := template.Must(template.ParseGlob("templates/*.gohtml"))
	err = templates.ExecuteTemplate(w, "imprint.gohtml", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
