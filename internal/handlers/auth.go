package handlers

import (
	"net/http"
	
	"github.com/yourusername/agent-coding-recipe-book/auth"
)

// LoginHandler renders the login form and handles login POSTs
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(`<form method="POST"><input name="username" placeholder="Username"><input name="password" type="password" placeholder="Password"><button type="submit">Login</button></form>`))
		return
	}
	if r.Method == http.MethodPost {
		r.ParseForm()
		username := r.FormValue("username")
		password := r.FormValue("password")
		if auth.Authenticate(username, password) {
			auth.SetSession(w, username)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		w.Write([]byte("<p>Invalid credentials</p>"))
	}
}

// LogoutHandler logs the user out
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{Name: "session", Value: "", Path: "/", MaxAge: -1}
	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
