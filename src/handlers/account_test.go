package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/mr-flannery/go-recipe-book/src/auth"
	"github.com/mr-flannery/go-recipe-book/src/models"
	"github.com/mr-flannery/go-recipe-book/src/store"
	"github.com/mr-flannery/go-recipe-book/src/store/mocks"
	tmocks "github.com/mr-flannery/go-recipe-book/src/templates/mocks"
)

type MockUserPreferencesStore struct {
	GetFunc         func(userID int) (*models.UserPreferences, error)
	SetPageSizeFunc func(userID, pageSize int) error
	SetViewModeFunc func(userID int, viewMode string) error
}

func (m *MockUserPreferencesStore) Get(userID int) (*models.UserPreferences, error) {
	if m.GetFunc != nil {
		return m.GetFunc(userID)
	}
	return nil, nil
}

func (m *MockUserPreferencesStore) SetPageSize(userID, pageSize int) error {
	if m.SetPageSizeFunc != nil {
		return m.SetPageSizeFunc(userID, pageSize)
	}
	return nil
}

func (m *MockUserPreferencesStore) SetViewMode(userID int, viewMode string) error {
	if m.SetViewModeFunc != nil {
		return m.SetViewModeFunc(userID, viewMode)
	}
	return nil
}

func TestGetAccountSettingsHandler_RendersAccountSettingsPage(t *testing.T) {
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

	req := httptest.NewRequest(http.MethodGet, "/account", nil)
	userInfo := &auth.UserInfo{IsLoggedIn: true, UserID: 1, Username: "testuser"}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.GetAccountSettingsHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if capturedTemplate != "account-settings.gohtml" {
		t.Errorf("expected template account-settings.gohtml, got %s", capturedTemplate)
	}

	data, ok := capturedData.(AccountSettingsData)
	if !ok {
		t.Fatal("expected capturedData to be AccountSettingsData")
	}

	if data.UserInfo == nil || data.UserInfo.Username != "testuser" {
		t.Error("expected UserInfo to be populated with testuser")
	}
}

func TestGetAccountSettingsHandler_IncludesSuccessMessage(t *testing.T) {
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

	req := httptest.NewRequest(http.MethodGet, "/account?success=Settings+saved", nil)
	userInfo := &auth.UserInfo{IsLoggedIn: true, UserID: 1, Username: "testuser"}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.GetAccountSettingsHandler(rec, req)

	data, ok := capturedData.(AccountSettingsData)
	if !ok {
		t.Fatal("expected capturedData to be AccountSettingsData")
	}

	if data.Success != "Settings saved" {
		t.Errorf("expected success message 'Settings saved', got '%s'", data.Success)
	}
}

func TestGetAccountSettingsHandler_IncludesErrorMessage(t *testing.T) {
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

	req := httptest.NewRequest(http.MethodGet, "/account?error=Something+went+wrong", nil)
	userInfo := &auth.UserInfo{IsLoggedIn: true, UserID: 1, Username: "testuser"}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.GetAccountSettingsHandler(rec, req)

	data, ok := capturedData.(AccountSettingsData)
	if !ok {
		t.Fatal("expected capturedData to be AccountSettingsData")
	}

	if data.Error != "Something went wrong" {
		t.Errorf("expected error message 'Something went wrong', got '%s'", data.Error)
	}
}

func TestExportUserDataHandler_ReturnsUnauthorizedWhenNotLoggedIn(t *testing.T) {
	var capturedStatus int
	mockRenderer := &tmocks.MockRenderer{
		RenderErrorFunc: func(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
			capturedStatus = statusCode
			w.WriteHeader(statusCode)
		},
	}

	h := &Handler{
		Renderer: mockRenderer,
	}

	req := httptest.NewRequest(http.MethodGet, "/account/export", nil)
	userInfo := &auth.UserInfo{IsLoggedIn: false}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.ExportUserDataHandler(rec, req)

	if capturedStatus != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, capturedStatus)
	}
}

