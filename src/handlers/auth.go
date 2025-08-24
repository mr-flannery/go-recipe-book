package handlers

import (
	"html/template"
	"net/http"

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
	email := r.FormValue("email")
	password := r.FormValue("password")
	redirectURL := r.FormValue("redirect")

	// Authenticate user
	user, err := auth.Authenticate(email, password)
	if err != nil {
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

	// Authentication successful - create secure session
	clientIP := auth.GetClientIP(r)
	userAgent := r.UserAgent()

	session, err := auth.CreateSession(user.ID, clientIP, userAgent)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	auth.SetSecureSessionCookie(w, session.ID)

	// Determine redirect URL
	finalRedirectURL := "/"
	if redirectURL != "" {
		finalRedirectURL = redirectURL
	}

	// regardless of whether the request has been made with htmx or not
	// we always use a normal redirect
	http.Redirect(w, r, finalRedirectURL, http.StatusSeeOther)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	// Get session ID from cookie and invalidate it
	sessionID, err := auth.GetSessionFromRequest(r)
	if err == nil {
		// Invalidate session in database
		auth.InvalidateSession(sessionID)
	}

	// Clear session cookie
	auth.ClearSessionCookie(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
