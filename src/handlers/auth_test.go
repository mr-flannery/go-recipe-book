package handlers

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/mr-flannery/go-recipe-book/src/auth"
	mailmocks "github.com/mr-flannery/go-recipe-book/src/mail/mocks"
	"github.com/mr-flannery/go-recipe-book/src/store"
	"github.com/mr-flannery/go-recipe-book/src/store/mocks"
	tmocks "github.com/mr-flannery/go-recipe-book/src/templates/mocks"
)

func TestGetLoginHandler_RendersLoginPage(t *testing.T) {
	var capturedTemplate string
	var capturedData any
	mockRenderer := &tmocks.MockRenderer{
		RenderPageFunc: func(w http.ResponseWriter, name string, data any) {
			capturedTemplate = name
			capturedData = data
			w.WriteHeader(http.StatusOK)
		},
	}

	h := &Handler{
		Renderer: mockRenderer,
	}

	req := httptest.NewRequest(http.MethodGet, "/login?redirect=/recipes/1", nil)
	userInfo := &auth.UserInfo{IsLoggedIn: false}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.GetLoginHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if capturedTemplate != "login.gohtml" {
		t.Errorf("expected template login.gohtml, got %s", capturedTemplate)
	}

	loginData, ok := capturedData.(LoginData)
	if !ok {
		t.Fatal("expected capturedData to be LoginData")
	}

	if loginData.RedirectURL != "/recipes/1" {
		t.Errorf("expected redirect URL /recipes/1, got %s", loginData.RedirectURL)
	}
}

func TestPostLoginHandler_RedirectsToHomeOnSuccessfulLogin(t *testing.T) {
	validPassword := "Correct#Pass1"
	hashedPassword, err := auth.HashPassword(validPassword)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	mockAuthStore := &mocks.MockAuthStore{
		GetUserByEmailFunc: func(email string) (*store.AuthUser, string, error) {
			return &store.AuthUser{
				ID:       1,
				Username: "testuser",
				Email:    email,
			}, hashedPassword, nil
		},
		CreateSessionFunc: func(session *store.Session) error {
			return nil
		},
		UpdateLastLoginFunc: func(userID int) error {
			return nil
		},
	}

	mockRenderer := &tmocks.MockRenderer{
		RenderPageFunc: func(w http.ResponseWriter, name string, data any) {
			w.WriteHeader(http.StatusOK)
		},
	}

	h := &Handler{
		AuthStore: mockAuthStore,
		Renderer:  mockRenderer,
	}

	form := url.Values{}
	form.Set("email", "test@example.com")
	form.Set("password", validPassword)

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	userInfo := &auth.UserInfo{IsLoggedIn: false}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.PostLoginHandler(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rec.Code)
	}

	location := rec.Header().Get("Location")
	if location != "/" {
		t.Errorf("expected redirect to /, got %s", location)
	}

	cookies := rec.Result().Cookies()
	foundSessionCookie := false
	for _, c := range cookies {
		if c.Name == "session" {
			foundSessionCookie = true
		}
	}
	if !foundSessionCookie {
		t.Error("expected session cookie to be set")
	}
}

func TestPostLoginHandler_RedirectsToCustomURLAfterLogin(t *testing.T) {
	validPassword := "Correct#Pass1"
	hashedPassword, err := auth.HashPassword(validPassword)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	mockAuthStore := &mocks.MockAuthStore{
		GetUserByEmailFunc: func(email string) (*store.AuthUser, string, error) {
			return &store.AuthUser{
				ID:       1,
				Username: "testuser",
				Email:    email,
			}, hashedPassword, nil
		},
		CreateSessionFunc: func(session *store.Session) error {
			return nil
		},
		UpdateLastLoginFunc: func(userID int) error {
			return nil
		},
	}

	mockRenderer := &tmocks.MockRenderer{
		RenderPageFunc: func(w http.ResponseWriter, name string, data any) {
			w.WriteHeader(http.StatusOK)
		},
	}

	h := &Handler{
		AuthStore: mockAuthStore,
		Renderer:  mockRenderer,
	}

	form := url.Values{}
	form.Set("email", "test@example.com")
	form.Set("password", validPassword)
	form.Set("redirect", "/recipes/42")

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	userInfo := &auth.UserInfo{IsLoggedIn: false}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.PostLoginHandler(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rec.Code)
	}

	location := rec.Header().Get("Location")
	if location != "/recipes/42" {
		t.Errorf("expected redirect to /recipes/42, got %s", location)
	}
}

