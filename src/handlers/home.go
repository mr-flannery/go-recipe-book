package handlers

import (
	"net/http"

	"github.com/mr-flannery/go-recipe-book/src/auth"
	"github.com/mr-flannery/go-recipe-book/src/templates"
	"github.com/mr-flannery/go-recipe-book/src/utils"
)

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	theme := utils.GetThemeFromRequest(r)
	templateName := utils.GetThemedTemplateName("home.gohtml", theme)

	userInfo := auth.GetUserInfoFromContext(r.Context())
	err := templates.Templates.ExecuteTemplate(w, templateName, userInfo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func ImprintHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	theme := utils.GetThemeFromRequest(r)
	templateName := utils.GetThemedTemplateName("imprint.gohtml", theme)

	userInfo := auth.GetUserInfoFromContext(r.Context())
	err := templates.Templates.ExecuteTemplate(w, templateName, userInfo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
