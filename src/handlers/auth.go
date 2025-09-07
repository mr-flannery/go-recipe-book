package handlers

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/mr-flannery/go-recipe-book/src/auth"
	"github.com/mr-flannery/go-recipe-book/src/config"
	"github.com/mr-flannery/go-recipe-book/src/db"
	"github.com/mr-flannery/go-recipe-book/src/mail"
	"github.com/mr-flannery/go-recipe-book/src/templates"
)

type LoginData struct {
	RedirectURL string
	Error       string
}

func GetLoginHandler(w http.ResponseWriter, r *http.Request) {

	redirectURL := r.URL.Query().Get("redirect")
	data := LoginData{
		RedirectURL: redirectURL,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := templates.Templates.ExecuteTemplate(w, "login.gohtml", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func PostLoginHandler(w http.ResponseWriter, r *http.Request) {

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
		err := templates.Templates.ExecuteTemplate(w, "login.gohtml", data)
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

type RegisterData struct {
	Username string
	Email    string
	Error    string
	Success  string
}

func GetRegisterHandler(w http.ResponseWriter, r *http.Request) {
	data := RegisterData{}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := templates.Templates.ExecuteTemplate(w, "register.gohtml", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func PostRegisterHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	username := r.FormValue("username")
	email := r.FormValue("email")
	password := r.FormValue("password")
	confirmPassword := r.FormValue("confirm_password")

	data := RegisterData{
		Username: username,
		Email:    email,
	}

	// Validate passwords match
	if password != confirmPassword {
		data.Error = "Passwords do not match"
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		templates.Templates.ExecuteTemplate(w, "register.gohtml", data)
		return
	}

	// Get database connection
	database, err := db.GetConnection()
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		data.Error = "Internal server error. Please try again later."
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		templates.Templates.ExecuteTemplate(w, "register.gohtml", data)
		return
	}
	defer database.Close()

	// Create registration request
	err = auth.CreateRegistrationRequest(database, username, email, password)
	if err != nil {
		slog.Error("Failed to create registration request", "error", err)
		data.Error = err.Error()
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		templates.Templates.ExecuteTemplate(w, "register.gohtml", data)
		return
	}

	// Send notification email to admin
	conf := config.GetConfig()
	approvalURL := fmt.Sprintf("http://localhost:8080/admin/registrations")
	err = mail.SendNewRegistrationNotification(conf.DB.Admin.Email, conf.DB.Admin.Username, username, email, approvalURL)
	if err != nil {
		slog.Error("Failed to send admin notification email", "error", err)
		// Don't fail the registration if email fails
	}

	data.Success = "Registration request submitted successfully! An administrator will review your request and you will receive an email when it's approved."
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templates.Templates.ExecuteTemplate(w, "register.gohtml", data)
}

type PendingRegistrationsData struct {
	Registrations []auth.RegistrationRequest
	Success       string
	Error         string
}

func GetPendingRegistrationsHandler(w http.ResponseWriter, r *http.Request) {
	// Get database connection
	database, err := db.GetConnection()
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer database.Close()

	// Get pending registrations
	registrations, err := auth.GetPendingRegistrations(database)
	if err != nil {
		slog.Error("Failed to get pending registrations", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	data := PendingRegistrationsData{
		Registrations: registrations,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err = templates.Templates.ExecuteTemplate(w, "pending-registrations.gohtml", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func ApproveRegistrationHandler(w http.ResponseWriter, r *http.Request) {
	// Get registration ID from URL path
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

	// Get database connection
	database, err := db.GetConnection()
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer database.Close()

	// Get current user (admin)
	user, err := auth.GetUserBySession(database, r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get registration request details before approving
	regRequest, err := auth.GetRegistrationRequestByID(database, registrationID)
	if err != nil {
		slog.Error("Failed to get registration request", "error", err)
		http.Error(w, "Registration request not found", http.StatusNotFound)
		return
	}

	// Approve registration
	err = auth.ApproveRegistration(database, registrationID, user.ID)
	if err != nil {
		slog.Error("Failed to approve registration", "error", err)
		http.Error(w, "Failed to approve registration", http.StatusInternalServerError)
		return
	}

	// Send approval email to user
	err = mail.SendRegistrationApprovedNotification(regRequest.Email, regRequest.Username)
	if err != nil {
		slog.Error("Failed to send approval email", "error", err)
		// Don't fail the approval if email fails
	}

	slog.Info("Registration approved", "admin_id", user.ID, "registration_id", registrationID, "username", regRequest.Username)

	// Redirect back to pending registrations with success message
	http.Redirect(w, r, "/admin/registrations?success=Registration approved successfully", http.StatusSeeOther)
}

func DenyRegistrationHandler(w http.ResponseWriter, r *http.Request) {
	// Get registration ID from URL path
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

	// Get database connection
	database, err := db.GetConnection()
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	defer database.Close()

	// Get current user (admin)
	user, err := auth.GetUserBySession(database, r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get registration request details before denying
	regRequest, err := auth.GetRegistrationRequestByID(database, registrationID)
	if err != nil {
		slog.Error("Failed to get registration request", "error", err)
		http.Error(w, "Registration request not found", http.StatusNotFound)
		return
	}

	// Deny registration
	reason := "Registration denied by administrator"
	err = auth.RejectRegistration(database, registrationID, user.ID, reason)
	if err != nil {
		slog.Error("Failed to deny registration", "error", err)
		http.Error(w, "Failed to deny registration", http.StatusInternalServerError)
		return
	}

	slog.Info("Registration denied", "admin_id", user.ID, "registration_id", registrationID, "username", regRequest.Username)

	// Redirect back to pending registrations with success message
	http.Redirect(w, r, "/admin/registrations?success=Registration denied successfully", http.StatusSeeOther)
}
