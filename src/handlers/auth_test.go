package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

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
		GetUserByEmailFunc: func(ctx context.Context, email string) (*store.AuthUser, string, error) {
			return &store.AuthUser{
				ID:       1,
				Username: "testuser",
				Email:    email,
			}, hashedPassword, nil
		},
		CreateSessionFunc: func(ctx context.Context, session *store.Session) error {
			return nil
		},
		UpdateLastLoginFunc: func(ctx context.Context, userID int) error {
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
		GetUserByEmailFunc: func(ctx context.Context, email string) (*store.AuthUser, string, error) {
			return &store.AuthUser{
				ID:       1,
				Username: "testuser",
				Email:    email,
			}, hashedPassword, nil
		},
		CreateSessionFunc: func(ctx context.Context, session *store.Session) error {
			return nil
		},
		UpdateLastLoginFunc: func(ctx context.Context, userID int) error {
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
		GetUserByEmailFunc: func(ctx context.Context, email string) (*store.AuthUser, string, error) {
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
		GetSessionFunc: func(ctx context.Context, sessionID string) (*store.Session, error) {
			return &store.Session{ID: sessionID, UserID: 1}, nil
		},
		DeleteSessionFunc: func(ctx context.Context, sessionID string) error {
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

func TestGetRegistrationsHandler_ListsRegistrations(t *testing.T) {
	mockAuthStore := &mocks.MockAuthStore{
		CountAllRegistrationsFunc: func(ctx context.Context) (int, error) {
			return 2, nil
		},
		GetAllRegistrationsPaginatedFunc: func(ctx context.Context, limit, offset int) ([]store.RegistrationRequest, error) {
			return []store.RegistrationRequest{
				{ID: 1, Username: "user1", Email: "user1@example.com", Status: "pending"},
				{ID: 2, Username: "user2", Email: "user2@example.com", Status: "approved"},
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

	h.GetRegistrationsHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	regData, ok := capturedData.(RegistrationsData)
	if !ok {
		t.Fatal("expected capturedData to be RegistrationsData")
	}

	if len(regData.Registrations) != 2 {
		t.Errorf("expected 2 registrations, got %d", len(regData.Registrations))
	}
}

func TestGetRegistrationsHandler_ReturnsErrorWhenStoreFails(t *testing.T) {
	mockAuthStore := &mocks.MockAuthStore{
		CountAllRegistrationsFunc: func(ctx context.Context) (int, error) {
			return 0, errors.New("database error")
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

	h.GetRegistrationsHandler(rec, req)

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
		GetSessionFunc: func(ctx context.Context, sessionID string) (*store.Session, error) {
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
		GetSessionFunc: func(ctx context.Context, sessionID string) (*store.Session, error) {
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
		GetAllUsersFunc: func(ctx context.Context) ([]store.AuthUser, error) {
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
		GetAllUsersFunc: func(ctx context.Context) ([]store.AuthUser, error) {
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
		GetUserByIDFunc: func(ctx context.Context, userID int) (*store.AuthUser, error) {
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
		GetUserByIDFunc: func(ctx context.Context, userID int) (*store.AuthUser, error) {
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
		GetUserByIDFunc: func(ctx context.Context, userID int) (*store.AuthUser, error) {
			return &store.AuthUser{ID: 2, Username: "regularuser", IsAdmin: false}, nil
		},
		DeleteUserFunc: func(ctx context.Context, userID int) error {
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
		SendEmailFunc: func(ctx context.Context, recipientEmail, recipientName, subject, plainContent string) error {
			emailSent = true
			capturedRecipient = recipientEmail
			return nil
		},
	}

	mockAuthStore := &mocks.MockAuthStore{
		CreateRegistrationRequestFunc: func(ctx context.Context, username, email, passwordHash string) error {
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
		SendEmailFunc: func(ctx context.Context, recipientEmail, recipientName, subject, plainContent string) error {
			return errors.New("email service unavailable")
		},
	}

	mockAuthStore := &mocks.MockAuthStore{
		CreateRegistrationRequestFunc: func(ctx context.Context, username, email, passwordHash string) error {
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
		SendEmailFunc: func(ctx context.Context, recipientEmail, recipientName, subject, plainContent string) error {
			emailSent = true
			capturedRecipient = recipientEmail
			return nil
		},
	}

	mockAuthStore := &mocks.MockAuthStore{
		GetSessionFunc: func(ctx context.Context, sessionID string) (*store.Session, error) {
			return &store.Session{ID: sessionID, UserID: 1}, nil
		},
		GetUserByIDFunc: func(ctx context.Context, userID int) (*store.AuthUser, error) {
			return &store.AuthUser{ID: 1, Username: "admin", IsAdmin: true}, nil
		},
		GetPendingRegistrationsFunc: func(ctx context.Context) ([]store.RegistrationRequest, error) {
			return []store.RegistrationRequest{
				{ID: 1, Username: "newuser", Email: "new@example.com", Status: "pending"},
			}, nil
		},
		ApproveRegistrationFunc: func(ctx context.Context, requestID, adminID int) error {
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
		SendEmailFunc: func(ctx context.Context, recipientEmail, recipientName, subject, plainContent string) error {
			return errors.New("email service unavailable")
		},
	}

	mockAuthStore := &mocks.MockAuthStore{
		GetSessionFunc: func(ctx context.Context, sessionID string) (*store.Session, error) {
			return &store.Session{ID: sessionID, UserID: 1}, nil
		},
		GetUserByIDFunc: func(ctx context.Context, userID int) (*store.AuthUser, error) {
			return &store.AuthUser{ID: 1, Username: "admin", IsAdmin: true}, nil
		},
		GetPendingRegistrationsFunc: func(ctx context.Context) ([]store.RegistrationRequest, error) {
			return []store.RegistrationRequest{
				{ID: 1, Username: "newuser", Email: "new@example.com", Status: "pending"},
			}, nil
		},
		ApproveRegistrationFunc: func(ctx context.Context, requestID, adminID int) error {
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

func TestGetForgotPasswordHandler_RendersForgotPasswordPage(t *testing.T) {
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

	req := httptest.NewRequest(http.MethodGet, "/forgot-password", nil)
	userInfo := &auth.UserInfo{IsLoggedIn: false}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.GetForgotPasswordHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if capturedTemplate != "forgot-password.gohtml" {
		t.Errorf("expected template forgot-password.gohtml, got %s", capturedTemplate)
	}
}

func TestPostForgotPasswordHandler_ShowsSuccessWhenUserNotFound(t *testing.T) {
	mockAuthStore := &mocks.MockAuthStore{
		GetUserByEmailFunc: func(ctx context.Context, email string) (*store.AuthUser, string, error) {
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
	form.Set("email", "nonexistent@example.com")

	req := httptest.NewRequest(http.MethodPost, "/forgot-password", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	userInfo := &auth.UserInfo{IsLoggedIn: false}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.PostForgotPasswordHandler(rec, req)

	forgotData, ok := capturedData.(ForgotPasswordData)
	if !ok {
		t.Fatal("expected capturedData to be ForgotPasswordData")
	}

	if forgotData.Success == "" {
		t.Error("expected success message to be set even when user not found")
	}

	if forgotData.Error != "" {
		t.Error("expected no error message when user not found")
	}
}

func TestPostForgotPasswordHandler_SendsEmailWhenUserExists(t *testing.T) {
	emailSent := false
	var capturedRecipient string

	mockMailClient := &mailmocks.MockMailClient{
		SendEmailFunc: func(ctx context.Context, recipientEmail, recipientName, subject, plainContent string) error {
			emailSent = true
			capturedRecipient = recipientEmail
			return nil
		},
	}

	mockAuthStore := &mocks.MockAuthStore{
		GetUserByEmailFunc: func(ctx context.Context, email string) (*store.AuthUser, string, error) {
			return &store.AuthUser{
				ID:       1,
				Username: "testuser",
				Email:    email,
			}, "hashedpassword", nil
		},
		CreatePasswordResetTokenFunc: func(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) error {
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
	form.Set("email", "test@example.com")

	req := httptest.NewRequest(http.MethodPost, "/forgot-password", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	userInfo := &auth.UserInfo{IsLoggedIn: false}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.PostForgotPasswordHandler(rec, req)

	if !emailSent {
		t.Error("expected password reset email to be sent")
	}

	if capturedRecipient != "test@example.com" {
		t.Errorf("expected recipient 'test@example.com', got '%s'", capturedRecipient)
	}

	forgotData, ok := capturedData.(ForgotPasswordData)
	if !ok {
		t.Fatal("expected capturedData to be ForgotPasswordData")
	}

	if forgotData.Success == "" {
		t.Error("expected success message to be set")
	}
}

func TestPostForgotPasswordHandler_SucceedsEvenWhenEmailFails(t *testing.T) {
	mockMailClient := &mailmocks.MockMailClient{
		SendEmailFunc: func(ctx context.Context, recipientEmail, recipientName, subject, plainContent string) error {
			return errors.New("email service unavailable")
		},
	}

	mockAuthStore := &mocks.MockAuthStore{
		GetUserByEmailFunc: func(ctx context.Context, email string) (*store.AuthUser, string, error) {
			return &store.AuthUser{
				ID:       1,
				Username: "testuser",
				Email:    email,
			}, "hashedpassword", nil
		},
		CreatePasswordResetTokenFunc: func(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) error {
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
	form.Set("email", "test@example.com")

	req := httptest.NewRequest(http.MethodPost, "/forgot-password", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	userInfo := &auth.UserInfo{IsLoggedIn: false}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.PostForgotPasswordHandler(rec, req)

	forgotData, ok := capturedData.(ForgotPasswordData)
	if !ok {
		t.Fatal("expected capturedData to be ForgotPasswordData")
	}

	if forgotData.Success == "" {
		t.Error("expected success message even when email fails")
	}
}

func TestGetResetPasswordHandler_RendersResetPasswordPage(t *testing.T) {
	mockAuthStore := &mocks.MockAuthStore{
		GetPasswordResetTokenFunc: func(ctx context.Context, tokenHash string) (*store.PasswordResetToken, error) {
			return &store.PasswordResetToken{
				ID:        1,
				UserID:    1,
				ExpiresAt: time.Now().Add(time.Hour),
				UsedAt:    nil,
			}, nil
		},
	}

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
		AuthStore: mockAuthStore,
		Renderer:  mockRenderer,
	}

	req := httptest.NewRequest(http.MethodGet, "/reset-password?token=validtoken123", nil)
	userInfo := &auth.UserInfo{IsLoggedIn: false}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.GetResetPasswordHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if capturedTemplate != "reset-password.gohtml" {
		t.Errorf("expected template reset-password.gohtml, got %s", capturedTemplate)
	}

	resetData, ok := capturedData.(ResetPasswordData)
	if !ok {
		t.Fatal("expected capturedData to be ResetPasswordData")
	}

	if resetData.InvalidToken {
		t.Error("expected token to be valid")
	}

	if resetData.Token != "validtoken123" {
		t.Errorf("expected token 'validtoken123', got '%s'", resetData.Token)
	}
}

func TestGetResetPasswordHandler_ShowsInvalidTokenWhenMissing(t *testing.T) {
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

	req := httptest.NewRequest(http.MethodGet, "/reset-password", nil)
	userInfo := &auth.UserInfo{IsLoggedIn: false}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.GetResetPasswordHandler(rec, req)

	resetData, ok := capturedData.(ResetPasswordData)
	if !ok {
		t.Fatal("expected capturedData to be ResetPasswordData")
	}

	if !resetData.InvalidToken {
		t.Error("expected InvalidToken to be true when token is missing")
	}
}

func TestGetResetPasswordHandler_ShowsInvalidTokenWhenExpired(t *testing.T) {
	mockAuthStore := &mocks.MockAuthStore{
		GetPasswordResetTokenFunc: func(ctx context.Context, tokenHash string) (*store.PasswordResetToken, error) {
			return &store.PasswordResetToken{
				ID:        1,
				UserID:    1,
				ExpiresAt: time.Now().Add(-time.Hour),
				UsedAt:    nil,
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

	req := httptest.NewRequest(http.MethodGet, "/reset-password?token=expiredtoken", nil)
	userInfo := &auth.UserInfo{IsLoggedIn: false}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.GetResetPasswordHandler(rec, req)

	resetData, ok := capturedData.(ResetPasswordData)
	if !ok {
		t.Fatal("expected capturedData to be ResetPasswordData")
	}

	if !resetData.InvalidToken {
		t.Error("expected InvalidToken to be true for expired token")
	}
}

func TestPostResetPasswordHandler_ResetsPasswordSuccessfully(t *testing.T) {
	resetCalled := false
	mockAuthStore := &mocks.MockAuthStore{
		ResetPasswordWithTokenFunc: func(ctx context.Context, tokenHash string, newPasswordHash string) (int, error) {
			resetCalled = true
			return 1, nil
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
	form.Set("token", "validtoken")
	form.Set("password", "NewStrongPass123!")
	form.Set("confirm_password", "NewStrongPass123!")

	req := httptest.NewRequest(http.MethodPost, "/reset-password", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	userInfo := &auth.UserInfo{IsLoggedIn: false}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.PostResetPasswordHandler(rec, req)

	resetData, ok := capturedData.(ResetPasswordData)
	if !ok {
		t.Fatal("expected capturedData to be ResetPasswordData")
	}

	if resetData.Success == "" {
		t.Error("expected success message to be set")
	}

	if !resetCalled {
		t.Error("expected ResetPasswordWithToken to be called")
	}
}

func TestPostResetPasswordHandler_ShowsErrorWhenPasswordsDontMatch(t *testing.T) {
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
	form.Set("token", "validtoken")
	form.Set("password", "NewPassword123!")
	form.Set("confirm_password", "DifferentPassword123!")

	req := httptest.NewRequest(http.MethodPost, "/reset-password", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	userInfo := &auth.UserInfo{IsLoggedIn: false}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.PostResetPasswordHandler(rec, req)

	resetData, ok := capturedData.(ResetPasswordData)
	if !ok {
		t.Fatal("expected capturedData to be ResetPasswordData")
	}

	if resetData.Error != "Passwords do not match" {
		t.Errorf("expected error 'Passwords do not match', got '%s'", resetData.Error)
	}
}

func TestPostResetPasswordHandler_ShowsErrorWhenPasswordTooWeak(t *testing.T) {
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
	form.Set("token", "validtoken")
	form.Set("password", "weak")
	form.Set("confirm_password", "weak")

	req := httptest.NewRequest(http.MethodPost, "/reset-password", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	userInfo := &auth.UserInfo{IsLoggedIn: false}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.PostResetPasswordHandler(rec, req)

	resetData, ok := capturedData.(ResetPasswordData)
	if !ok {
		t.Fatal("expected capturedData to be ResetPasswordData")
	}

	if resetData.Error == "" {
		t.Error("expected error message for weak password")
	}
}

func TestPostResetPasswordHandler_ShowsInvalidTokenWhenTokenExpired(t *testing.T) {
	mockAuthStore := &mocks.MockAuthStore{
		ResetPasswordWithTokenFunc: func(ctx context.Context, tokenHash string, newPasswordHash string) (int, error) {
			return 0, errors.New("reset token has expired")
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
	form.Set("token", "expiredtoken")
	form.Set("password", "NewStrongPass123!")
	form.Set("confirm_password", "NewStrongPass123!")

	req := httptest.NewRequest(http.MethodPost, "/reset-password", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	userInfo := &auth.UserInfo{IsLoggedIn: false}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.PostResetPasswordHandler(rec, req)

	resetData, ok := capturedData.(ResetPasswordData)
	if !ok {
		t.Fatal("expected capturedData to be ResetPasswordData")
	}

	if !resetData.InvalidToken {
		t.Error("expected InvalidToken to be true for expired token")
	}
}

func TestPostResetPasswordHandler_ShowsInvalidTokenWhenAlreadyUsed(t *testing.T) {
	mockAuthStore := &mocks.MockAuthStore{
		ResetPasswordWithTokenFunc: func(ctx context.Context, tokenHash string, newPasswordHash string) (int, error) {
			return 0, errors.New("reset token has already been used")
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
	form.Set("token", "usedtoken")
	form.Set("password", "NewStrongPass123!")
	form.Set("confirm_password", "NewStrongPass123!")

	req := httptest.NewRequest(http.MethodPost, "/reset-password", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	userInfo := &auth.UserInfo{IsLoggedIn: false}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.PostResetPasswordHandler(rec, req)

	resetData, ok := capturedData.(ResetPasswordData)
	if !ok {
		t.Fatal("expected capturedData to be ResetPasswordData")
	}

	if !resetData.InvalidToken {
		t.Error("expected InvalidToken to be true for already used token")
	}
}
