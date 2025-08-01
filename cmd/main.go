package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/yourusername/agent-coding-recipe-book/auth"
)

// importHandlersLogin calls the Login handler from internal/handlers
func importHandlersLogin(w http.ResponseWriter, r *http.Request) {
	templates := template.Must(template.ParseGlob("templates/*.gohtml"))
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err := templates.ExecuteTemplate(w, "login.gohtml", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	if r.Method == http.MethodPost {
		r.ParseForm()
		username := r.FormValue("username")
		password := r.FormValue("password")

		ok := auth.Authenticate(username, password)

		if !ok {
			w.Write([]byte("<p>Invalid credentials</p>"))
			return
		}

		auth.SetSession(w, username)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
}

// importHandlersLogout logs the user out
func importHandlersLogout(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{Name: "session", Value: "", Path: "/", MaxAge: -1, Expires: time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)}
	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func main() {
	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}
	log.Printf("Starting server on %s", addr)

	// Home page
	http.HandleFunc("/", importHandlersHome)

	// Auth routes
	http.HandleFunc("/login", importHandlersLogin)
	http.HandleFunc("/logout", importHandlersLogout)

	log.Fatal(http.ListenAndServe(addr, nil))
}

// importHandlersHome calls the Home handler from internal/handlers
func importHandlersHome(w http.ResponseWriter, r *http.Request) {
	// Import the Home handler from the handlers package
	// This import path assumes your module name is github.com/yourusername/agent-coding-recipe-book
	// Adjust the import path if your module name is different
	//
	// import "github.com/yourusername/agent-coding-recipe-book/internal/handlers"
	// handlers.Home(w, r)

	// To avoid import issues in this patch, inline the handler logic for now
	// Replace this with a direct call to handlers.Home when possible
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
