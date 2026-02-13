package handlers

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/mr-flannery/go-recipe-book/src/auth"
	"github.com/mr-flannery/go-recipe-book/src/config"
	"github.com/mr-flannery/go-recipe-book/src/mail"
)

type LoginData struct {
	RedirectURL string
	Error       string
	UserInfo    *auth.UserInfo
}

func (h *Handler) GetLoginHandler(w http.ResponseWriter, r *http.Request) {
	redirectURL := r.URL.Query().Get("redirect")
	data := LoginData{
		RedirectURL: redirectURL,
		UserInfo:    auth.GetUserInfoFromContext(r.Context()),
	}
	h.Renderer.RenderPage(w, "login.gohtml", data)
}

func (h *Handler) PostLoginHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	email := r.FormValue("email")
	password := r.FormValue("password")
	redirectURL := r.FormValue("redirect")

	user, err := auth.Authenticate(h.AuthStore, email, password)
	if err != nil {
		data := LoginData{
			RedirectURL: redirectURL,
			Error:       "Invalid username or password. Please try again.",
			UserInfo:    auth.GetUserInfoFromContext(r.Context()),
		}
		h.Renderer.RenderPage(w, "login.gohtml", data)
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
	}

	http.Redirect(w, r, finalRedirectURL, http.StatusSeeOther)
}

func (h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	sessionID, err := auth.GetSessionFromRequest(r)
	if err == nil {
		auth.InvalidateSession(h.AuthStore, sessionID)
	}

	auth.ClearSessionCookie(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type RegisterData struct {
	Username string
	Email    string
	Error    string
	Success  string
	UserInfo *auth.UserInfo
}

func (h *Handler) GetRegisterHandler(w http.ResponseWriter, r *http.Request) {
	data := RegisterData{
		UserInfo: auth.GetUserInfoFromContext(r.Context()),
	}
	h.Renderer.RenderPage(w, "register.gohtml", data)
}

func (h *Handler) PostRegisterHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	data := RegisterData{
		Username: username,
		Email:    email,
		UserInfo: auth.GetUserInfoFromContext(r.Context()),
	}

	if password != confirmPassword {
		data.Error = "Passwords do not match"
		h.Renderer.RenderPage(w, "register.gohtml", data)
		return
	}

	err := auth.CreateRegistrationRequest(h.AuthStore, username, email, password)
	if err != nil {
		slog.Error("Failed to create registration request", "error", err)
		data.Error = err.Error()
		h.Renderer.RenderPage(w, "register.gohtml", data)
		return
	}

	conf := config.GetConfig()
	approvalURL := "http://localhost:8080/admin/registrations"
	err = mail.SendNewRegistrationNotification(conf.DB.Admin.Email, conf.DB.Admin.Username, username, email, approvalURL)
	if err != nil {
		slog.Error("Failed to send admin notification email", "error", err)
	}

	data.Success = "Registration request submitted successfully! An administrator will review your request and you will receive an email when it's approved."
	h.Renderer.RenderPage(w, "register.gohtml", data)
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
	h.Renderer.RenderPage(w, "pending-registrations.gohtml", data)
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

	http.Redirect(w, r, "/admin/registrations?success=Registration approved successfully", http.StatusSeeOther)
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

	err = auth.RejectRegistration(h.AuthStore, registrationID, user.ID)
	if err != nil {
		slog.Error("Failed to deny registration", "error", err)
		http.Error(w, "Failed to deny registration", http.StatusInternalServerError)
		return
	}

	slog.Info("Registration denied", "admin_id", user.ID, "registration_id", registrationID, "username", regRequest.Username)

	http.Redirect(w, r, "/admin/registrations?success=Registration denied successfully", http.StatusSeeOther)
}

type UsersData struct {
	Users    []auth.User
	Success  string
	Error    string
	UserInfo *auth.UserInfo
}

func (h *Handler) GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := auth.GetAllUsers(h.AuthStore)
	if err != nil {
		slog.Error("Failed to get users", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := UsersData{
		Users:    users,
		Success:  r.URL.Query().Get("success"),
		UserInfo: auth.GetUserInfoFromContext(r.Context()),
	}
	h.Renderer.RenderPage(w, "users.gohtml", data)
}

func (h *Handler) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		http.Error(w, "Missing user ID", http.StatusBadRequest)
		return
	}

	userID, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	currentUser := auth.GetUserInfoFromContext(r.Context())
	if currentUser.UserID == userID {
		http.Error(w, "Cannot delete your own account", http.StatusForbidden)
		return
	}

	targetUser, err := h.AuthStore.GetUserByID(userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if targetUser.IsAdmin {
		http.Error(w, "Cannot delete admin users", http.StatusForbidden)
		return
	}

	err = auth.DeleteUser(h.AuthStore, userID)
	if err != nil {
		slog.Error("Failed to delete user", "error", err, "user_id", userID)
		http.Error(w, "Failed to delete user", http.StatusInternalServerError)
		return
	}

	slog.Info("User deleted", "admin_id", currentUser.UserID, "deleted_user_id", userID, "deleted_username", targetUser.Username)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("User deleted successfully"))
}
