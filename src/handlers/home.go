package handlers

import (
	"net/http"

	"github.com/mr-flannery/go-recipe-book/src/auth"
	"github.com/mr-flannery/go-recipe-book/src/templates"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Get user info from context and pass directly to template
	userInfo := auth.GetUserInfoFromContext(r.Context())
	err := templates.Templates.ExecuteTemplate(w, "home.gohtml", userInfo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func ImprintHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// Get user info from context and pass directly to template
	userInfo := auth.GetUserInfoFromContext(r.Context())
	err := templates.Templates.ExecuteTemplate(w, "imprint.gohtml", userInfo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
