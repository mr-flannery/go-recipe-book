package handlers

import (
	"net/http"

	"github.com/mr-flannery/go-recipe-book/src/auth"
	"github.com/mr-flannery/go-recipe-book/src/config"
)

func (h *Handler) HomeHandler(w http.ResponseWriter, r *http.Request) {
	userInfo := auth.GetUserInfoFromContext(r.Context())
	h.Renderer.RenderPage(w, "home.gohtml", userInfo)
}

type ImprintData struct {
	*auth.UserInfo
	Name    string
	Address string
	Email   string
	Phone   string
}

func (h *Handler) ImprintHandler(w http.ResponseWriter, r *http.Request) {
	userInfo := auth.GetUserInfoFromContext(r.Context())
	cfg := config.GetConfig()
	data := ImprintData{
		UserInfo: userInfo,
		Name:     cfg.Imprint.Name,
		Address:  cfg.Imprint.Address,
		Email:    cfg.Imprint.Email,
		Phone:    cfg.Imprint.Phone,
	}
	h.Renderer.RenderPage(w, "imprint.gohtml", data)
}

type PrivacyData struct {
	*auth.UserInfo
	ContactEmail string
}

func (h *Handler) PrivacyHandler(w http.ResponseWriter, r *http.Request) {
	userInfo := auth.GetUserInfoFromContext(r.Context())
	cfg := config.GetConfig()
	data := PrivacyData{
		UserInfo:     userInfo,
		ContactEmail: cfg.Imprint.Email,
	}
	h.Renderer.RenderPage(w, "privacy.gohtml", data)
}