func TestPostLoginHandler_ShowsErrorOnInvalidCredentials(t *testing.T) {
	mockAuthStore := &mocks.MockAuthStore{
		GetUserByEmailFunc: func(email string) (*store.AuthUser, string, error) {
			return nil, "", errors.New("user not found")
		},
	}

	var capturedData any
	mockRenderer := &tmocks.MockRenderer{
		RenderPageFunc: func(w http.ResponseWriter, name string, data any) {
			capturedData = data
			w.WriteHeader(http.StatusOK)
		},
	}

	h := &Handler{
		AuthStore: mockAuthStore,
		Renderer:  mockRenderer,
	}

	form := url.Values{}
	form.Set("email", "wrong@example.com")
	form.Set("password", "wrongpassword")

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	userInfo := &auth.UserInfo{IsLoggedIn: false}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.PostLoginHandler(rec, req)

	loginData, ok := capturedData.(LoginData)
	if !ok {
		t.Fatal("expected capturedData to be LoginData")
	}

	if loginData.Error == "" {
		t.Error("expected error message to be set")
	}
}

func TestLogoutHandler_ClearsSessionAndRedirects(t *testing.T) {
	invalidateCalled := false
	mockAuthStore := &mocks.MockAuthStore{
		GetSessionFunc: func(sessionID string) (*store.Session, error) {
			return &store.Session{ID: sessionID, UserID: 1}, nil
		},
		DeleteSessionFunc: func(sessionID string) error {
			invalidateCalled = true
			return nil
		},
	}

	h := &Handler{
		AuthStore: mockAuthStore,
	}

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
	rec := httptest.NewRecorder()

	h.LogoutHandler(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rec.Code)
	}

	location := rec.Header().Get("Location")
	if location != "/" {
		t.Errorf("expected redirect to /, got %s", location)
	}

	if !invalidateCalled {
		t.Error("expected session to be invalidated")
	}

	cookies := rec.Result().Cookies()
	for _, c := range cookies {
		if c.Name == "session" && c.MaxAge < 0 {
			return
		}
	}
	t.Error("expected session cookie to be cleared")
}

func TestGetRegisterHandler_RendersRegisterPage(t *testing.T) {
	var capturedTemplate string
	mockRenderer := &tmocks.MockRenderer{
		RenderPageFunc: func(w http.ResponseWriter, name string, data any) {
			capturedTemplate = name
			w.WriteHeader(http.StatusOK)
		},
	}

	h := &Handler{
		Renderer: mockRenderer,
	}

	req := httptest.NewRequest(http.MethodGet, "/register", nil)
	userInfo := &auth.UserInfo{IsLoggedIn: false}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.GetRegisterHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if capturedTemplate != "register.gohtml" {
		t.Errorf("expected template register.gohtml, got %s", capturedTemplate)
	}
}

func TestPostRegisterHandler_ShowsErrorWhenPasswordsDontMatch(t *testing.T) {
	var capturedData any
	mockRenderer := &tmocks.MockRenderer{
		RenderPageFunc: func(w http.ResponseWriter, name string, data any) {
			capturedData = data
			w.WriteHeader(http.StatusOK)
		},
	}

	h := &Handler{
		Renderer: mockRenderer,
	}

	form := url.Values{}
	form.Set("username", "testuser")
	form.Set("email", "test@example.com")
	form.Set("password", "password123")
	form.Set("confirm_password", "differentpassword")

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	userInfo := &auth.UserInfo{IsLoggedIn: false}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.PostRegisterHandler(rec, req)

	registerData, ok := capturedData.(RegisterData)
	if !ok {
		t.Fatal("expected capturedData to be RegisterData")
	}

	if registerData.Error != "Passwords do not match" {
		t.Errorf("expected error 'Passwords do not match', got '%s'", registerData.Error)
	}
}

