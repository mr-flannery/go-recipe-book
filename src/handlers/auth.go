package handlers

import (
	"html/template"
	"net/http"
	"time"

	"github.com/mr-flannery/go-recipe-book/src/auth"
)

type LoginData struct {
	RedirectURL string
	Error       string
}

func GetLoginHandler(w http.ResponseWriter, r *http.Request) {
	templates := template.Must(template.ParseGlob("templates/*.gohtml"))
	
	redirectURL := r.URL.Query().Get("redirect")
	data := LoginData{
		RedirectURL: redirectURL,
	}
	
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := templates.ExecuteTemplate(w, "login.gohtml", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func PostLoginHandler(w http.ResponseWriter, r *http.Request) {
	templates := template.Must(template.ParseGlob("templates/*.gohtml"))
	
	r.ParseForm()
	username := r.FormValue("username")
	password := r.FormValue("password")
	redirectURL := r.FormValue("redirect")

	ok := auth.Authenticate(username, password)

	if !ok {
		// Check if this is an HTMX request
		if r.Header.Get("HX-Request") == "true" {
			// Return just the form with error message for HTMX
			data := LoginData{
				RedirectURL: redirectURL,
				Error:       "Invalid username or password. Please try again.",
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			err := templates.ExecuteTemplate(w, "login-form", data)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		} else {
			// Regular form submission - redirect back to login with error
			data := LoginData{
				RedirectURL: redirectURL,
				Error:       "Invalid username or password. Please try again.",
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			err := templates.ExecuteTemplate(w, "login.gohtml", data)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
			return
		}
	}

	// Authentication successful
	auth.SetSession(w, username)
	
	// Determine redirect URL
	finalRedirectURL := "/"
	if redirectURL != "" {
		finalRedirectURL = redirectURL
	}
	
	// Check if this is an HTMX request
	if r.Header.Get("HX-Request") == "true" {
		// For HTMX, send a redirect header
		w.Header().Set("HX-Redirect", finalRedirectURL)
		w.WriteHeader(http.StatusOK)
		return
	} else {
		// Regular form submission
		http.Redirect(w, r, finalRedirectURL, http.StatusSeeOther)
		return
	}
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:    "session",
		Value:   "",
		Path:    "/",
		MaxAge:  -1,
		Expires: time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC),
	}
	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
