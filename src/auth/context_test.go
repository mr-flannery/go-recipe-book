package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mr-flannery/go-recipe-book/src/store"
	"github.com/mr-flannery/go-recipe-book/src/store/mocks"
)

func TestUserContextMiddleware_ValidSession(t *testing.T) {
	mockStore := &mocks.MockAuthStore{
		GetSessionFunc: func(sessionID string) (*store.Session, error) {
			return &store.Session{ID: sessionID, UserID: 1}, nil
		},
		GetUserByIDFunc: func(userID int) (*store.AuthUser, error) {
			return &store.AuthUser{
				ID:       userID,
				Username: "testuser",
				Email:    "test@example.com",
				IsAdmin:  true,
				IsActive: true,
			}, nil
		},
	}

	middleware := UserContextMiddleware(mockStore)
	var capturedUserInfo *UserInfo

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserInfo = GetUserInfoFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "valid-session"})
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if capturedUserInfo == nil {
		t.Fatal("expected user info in context")
	}
	if !capturedUserInfo.IsLoggedIn {
		t.Error("expected IsLoggedIn to be true")
	}
	if !capturedUserInfo.IsAdmin {
		t.Error("expected IsAdmin to be true")
	}
	if capturedUserInfo.Username != "testuser" {
		t.Errorf("expected username 'testuser', got %s", capturedUserInfo.Username)
	}
	if capturedUserInfo.UserID != 1 {
		t.Errorf("expected UserID 1, got %d", capturedUserInfo.UserID)
	}
}

func TestUserContextMiddleware_NoSession(t *testing.T) {
	mockStore := &mocks.MockAuthStore{}
	middleware := UserContextMiddleware(mockStore)
	var capturedUserInfo *UserInfo

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		capturedUserInfo = GetUserInfoFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if capturedUserInfo == nil {
		t.Fatal("expected user info in context even without session")
	}
	if capturedUserInfo.IsLoggedIn {
		t.Error("expected IsLoggedIn to be false")
	}
	if capturedUserInfo.IsAdmin {
		t.Error("expected IsAdmin to be false")
	}
	if capturedUserInfo.Username != "" {
		t.Errorf("expected empty username, got %s", capturedUserInfo.Username)
	}
	if capturedUserInfo.UserID != 0 {
		t.Errorf("expected UserID 0, got %d", capturedUserInfo.UserID)
	}
}

func TestGetUserInfoFromContext_WithUserInfo(t *testing.T) {
	userInfo := &UserInfo{
		IsLoggedIn: true,
		IsAdmin:    false,
		Username:   "contextuser",
		UserID:     42,
	}
	ctx := context.WithValue(context.Background(), userInfoKey, userInfo)

	result := GetUserInfoFromContext(ctx)
	if result == nil {
		t.Fatal("expected user info")
	}
	if result.Username != "contextuser" {
		t.Errorf("expected 'contextuser', got %s", result.Username)
	}
	if result.UserID != 42 {
		t.Errorf("expected UserID 42, got %d", result.UserID)
	}
}

func TestGetUserInfoFromContext_NoUserInfo(t *testing.T) {
	ctx := context.Background()

	result := GetUserInfoFromContext(ctx)
	if result == nil {
		t.Fatal("expected default user info")
	}
	if result.IsLoggedIn {
		t.Error("expected IsLoggedIn to be false")
	}
	if result.Username != "" {
		t.Errorf("expected empty username, got %s", result.Username)
	}
}

func TestGetUserFromContext_LoggedIn(t *testing.T) {
	userInfo := &UserInfo{
		IsLoggedIn: true,
		IsAdmin:    true,
		Username:   "admin",
		UserID:     1,
	}
	ctx := context.WithValue(context.Background(), userInfoKey, userInfo)

	user := GetUserFromContext(ctx)
	if user == nil {
		t.Fatal("expected user")
	}
	if user.ID != 1 {
		t.Errorf("expected ID 1, got %d", user.ID)
	}
	if user.Username != "admin" {
		t.Errorf("expected 'admin', got %s", user.Username)
	}
	if !user.IsAdmin {
		t.Error("expected IsAdmin to be true")
	}
}

func TestGetUserFromContext_NotLoggedIn(t *testing.T) {
	userInfo := &UserInfo{
		IsLoggedIn: false,
	}
	ctx := context.WithValue(context.Background(), userInfoKey, userInfo)

	user := GetUserFromContext(ctx)
	if user != nil {
		t.Error("expected nil user when not logged in")
	}
}

func TestIsUserAdmin(t *testing.T) {
	tests := []struct {
		name     string
		userInfo *UserInfo
		expected bool
	}{
		{
			name:     "logged in admin",
			userInfo: &UserInfo{IsLoggedIn: true, IsAdmin: true},
			expected: true,
		},
		{
			name:     "logged in non-admin",
			userInfo: &UserInfo{IsLoggedIn: true, IsAdmin: false},
			expected: false,
		},
		{
			name:     "not logged in with admin flag",
			userInfo: &UserInfo{IsLoggedIn: false, IsAdmin: true},
			expected: false,
		},
		{
			name:     "not logged in",
			userInfo: &UserInfo{IsLoggedIn: false, IsAdmin: false},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.WithValue(context.Background(), userInfoKey, tt.userInfo)
			result := IsUserAdmin(ctx)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetUsernameFromContext(t *testing.T) {
	userInfo := &UserInfo{Username: "myuser"}
	ctx := context.WithValue(context.Background(), userInfoKey, userInfo)

	result := GetUsernameFromContext(ctx)
	if result != "myuser" {
		t.Errorf("expected 'myuser', got %s", result)
	}
}

func TestGetUserIDFromContext(t *testing.T) {
	userInfo := &UserInfo{UserID: 123}
	ctx := context.WithValue(context.Background(), userInfoKey, userInfo)

	result := GetUserIDFromContext(ctx)
	if result != 123 {
		t.Errorf("expected 123, got %d", result)
	}
}
