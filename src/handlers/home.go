package handlers

import (
	"net/http"

	"github.com/mr-flannery/go-recipe-book/src/auth"
)

func (h *Handler) HomeHandler(w http.ResponseWriter, r *http.Request) {
	userInfo := auth.GetUserInfoFromContext(r.Context())
	h.Renderer.RenderPage(w, "home.gohtml", userInfo)
}

func (h *Handler) ImprintHandler(w http.ResponseWriter, r *http.Request) {
	userInfo := auth.GetUserInfoFromContext(r.Context())
	h.Renderer.RenderPage(w, "imprint.gohtml", userInfo)
}
