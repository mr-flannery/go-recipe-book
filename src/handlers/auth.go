package handlers

import (
	"html/template"
	"net/http"
	"time"

	"github.com/mr-flannery/go-recipe-book/src/auth"
)

func LoginHandler(w http.ResponseWriter, r *http.Request) {
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
