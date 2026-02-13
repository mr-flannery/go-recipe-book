package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/mr-flannery/go-recipe-book/src/auth"
	"github.com/mr-flannery/go-recipe-book/src/config"
	"github.com/mr-flannery/go-recipe-book/src/mail"
	"github.com/mr-flannery/go-recipe-book/src/templates"
	"github.com/mr-flannery/go-recipe-book/src/utils"
)

type LoginData struct {
	RedirectURL string
	Error       string
	UserInfo    *auth.UserInfo
}

func GetLoginHandler(w http.ResponseWriter, r *http.Request) {
	redirectURL := r.URL.Query().Get("redirect")
	data := LoginData{
		RedirectURL: redirectURL,
		UserInfo:    auth.GetUserInfoFromContext(r.Context()),
	}

	theme := utils.GetThemeFromRequest(r)
	templateName := utils.GetThemedTemplateName("login.gohtml", theme)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := templates.Templates.ExecuteTemplate(w, templateName, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) PostLoginHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	email := r.FormValue("email")
	password := r.FormValue("password")
	redirectURL := r.FormValue("redirect")

	theme := utils.GetThemeFromRequest(r)
	templateName := utils.GetThemedTemplateName("login.gohtml", theme)

	user, err := auth.Authenticate(h.AuthStore, email, password)
	if err != nil {
		data := LoginData{
			RedirectURL: redirectURL,
			Error:       "Invalid username or password. Please try again.",
			UserInfo:    auth.GetUserInfoFromContext(r.Context()),
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err := templates.Templates.ExecuteTemplate(w, templateName, data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	clientIP := auth.GetClientIP(r)
	userAgent := r.UserAgent()

	session, err := auth.CreateSession(h.AuthStore, user.ID, clientIP, userAgent)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}

	auth.SetSecureSessionCookie(w, session.ID)

	finalRedirectURL := "/"
	if redirectURL != "" {
		finalRedirectURL = redirectURL
	} else {
		finalRedirectURL = utils.BuildURLWithTheme("/", theme)
	}

	http.Redirect(w, r, finalRedirectURL, http.StatusSeeOther)
}

func (h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	sessionID, err := auth.GetSessionFromRequest(r)
	if err == nil {
		auth.InvalidateSession(h.AuthStore, sessionID)
	}

	auth.ClearSessionCookie(w)
	theme := utils.GetThemeFromRequest(r)
	redirectURL := utils.BuildURLWithTheme("/", theme)
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

type RegisterData struct {
	Username string
	Email    string
	Error    string
	Success  string
	UserInfo *auth.UserInfo
}

func GetRegisterHandler(w http.ResponseWriter, r *http.Request) {
	data := RegisterData{
		UserInfo: auth.GetUserInfoFromContext(r.Context()),
	}

	theme := utils.GetThemeFromRequest(r)
	templateName := utils.GetThemedTemplateName("register.gohtml", theme)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := templates.Templates.ExecuteTemplate(w, templateName, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) PostRegisterHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	theme := utils.GetThemeFromRequest(r)
	templateName := utils.GetThemedTemplateName("register.gohtml", theme)

	data := RegisterData{
		Username: username,
		Email:    email,
		UserInfo: auth.GetUserInfoFromContext(r.Context()),
	}

	if password != confirmPassword {
		data.Error = "Passwords do not match"
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		templates.Templates.ExecuteTemplate(w, templateName, data)
		return
	}

	err := auth.CreateRegistrationRequest(h.AuthStore, username, email, password)
	if err != nil {
		slog.Error("Failed to create registration request", "error", err)
		data.Error = err.Error()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		templates.Templates.ExecuteTemplate(w, templateName, data)
		return
	}

	conf := config.GetConfig()
	approvalURL := fmt.Sprintf("http://localhost:8080/admin/registrations")
	err = mail.SendNewRegistrationNotification(conf.DB.Admin.Email, conf.DB.Admin.Username, username, email, approvalURL)
	if err != nil {
		slog.Error("Failed to send admin notification email", "error", err)
	}

	data.Success = "Registration request submitted successfully! An administrator will review your request and you will receive an email when it's approved."
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templates.Templates.ExecuteTemplate(w, templateName, data)
}

type PendingRegistrationsData struct {
	Registrations []auth.RegistrationRequest
	Success       string
	Error         string
	UserInfo      *auth.UserInfo
}

func (h *Handler) GetPendingRegistrationsHandler(w http.ResponseWriter, r *http.Request) {
	registrations, err := auth.GetPendingRegistrations(h.AuthStore)
	if err != nil {
		slog.Error("Failed to get pending registrations", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := PendingRegistrationsData{
		Registrations: registrations,
		UserInfo:      auth.GetUserInfoFromContext(r.Context()),
	}

	theme := utils.GetThemeFromRequest(r)
	templateName := utils.GetThemedTemplateName("pending-registrations.gohtml", theme)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = templates.Templates.ExecuteTemplate(w, templateName, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Handler) ApproveRegistrationHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "Missing registration ID", http.StatusBadRequest)
		return
	}

	registrationID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid registration ID", http.StatusBadRequest)
		return
	}

	user, err := auth.GetUserBySession(h.AuthStore, r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	regRequest, err := auth.GetRegistrationRequestByID(h.AuthStore, registrationID)
	if err != nil {
		slog.Error("Failed to get registration request", "error", err)
		http.Error(w, "Registration request not found", http.StatusNotFound)
		return
	}

	err = auth.ApproveRegistration(h.AuthStore, registrationID, user.ID)
	if err != nil {
		slog.Error("Failed to approve registration", "error", err)
		http.Error(w, "Failed to approve registration", http.StatusInternalServerError)
		return
	}

	err = mail.SendRegistrationApprovedNotification(regRequest.Email, regRequest.Username)
	if err != nil {
		slog.Error("Failed to send approval email", "error", err)
	}

	slog.Info("Registration approved", "admin_id", user.ID, "registration_id", registrationID, "username", regRequest.Username)

	theme := utils.GetThemeFromRequest(r)
	redirectURL := "/admin/registrations?success=Registration approved successfully"
	if theme != utils.ThemeDefault {
		redirectURL += "&theme=" + string(theme)
	}
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}

func (h *Handler) DenyRegistrationHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "Missing registration ID", http.StatusBadRequest)
		return
	}

	registrationID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid registration ID", http.StatusBadRequest)
		return
	}

	user, err := auth.GetUserBySession(h.AuthStore, r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	regRequest, err := auth.GetRegistrationRequestByID(h.AuthStore, registrationID)
	if err != nil {
		slog.Error("Failed to get registration request", "error", err)
		http.Error(w, "Registration request not found", http.StatusNotFound)
		return
	}

	reason := "Registration denied by administrator"
	err = auth.RejectRegistration(h.AuthStore, registrationID, user.ID, reason)
	if err != nil {
		slog.Error("Failed to deny registration", "error", err)
		http.Error(w, "Failed to deny registration", http.StatusInternalServerError)
		return
	}

	slog.Info("Registration denied", "admin_id", user.ID, "registration_id", registrationID, "username", regRequest.Username)

	theme := utils.GetThemeFromRequest(r)
	redirectURL := "/admin/registrations?success=Registration denied successfully"
	if theme != utils.ThemeDefault {
		redirectURL += "&theme=" + string(theme)
	}
	http.Redirect(w, r, redirectURL, http.StatusSeeOther)
}