func TestExportUserDataHandler_ReturnsErrorWhenUserNotFound(t *testing.T) {
	mockAuthStore := &mocks.MockAuthStore{
		GetFullUserByIDFunc: func(userID int) (*store.FullAuthUser, error) {
			return nil, errors.New("user not found")
		},
	}

	var capturedStatus int
	mockRenderer := &tmocks.MockRenderer{
		RenderErrorFunc: func(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
			capturedStatus = statusCode
			w.WriteHeader(statusCode)
		},
	}

	h := &Handler{
		AuthStore: mockAuthStore,
		Renderer:  mockRenderer,
	}

	req := httptest.NewRequest(http.MethodGet, "/account/export", nil)
	userInfo := &auth.UserInfo{IsLoggedIn: true, UserID: 999, Username: "ghost"}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.ExportUserDataHandler(rec, req)

	if capturedStatus != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, capturedStatus)
	}
}

func TestExportUserDataHandler_ExportsUserDataAsJSON(t *testing.T) {
	now := time.Now()
	lastLogin := now.Add(-24 * time.Hour)

	mockAuthStore := &mocks.MockAuthStore{
		GetFullUserByIDFunc: func(userID int) (*store.FullAuthUser, error) {
			return &store.FullAuthUser{
				ID:        1,
				Username:  "testuser",
				Email:     "test@example.com",
				IsAdmin:   false,
				IsActive:  true,
				CreatedAt: now.Add(-30 * 24 * time.Hour),
				LastLogin: &lastLogin,
			}, nil
		},
	}

	mockUserPreferencesStore := &MockUserPreferencesStore{
		GetFunc: func(userID int) (*models.UserPreferences, error) {
			return &models.UserPreferences{
				UserID:   1,
				PageSize: 25,
			}, nil
		},
	}

	mockRecipeStore := &mocks.MockRecipeStore{
		GetFilteredFunc: func(params models.FilterParams) ([]models.Recipe, error) {
			return []models.Recipe{
				{
					ID:             1,
					Title:          "Test Recipe",
					IngredientsMD:  "- ingredient 1",
					InstructionsMD: "1. Do stuff",
					PrepTime:       10,
					CookTime:       20,
					Calories:       300,
					Tags:           []models.Tag{{Name: "dinner"}},
					CreatedAt:      now,
					UpdatedAt:      now,
				},
			}, nil
		},
	}

	mockCommentStore := &mocks.MockCommentStore{
		GetByUserIDFunc: func(userID int) ([]models.Comment, error) {
			return []models.Comment{
				{
					ID:        1,
					RecipeID:  1,
					ContentMD: "Great recipe!",
					CreatedAt: now,
					UpdatedAt: now,
				},
			}, nil
		},
	}

	mockUserTagStore := &mocks.MockUserTagStore{
		GetByUserIDFunc: func(userID int) ([]models.UserTag, error) {
			return []models.UserTag{
				{
					ID:       1,
					RecipeID: 1,
					Name:     "favorite",
				},
			}, nil
		},
	}

	h := &Handler{
		AuthStore:            mockAuthStore,
		UserPreferencesStore: mockUserPreferencesStore,
		RecipeStore:          mockRecipeStore,
		CommentStore:         mockCommentStore,
		UserTagStore:         mockUserTagStore,
	}

	req := httptest.NewRequest(http.MethodGet, "/account/export", nil)
	userInfo := &auth.UserInfo{IsLoggedIn: true, UserID: 1, Username: "testuser"}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.ExportUserDataHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	contentDisposition := rec.Header().Get("Content-Disposition")
	if !strings.HasPrefix(contentDisposition, "attachment; filename=\"recipe-book-data-") {
		t.Errorf("expected Content-Disposition with attachment filename, got %s", contentDisposition)
	}

	var export UserDataExport
	if err := json.Unmarshal(rec.Body.Bytes(), &export); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}

	if export.Account.Username != "testuser" {
		t.Errorf("expected username testuser, got %s", export.Account.Username)
	}

	if export.Account.Email != "test@example.com" {
		t.Errorf("expected email test@example.com, got %s", export.Account.Email)
	}

	if export.Preferences == nil || export.Preferences.PageSize != 25 {
		t.Error("expected preferences with PageSize 25")
	}

	if len(export.Recipes) != 1 || export.Recipes[0].Title != "Test Recipe" {
		t.Error("expected one recipe with title 'Test Recipe'")
	}

	if len(export.Comments) != 1 || export.Comments[0].Content != "Great recipe!" {
		t.Error("expected one comment with content 'Great recipe!'")
	}

	if len(export.UserTags) != 1 || export.UserTags[0].Name != "favorite" {
		t.Error("expected one user tag with name 'favorite'")
	}
}

