package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mr-flannery/go-recipe-book/src/store"
	"github.com/mr-flannery/go-recipe-book/src/store/mocks"
)

func TestAuthenticate_Success(t *testing.T) {
	password := "TestPassword123!"
	hash, _ := HashPassword(password)

	mockStore := &mocks.MockAuthStore{
		GetUserByEmailFunc: func(email string) (*store.AuthUser, string, error) {
			return &store.AuthUser{
				ID:       1,
				Username: "testuser",
				Email:    email,
				IsAdmin:  false,
				IsActive: true,
			}, hash, nil
		},
		UpdateLastLoginFunc: func(userID int) error {
			return nil
		},
	}

	user, err := Authenticate(mockStore, "test@example.com", password)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user == nil {
		t.Fatal("expected user, got nil")
	}
	if user.ID != 1 {
		t.Errorf("expected user ID 1, got %d", user.ID)
	}
	if user.Username != "testuser" {
		t.Errorf("expected username 'testuser', got %s", user.Username)
	}
	if user.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got %s", user.Email)
	}
}

func TestAuthenticate_UserNotFound(t *testing.T) {
	mockStore := &mocks.MockAuthStore{
		GetUserByEmailFunc: func(email string) (*store.AuthUser, string, error) {
			return nil, "", errors.New("user not found")
		},
	}

	user, err := Authenticate(mockStore, "notfound@example.com", "password")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if user != nil {
		t.Fatal("expected nil user, got user")
	}
	if err.Error() != "invalid email or password" {
		t.Errorf("expected 'invalid email or password', got %s", err.Error())
	}
}

func TestAuthenticate_WrongPassword(t *testing.T) {
	hash, _ := HashPassword("CorrectPassword123!")

	mockStore := &mocks.MockAuthStore{
		GetUserByEmailFunc: func(email string) (*store.AuthUser, string, error) {
			return &store.AuthUser{
				ID:       1,
				Username: "testuser",
				Email:    email,
				IsAdmin:  false,
				IsActive: true,
			}, hash, nil
		},
	}

	user, err := Authenticate(mockStore, "test@example.com", "WrongPassword123!")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if user != nil {
		t.Fatal("expected nil user, got user")
	}
	if err.Error() != "invalid email or password" {
		t.Errorf("expected 'invalid email or password', got %s", err.Error())
	}
}

func TestAuthenticate_UpdateLastLoginError(t *testing.T) {
	password := "TestPassword123!"
	hash, _ := HashPassword(password)

	mockStore := &mocks.MockAuthStore{
		GetUserByEmailFunc: func(email string) (*store.AuthUser, string, error) {
			return &store.AuthUser{
				ID:       1,
				Username: "testuser",
				Email:    email,
				IsAdmin:  false,
				IsActive: true,
			}, hash, nil
		},
		UpdateLastLoginFunc: func(userID int) error {
			return errors.New("database error")
		},
	}

	user, err := Authenticate(mockStore, "test@example.com", password)
	if err != nil {
		t.Fatalf("expected no error (UpdateLastLogin failure is non-fatal), got %v", err)
	}
	if user == nil {
		t.Fatal("expected user, got nil")
	}
}