func TestGetPendingRegistrationsHandler_ListsRegistrations(t *testing.T) {
	mockAuthStore := &mocks.MockAuthStore{
		GetPendingRegistrationsFunc: func() ([]store.RegistrationRequest, error) {
			return []store.RegistrationRequest{
				{ID: 1, Username: "user1", Email: "user1@example.com", Status: "pending"},
				{ID: 2, Username: "user2", Email: "user2@example.com", Status: "pending"},
			}, nil
		},
	}

	var capturedData any
	mockRenderer := &tmocks.MockRenderer{
		RenderPageFunc: func(w http.ResponseWriter, name string, data any) {
			capturedData = data
			w.WriteHeader(http.StatusOK)
		},
	}

	h := &Handler{
		AuthStore: mockAuthStore,
		Renderer:  mockRenderer,
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/registrations", nil)
	userInfo := &auth.UserInfo{IsLoggedIn: true, IsAdmin: true, UserID: 1, Username: "admin"}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.GetPendingRegistrationsHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	regData, ok := capturedData.(PendingRegistrationsData)
	if !ok {
		t.Fatal("expected capturedData to be PendingRegistrationsData")
	}

	if len(regData.Registrations) != 2 {
		t.Errorf("expected 2 registrations, got %d", len(regData.Registrations))
	}
}

func TestGetPendingRegistrationsHandler_ReturnsErrorWhenStoreFails(t *testing.T) {
	mockAuthStore := &mocks.MockAuthStore{
		GetPendingRegistrationsFunc: func() ([]store.RegistrationRequest, error) {
			return nil, errors.New("database error")
		},
	}

	mockRenderer := &tmocks.MockRenderer{
		RenderErrorFunc: func(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
			w.WriteHeader(statusCode)
		},
	}

	h := &Handler{
		AuthStore: mockAuthStore,
		Renderer:  mockRenderer,
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/registrations", nil)
	rec := httptest.NewRecorder()

	h.GetPendingRegistrationsHandler(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestApproveRegistrationHandler_ReturnsBadRequestWhenIDMissing(t *testing.T) {
	mockRenderer := &tmocks.MockRenderer{
		RenderErrorFunc: func(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
			w.WriteHeader(statusCode)
		},
	}

	h := &Handler{
		Renderer: mockRenderer,
	}

	req := httptest.NewRequest(http.MethodPost, "/admin/registrations//approve", nil)
	req.SetPathValue("id", "")
	rec := httptest.NewRecorder()

	h.ApproveRegistrationHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestApproveRegistrationHandler_ReturnsBadRequestWhenIDInvalid(t *testing.T) {
	mockRenderer := &tmocks.MockRenderer{
		RenderErrorFunc: func(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
			w.WriteHeader(statusCode)
		},
	}

	h := &Handler{
		Renderer: mockRenderer,
	}

	req := httptest.NewRequest(http.MethodPost, "/admin/registrations/abc/approve", nil)
	req.SetPathValue("id", "abc")
	rec := httptest.NewRecorder()

	h.ApproveRegistrationHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestApproveRegistrationHandler_ReturnsUnauthorizedWhenNotLoggedIn(t *testing.T) {
	mockAuthStore := &mocks.MockAuthStore{
		GetSessionFunc: func(sessionID string) (*store.Session, error) {
			return nil, errors.New("no session")
		},
	}

	mockRenderer := &tmocks.MockRenderer{
		RenderErrorFunc: func(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
			w.WriteHeader(statusCode)
		},
	}

	h := &Handler{
		AuthStore: mockAuthStore,
		Renderer:  mockRenderer,
	}

	req := httptest.NewRequest(http.MethodPost, "/admin/registrations/1/approve", nil)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	h.ApproveRegistrationHandler(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestDenyRegistrationHandler_ReturnsBadRequestWhenIDMissing(t *testing.T) {
	mockRenderer := &tmocks.MockRenderer{
		RenderErrorFunc: func(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
			w.WriteHeader(statusCode)
		},
	}

	h := &Handler{
		Renderer: mockRenderer,
	}

	req := httptest.NewRequest(http.MethodPost, "/admin/registrations//deny", nil)
	req.SetPathValue("id", "")
	rec := httptest.NewRecorder()

	h.DenyRegistrationHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestDenyRegistrationHandler_ReturnsUnauthorizedWhenNotLoggedIn(t *testing.T) {
	mockAuthStore := &mocks.MockAuthStore{
		GetSessionFunc: func(sessionID string) (*store.Session, error) {
			return nil, errors.New("no session")
		},
	}

	mockRenderer := &tmocks.MockRenderer{
		RenderErrorFunc: func(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
			w.WriteHeader(statusCode)
		},
	}

	h := &Handler{
		AuthStore: mockAuthStore,
		Renderer:  mockRenderer,
	}

	req := httptest.NewRequest(http.MethodPost, "/admin/registrations/1/deny", nil)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	h.DenyRegistrationHandler(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestGetUsersHandler_ListsUsers(t *testing.T) {
	mockAuthStore := &mocks.MockAuthStore{
		GetAllUsersFunc: func() ([]store.AuthUser, error) {
			return []store.AuthUser{
				{ID: 1, Username: "admin", Email: "admin@example.com", IsAdmin: true},
				{ID: 2, Username: "user", Email: "user@example.com", IsAdmin: false},
			}, nil
		},
	}

	var capturedData any
	mockRenderer := &tmocks.MockRenderer{
		RenderPageFunc: func(w http.ResponseWriter, name string, data any) {
			capturedData = data
			w.WriteHeader(http.StatusOK)
		},
	}

	h := &Handler{
		AuthStore: mockAuthStore,
		Renderer:  mockRenderer,
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
	userInfo := &auth.UserInfo{IsLoggedIn: true, IsAdmin: true, UserID: 1, Username: "admin"}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.GetUsersHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	usersData, ok := capturedData.(UsersData)
	if !ok {
		t.Fatal("expected capturedData to be UsersData")
	}

	if len(usersData.Users) != 2 {
		t.Errorf("expected 2 users, got %d", len(usersData.Users))
	}
}

func TestGetUsersHandler_ReturnsErrorWhenStoreFails(t *testing.T) {
	mockAuthStore := &mocks.MockAuthStore{
		GetAllUsersFunc: func() ([]store.AuthUser, error) {
			return nil, errors.New("database error")
		},
	}

	mockRenderer := &tmocks.MockRenderer{
		RenderErrorFunc: func(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
			w.WriteHeader(statusCode)
		},
	}

	h := &Handler{
		AuthStore: mockAuthStore,
		Renderer:  mockRenderer,
	}

	req := httptest.NewRequest(http.MethodGet, "/admin/users", nil)
	rec := httptest.NewRecorder()

	h.GetUsersHandler(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestDeleteUserHandler_ReturnsBadRequestWhenIDMissing(t *testing.T) {
	h := &Handler{}

	req := httptest.NewRequest(http.MethodDelete, "/admin/users/", nil)
	req.SetPathValue("id", "")
	rec := httptest.NewRecorder()

	h.DeleteUserHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestDeleteUserHandler_ReturnsBadRequestWhenIDInvalid(t *testing.T) {
	h := &Handler{}

	req := httptest.NewRequest(http.MethodDelete, "/admin/users/abc", nil)
	req.SetPathValue("id", "abc")
	rec := httptest.NewRecorder()

	h.DeleteUserHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestDeleteUserHandler_ReturnsForbiddenWhenDeletingOwnAccount(t *testing.T) {
	h := &Handler{}

	req := httptest.NewRequest(http.MethodDelete, "/admin/users/1", nil)
	req.SetPathValue("id", "1")
	userInfo := &auth.UserInfo{IsLoggedIn: true, IsAdmin: true, UserID: 1, Username: "admin"}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.DeleteUserHandler(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestDeleteUserHandler_ReturnsNotFoundWhenUserDoesNotExist(t *testing.T) {
	mockAuthStore := &mocks.MockAuthStore{
		GetUserByIDFunc: func(userID int) (*store.AuthUser, error) {
			return nil, errors.New("user not found")
		},
	}

	h := &Handler{
		AuthStore: mockAuthStore,
	}

	req := httptest.NewRequest(http.MethodDelete, "/admin/users/999", nil)
	req.SetPathValue("id", "999")
	userInfo := &auth.UserInfo{IsLoggedIn: true, IsAdmin: true, UserID: 1, Username: "admin"}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.DeleteUserHandler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestDeleteUserHandler_ReturnsForbiddenWhenDeletingAdmin(t *testing.T) {
	mockAuthStore := &mocks.MockAuthStore{
		GetUserByIDFunc: func(userID int) (*store.AuthUser, error) {
			return &store.AuthUser{ID: 2, Username: "otheradmin", IsAdmin: true}, nil
		},
	}

	h := &Handler{
		AuthStore: mockAuthStore,
	}

	req := httptest.NewRequest(http.MethodDelete, "/admin/users/2", nil)
	req.SetPathValue("id", "2")
	userInfo := &auth.UserInfo{IsLoggedIn: true, IsAdmin: true, UserID: 1, Username: "admin"}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.DeleteUserHandler(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestDeleteUserHandler_DeletesUserSuccessfully(t *testing.T) {
	deleteCalled := false
	mockAuthStore := &mocks.MockAuthStore{
		GetUserByIDFunc: func(userID int) (*store.AuthUser, error) {
			return &store.AuthUser{ID: 2, Username: "regularuser", IsAdmin: false}, nil
		},
		DeleteUserFunc: func(userID int) error {
			deleteCalled = true
			return nil
		},
	}

	h := &Handler{
		AuthStore: mockAuthStore,
	}

	req := httptest.NewRequest(http.MethodDelete, "/admin/users/2", nil)
	req.SetPathValue("id", "2")
	userInfo := &auth.UserInfo{IsLoggedIn: true, IsAdmin: true, UserID: 1, Username: "admin"}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.DeleteUserHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if !deleteCalled {
		t.Error("expected delete to be called")
	}
}

func TestPostRegisterHandler_SendsNotificationEmailOnSuccess(t *testing.T) {
	emailSent := false
	var capturedRecipient string
	mockMailClient := &mailmocks.MockMailClient{
		SendEmailFunc: func(recipientEmail, recipientName, subject, plainContent string) error {
			emailSent = true
			capturedRecipient = recipientEmail
			return nil
		},
	}

	mockAuthStore := &mocks.MockAuthStore{
		CreateRegistrationRequestFunc: func(username, email, passwordHash string) error {
			return nil
		},
	}

	var capturedData any
	mockRenderer := &tmocks.MockRenderer{
		RenderPageFunc: func(w http.ResponseWriter, name string, data any) {
			capturedData = data
			w.WriteHeader(http.StatusOK)
		},
	}

	h := &Handler{
		AuthStore:  mockAuthStore,
		Renderer:   mockRenderer,
		MailClient: mockMailClient,
	}

	form := url.Values{}
	form.Set("username", "newuser")
	form.Set("email", "new@example.com")
	form.Set("password", "StrongPass123!")
	form.Set("confirm_password", "StrongPass123!")

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	userInfo := &auth.UserInfo{IsLoggedIn: false}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.PostRegisterHandler(rec, req)

	if !emailSent {
		t.Error("expected notification email to be sent")
	}

	if capturedRecipient == "" {
		t.Error("expected recipient to be captured")
	}

	registerData, ok := capturedData.(RegisterData)
	if !ok {
		t.Fatal("expected capturedData to be RegisterData")
	}

	if registerData.Success == "" {
		t.Error("expected success message to be set")
	}
}

func TestPostRegisterHandler_SucceedsEvenWhenEmailFails(t *testing.T) {
	mockMailClient := &mailmocks.MockMailClient{
		SendEmailFunc: func(recipientEmail, recipientName, subject, plainContent string) error {
			return errors.New("email service unavailable")
		},
	}

	mockAuthStore := &mocks.MockAuthStore{
		CreateRegistrationRequestFunc: func(username, email, passwordHash string) error {
			return nil
		},
	}

	var capturedData any
	mockRenderer := &tmocks.MockRenderer{
		RenderPageFunc: func(w http.ResponseWriter, name string, data any) {
			capturedData = data
			w.WriteHeader(http.StatusOK)
		},
	}

	h := &Handler{
		AuthStore:  mockAuthStore,
		Renderer:   mockRenderer,
		MailClient: mockMailClient,
	}

	form := url.Values{}
	form.Set("username", "newuser")
	form.Set("email", "new@example.com")
	form.Set("password", "StrongPass123!")
	form.Set("confirm_password", "StrongPass123!")

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	userInfo := &auth.UserInfo{IsLoggedIn: false}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.PostRegisterHandler(rec, req)

	registerData, ok := capturedData.(RegisterData)
	if !ok {
		t.Fatal("expected capturedData to be RegisterData")
	}

	if registerData.Success == "" {
		t.Error("expected success message to be set even when email fails")
	}
}

func TestApproveRegistrationHandler_SendsApprovalEmailOnSuccess(t *testing.T) {
	emailSent := false
	var capturedRecipient string
	mockMailClient := &mailmocks.MockMailClient{
		SendEmailFunc: func(recipientEmail, recipientName, subject, plainContent string) error {
			emailSent = true
			capturedRecipient = recipientEmail
			return nil
		},
	}

	mockAuthStore := &mocks.MockAuthStore{
		GetSessionFunc: func(sessionID string) (*store.Session, error) {
			return &store.Session{ID: sessionID, UserID: 1}, nil
		},
		GetUserByIDFunc: func(userID int) (*store.AuthUser, error) {
			return &store.AuthUser{ID: 1, Username: "admin", IsAdmin: true}, nil
		},
		GetPendingRegistrationsFunc: func() ([]store.RegistrationRequest, error) {
			return []store.RegistrationRequest{
				{ID: 1, Username: "newuser", Email: "new@example.com", Status: "pending"},
			}, nil
		},
		ApproveRegistrationFunc: func(requestID, adminID int) error {
			return nil
		},
	}

	mockRenderer := &tmocks.MockRenderer{}

	h := &Handler{
		AuthStore:  mockAuthStore,
		Renderer:   mockRenderer,
		MailClient: mockMailClient,
	}

	req := httptest.NewRequest(http.MethodPost, "/admin/registrations/1/approve", nil)
	req.SetPathValue("id", "1")
	req.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
	userInfo := &auth.UserInfo{IsLoggedIn: true, IsAdmin: true, UserID: 1, Username: "admin"}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.ApproveRegistrationHandler(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rec.Code)
	}

	if !emailSent {
		t.Error("expected approval email to be sent")
	}

	if capturedRecipient != "new@example.com" {
		t.Errorf("expected recipient 'new@example.com', got '%s'", capturedRecipient)
	}
}

func TestApproveRegistrationHandler_SucceedsEvenWhenEmailFails(t *testing.T) {
	mockMailClient := &mailmocks.MockMailClient{
		SendEmailFunc: func(recipientEmail, recipientName, subject, plainContent string) error {
			return errors.New("email service unavailable")
		},
	}

	mockAuthStore := &mocks.MockAuthStore{
		GetSessionFunc: func(sessionID string) (*store.Session, error) {
			return &store.Session{ID: sessionID, UserID: 1}, nil
		},
		GetUserByIDFunc: func(userID int) (*store.AuthUser, error) {
			return &store.AuthUser{ID: 1, Username: "admin", IsAdmin: true}, nil
		},
		GetPendingRegistrationsFunc: func() ([]store.RegistrationRequest, error) {
			return []store.RegistrationRequest{
				{ID: 1, Username: "newuser", Email: "new@example.com", Status: "pending"},
			}, nil
		},
		ApproveRegistrationFunc: func(requestID, adminID int) error {
			return nil
		},
	}

	mockRenderer := &tmocks.MockRenderer{}

	h := &Handler{
		AuthStore:  mockAuthStore,
		Renderer:   mockRenderer,
		MailClient: mockMailClient,
	}

	req := httptest.NewRequest(http.MethodPost, "/admin/registrations/1/approve", nil)
	req.SetPathValue("id", "1")
	req.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
	userInfo := &auth.UserInfo{IsLoggedIn: true, IsAdmin: true, UserID: 1, Username: "admin"}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.ApproveRegistrationHandler(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d - registration should succeed even if email fails", http.StatusSeeOther, rec.Code)
	}
}