func TestExportUserDataHandler_HandlesEmptyData(t *testing.T) {
	now := time.Now()

	mockAuthStore := &mocks.MockAuthStore{
		GetFullUserByIDFunc: func(userID int) (*store.FullAuthUser, error) {
			return &store.FullAuthUser{
				ID:        1,
				Username:  "newuser",
				Email:     "new@example.com",
				CreatedAt: now,
				LastLogin: nil,
			}, nil
		},
	}

	mockUserPreferencesStore := &MockUserPreferencesStore{
		GetFunc: func(userID int) (*models.UserPreferences, error) {
			return nil, errors.New("no preferences")
		},
	}

	mockRecipeStore := &mocks.MockRecipeStore{
		GetFilteredFunc: func(params models.FilterParams) ([]models.Recipe, error) {
			return []models.Recipe{}, nil
		},
	}

	mockCommentStore := &mocks.MockCommentStore{
		GetByUserIDFunc: func(userID int) ([]models.Comment, error) {
			return []models.Comment{}, nil
		},
	}

	mockUserTagStore := &mocks.MockUserTagStore{
		GetByUserIDFunc: func(userID int) ([]models.UserTag, error) {
			return []models.UserTag{}, nil
		},
	}

	h := &Handler{
		AuthStore:            mockAuthStore,
		UserPreferencesStore: mockUserPreferencesStore,
		RecipeStore:          mockRecipeStore,
		CommentStore:         mockCommentStore,
		UserTagStore:         mockUserTagStore,
	}

	req := httptest.NewRequest(http.MethodGet, "/account/export", nil)
	userInfo := &auth.UserInfo{IsLoggedIn: true, UserID: 1, Username: "newuser"}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.ExportUserDataHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var export UserDataExport
	if err := json.Unmarshal(rec.Body.Bytes(), &export); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}

	if export.Preferences != nil {
		t.Error("expected nil preferences")
	}

	if export.Account.LastLogin != nil {
		t.Error("expected nil last_login")
	}

	if len(export.Recipes) != 0 {
		t.Error("expected empty recipes")
	}

	if len(export.Comments) != 0 {
		t.Error("expected empty comments")
	}

	if len(export.UserTags) != 0 {
		t.Error("expected empty user tags")
	}
}

func TestDeleteOwnAccountHandler_ReturnsUnauthorizedWhenNotLoggedIn(t *testing.T) {
	var capturedStatus int
	mockRenderer := &tmocks.MockRenderer{
		RenderErrorFunc: func(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
			capturedStatus = statusCode
			w.WriteHeader(statusCode)
		},
	}

	h := &Handler{
		Renderer: mockRenderer,
	}

	req := httptest.NewRequest(http.MethodPost, "/account/delete", nil)
	userInfo := &auth.UserInfo{IsLoggedIn: false}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.DeleteOwnAccountHandler(rec, req)

	if capturedStatus != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, capturedStatus)
	}
}