func TestGetUserBySession_Success(t *testing.T) {
	mockStore := &mocks.MockAuthStore{
		GetSessionFunc: func(sessionID string) (*store.Session, error) {
			return &store.Session{
				ID:     sessionID,
				UserID: 1,
			}, nil
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

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "test-session-id"})

	user, err := GetUserBySession(mockStore, req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if user == nil {
		t.Fatal("expected user, got nil")
	}
	if user.ID != 1 {
		t.Errorf("expected user ID 1, got %d", user.ID)
	}
	if !user.IsAdmin {
		t.Error("expected user to be admin")
	}
}

func TestGetUserBySession_NoCookie(t *testing.T) {
	mockStore := &mocks.MockAuthStore{}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	user, err := GetUserBySession(mockStore, req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if user != nil {
		t.Fatal("expected nil user, got user")
	}
}

func TestGetUserBySession_InvalidSession(t *testing.T) {
	mockStore := &mocks.MockAuthStore{
		GetSessionFunc: func(sessionID string) (*store.Session, error) {
			return nil, errors.New("session not found")
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "invalid-session"})

	user, err := GetUserBySession(mockStore, req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if user != nil {
		t.Fatal("expected nil user, got user")
	}
}

func TestGetUserBySession_UserNotFound(t *testing.T) {
	mockStore := &mocks.MockAuthStore{
		GetSessionFunc: func(sessionID string) (*store.Session, error) {
			return &store.Session{
				ID:     sessionID,
				UserID: 999,
			}, nil
		},
		GetUserByIDFunc: func(userID int) (*store.AuthUser, error) {
			return nil, errors.New("user not found")
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})

	user, err := GetUserBySession(mockStore, req)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if user != nil {
		t.Fatal("expected nil user, got user")
	}
}

func TestIsSessionValid(t *testing.T) {
	tests := []struct {
		name     string
		mock     *mocks.MockAuthStore
		cookie   *http.Cookie
		expected bool
	}{
		{
			name: "valid session",
			mock: &mocks.MockAuthStore{
				GetSessionFunc: func(sessionID string) (*store.Session, error) {
					return &store.Session{ID: sessionID, UserID: 1}, nil
				},
				GetUserByIDFunc: func(userID int) (*store.AuthUser, error) {
					return &store.AuthUser{ID: userID}, nil
				},
			},
			cookie:   &http.Cookie{Name: "session", Value: "valid"},
			expected: true,
		},
		{
			name:     "no cookie",
			mock:     &mocks.MockAuthStore{},
			cookie:   nil,
			expected: false,
		},
		{
			name: "invalid session",
			mock: &mocks.MockAuthStore{
				GetSessionFunc: func(sessionID string) (*store.Session, error) {
					return nil, errors.New("not found")
				},
			},
			cookie:   &http.Cookie{Name: "session", Value: "invalid"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}

			result := IsSessionValid(tt.mock, req)
			if result != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestGetUserIDByUsername(t *testing.T) {
	mockStore := &mocks.MockAuthStore{
		GetUserIDByUsernameFunc: func(username string) (int, error) {
			if username == "admin" {
				return 1, nil
			}
			return 0, errors.New("user not found")
		},
	}

	id, err := GetUserIDByUsername(mockStore, "admin")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != 1 {
		t.Errorf("expected ID 1, got %d", id)
	}

	_, err = GetUserIDByUsername(mockStore, "unknown")
	if err == nil {
		t.Fatal("expected error for unknown user, got nil")
	}
}

func TestRequireAuth(t *testing.T) {
	middleware := RequireAuth()

	t.Run("redirects when not logged in", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/protected?foo=bar", nil)
		rec := httptest.NewRecorder()

		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t.Error("handler should not be called when not logged in")
		}))

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusSeeOther {
			t.Errorf("expected status %d, got %d", http.StatusSeeOther, rec.Code)
		}

		location := rec.Header().Get("Location")
		expected := "/login?redirect=/protected?foo=bar"
		if location != expected {
			t.Errorf("expected redirect to %s, got %s", expected, location)
		}
	})

	t.Run("allows access when logged in", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		userInfo := &UserInfo{IsLoggedIn: true, UserID: 1, Username: "testuser"}
		req = req.WithContext(context.WithValue(req.Context(), userInfoKey, userInfo))
		rec := httptest.NewRecorder()

		called := false
		handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusOK)
		}))

		handler.ServeHTTP(rec, req)

		if !called {
			t.Error("handler should be called when logged in")
		}
		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
	})
}

func TestUserFromAuthUser(t *testing.T) {
	t.Run("nil input returns nil", func(t *testing.T) {
		result := userFromAuthUser(nil)
		if result != nil {
			t.Error("expected nil, got non-nil")
		}
	})

	t.Run("converts correctly", func(t *testing.T) {
		authUser := &store.AuthUser{
			ID:       42,
			Username: "testuser",
			Email:    "test@example.com",
			IsAdmin:  true,
			IsActive: true,
		}

		result := userFromAuthUser(authUser)
		if result == nil {
			t.Fatal("expected non-nil, got nil")
		}
		if result.ID != 42 {
			t.Errorf("expected ID 42, got %d", result.ID)
		}
		if result.Username != "testuser" {
			t.Errorf("expected username 'testuser', got %s", result.Username)
		}
		if result.Email != "test@example.com" {
			t.Errorf("expected email 'test@example.com', got %s", result.Email)
		}
		if !result.IsAdmin {
			t.Error("expected IsAdmin true")
		}
		if !result.IsActive {
			t.Error("expected IsActive true")
		}
	})
}