func TestDeleteOwnAccountHandler_RedirectsWhenAdminTriesToDelete(t *testing.T) {
	h := &Handler{}

	req := httptest.NewRequest(http.MethodPost, "/account/delete", nil)
	userInfo := &auth.UserInfo{IsLoggedIn: true, UserID: 1, Username: "admin", IsAdmin: true}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.DeleteOwnAccountHandler(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rec.Code)
	}

	location := rec.Header().Get("Location")
	if !strings.Contains(location, "/account") || !strings.Contains(location, "error=") {
		t.Errorf("expected redirect to /account with error, got %s", location)
	}

	if !strings.Contains(location, "Admin") {
		t.Error("expected error message to mention Admin accounts")
	}
}

func TestDeleteOwnAccountHandler_RedirectsWhenConfirmDeleteMissing(t *testing.T) {
	h := &Handler{}

	form := url.Values{}
	form.Set("password", "somepassword")

	req := httptest.NewRequest(http.MethodPost, "/account/delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	userInfo := &auth.UserInfo{IsLoggedIn: true, UserID: 1, Username: "testuser", IsAdmin: false}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.DeleteOwnAccountHandler(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rec.Code)
	}

	location := rec.Header().Get("Location")
	if !strings.Contains(location, "DELETE") {
		t.Errorf("expected error about typing DELETE, got %s", location)
	}
}

func TestDeleteOwnAccountHandler_RedirectsWhenConfirmDeleteWrong(t *testing.T) {
	h := &Handler{}

	form := url.Values{}
	form.Set("password", "somepassword")
	form.Set("confirm_delete", "delete")

	req := httptest.NewRequest(http.MethodPost, "/account/delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	userInfo := &auth.UserInfo{IsLoggedIn: true, UserID: 1, Username: "testuser", IsAdmin: false}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.DeleteOwnAccountHandler(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rec.Code)
	}

	location := rec.Header().Get("Location")
	if !strings.Contains(location, "DELETE") {
		t.Errorf("expected error about typing DELETE, got %s", location)
	}
}

func TestDeleteOwnAccountHandler_RedirectsWhenUserNotFound(t *testing.T) {
	mockAuthStore := &mocks.MockAuthStore{
		GetUserByIDFunc: func(userID int) (*store.AuthUser, error) {
			return nil, errors.New("user not found")
		},
	}

	h := &Handler{
		AuthStore: mockAuthStore,
	}

	form := url.Values{}
	form.Set("password", "somepassword")
	form.Set("confirm_delete", "DELETE")

	req := httptest.NewRequest(http.MethodPost, "/account/delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	userInfo := &auth.UserInfo{IsLoggedIn: true, UserID: 999, Username: "ghost", IsAdmin: false}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.DeleteOwnAccountHandler(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rec.Code)
	}

	location := rec.Header().Get("Location")
	if !strings.Contains(location, "error=") {
		t.Errorf("expected error in redirect, got %s", location)
	}
}

func TestDeleteOwnAccountHandler_RedirectsWhenPasswordIncorrect(t *testing.T) {
	validPassword := "Correct#Pass1"
	hashedPassword, err := auth.HashPassword(validPassword)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	mockAuthStore := &mocks.MockAuthStore{
		GetUserByIDFunc: func(userID int) (*store.AuthUser, error) {
			return &store.AuthUser{
				ID:       1,
				Username: "testuser",
				Email:    "test@example.com",
			}, nil
		},
		GetUserByEmailFunc: func(email string) (*store.AuthUser, string, error) {
			return &store.AuthUser{
				ID:       1,
				Username: "testuser",
				Email:    email,
			}, hashedPassword, nil
		},
	}

	h := &Handler{
		AuthStore: mockAuthStore,
	}

	form := url.Values{}
	form.Set("password", "WrongPassword123!")
	form.Set("confirm_delete", "DELETE")

	req := httptest.NewRequest(http.MethodPost, "/account/delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	userInfo := &auth.UserInfo{IsLoggedIn: true, UserID: 1, Username: "testuser", IsAdmin: false}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.DeleteOwnAccountHandler(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rec.Code)
	}

	location := rec.Header().Get("Location")
	if !strings.Contains(location, "password") {
		t.Errorf("expected error about password, got %s", location)
	}
}

func TestDeleteOwnAccountHandler_DeletesAccountSuccessfully(t *testing.T) {
	validPassword := "Correct#Pass1"
	hashedPassword, err := auth.HashPassword(validPassword)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	deleteCalled := false
	deleteSessionsCalled := false

	mockAuthStore := &mocks.MockAuthStore{
		GetUserByIDFunc: func(userID int) (*store.AuthUser, error) {
			return &store.AuthUser{
				ID:       1,
				Username: "testuser",
				Email:    "test@example.com",
			}, nil
		},
		GetUserByEmailFunc: func(email string) (*store.AuthUser, string, error) {
			return &store.AuthUser{
				ID:       1,
				Username: "testuser",
				Email:    email,
			}, hashedPassword, nil
		},
		DeleteUserFunc: func(userID int) error {
			deleteCalled = true
			return nil
		},
		DeleteUserSessionsFunc: func(userID int) error {
			deleteSessionsCalled = true
			return nil
		},
	}

	h := &Handler{
		AuthStore: mockAuthStore,
	}

	form := url.Values{}
	form.Set("password", validPassword)
	form.Set("confirm_delete", "DELETE")

	req := httptest.NewRequest(http.MethodPost, "/account/delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	userInfo := &auth.UserInfo{IsLoggedIn: true, UserID: 1, Username: "testuser", IsAdmin: false}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.DeleteOwnAccountHandler(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rec.Code)
	}

	location := rec.Header().Get("Location")
	if !strings.HasPrefix(location, "/?account_deleted=true") {
		t.Errorf("expected redirect to /?account_deleted=true, got %s", location)
	}

	if !deleteCalled {
		t.Error("expected DeleteUser to be called")
	}

	cookies := rec.Result().Cookies()
	sessionCleared := false
	for _, c := range cookies {
		if c.Name == "session" && c.MaxAge < 0 {
			sessionCleared = true
		}
	}
	if !sessionCleared {
		t.Error("expected session cookie to be cleared")
	}

	_ = deleteSessionsCalled
}

func TestDeleteOwnAccountHandler_RedirectsWhenDeleteFails(t *testing.T) {
	validPassword := "Correct#Pass1"
	hashedPassword, err := auth.HashPassword(validPassword)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	mockAuthStore := &mocks.MockAuthStore{
		GetUserByIDFunc: func(userID int) (*store.AuthUser, error) {
			return &store.AuthUser{
				ID:       1,
				Username: "testuser",
				Email:    "test@example.com",
			}, nil
		},
		GetUserByEmailFunc: func(email string) (*store.AuthUser, string, error) {
			return &store.AuthUser{
				ID:       1,
				Username: "testuser",
				Email:    email,
			}, hashedPassword, nil
		},
		DeleteUserFunc: func(userID int) error {
			return errors.New("database error")
		},
		DeleteUserSessionsFunc: func(userID int) error {
			return nil
		},
	}

	h := &Handler{
		AuthStore: mockAuthStore,
	}

	form := url.Values{}
	form.Set("password", validPassword)
	form.Set("confirm_delete", "DELETE")

	req := httptest.NewRequest(http.MethodPost, "/account/delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	userInfo := &auth.UserInfo{IsLoggedIn: true, UserID: 1, Username: "testuser", IsAdmin: false}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.DeleteOwnAccountHandler(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rec.Code)
	}

	location := rec.Header().Get("Location")
	if !strings.Contains(location, "/account") || !strings.Contains(location, "error=") {
		t.Errorf("expected redirect to /account with error, got %s", location)
	}
}
