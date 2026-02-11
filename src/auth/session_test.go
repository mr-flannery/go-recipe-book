package auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mr-flannery/go-recipe-book/src/store"
	"github.com/mr-flannery/go-recipe-book/src/store/mocks"
)

func TestCreateSession_ReturnsSessionWhenStoreSucceeds(t *testing.T) {
	var capturedSession *store.Session
	mockStore := &mocks.MockAuthStore{
		CreateSessionFunc: func(session *store.Session) error {
			capturedSession = session
			return nil
		},
	}

	session, err := CreateSession(mockStore, 1, "192.168.1.1", "Mozilla/5.0")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if session == nil {
		t.Fatal("expected session, got nil")
	}
	if session.UserID != 1 {
		t.Errorf("expected UserID 1, got %d", session.UserID)
	}
	if session.IPAddress != "192.168.1.1" {
		t.Errorf("expected IPAddress '192.168.1.1', got %s", session.IPAddress)
	}
	if session.UserAgent != "Mozilla/5.0" {
		t.Errorf("expected UserAgent 'Mozilla/5.0', got %s", session.UserAgent)
	}
	if len(session.ID) != 64 { // 32 bytes hex encoded = 64 chars
		t.Errorf("expected session ID length 64, got %d", len(session.ID))
	}
	if capturedSession == nil {
		t.Fatal("expected store to receive session")
	}
	if capturedSession.ID != session.ID {
		t.Error("store should receive same session ID")
	}
}

func TestCreateSession_ReturnsErrorWhenStoreFails(t *testing.T) {
	mockStore := &mocks.MockAuthStore{
		CreateSessionFunc: func(session *store.Session) error {
			return errors.New("database error")
		},
	}

	session, err := CreateSession(mockStore, 1, "192.168.1.1", "Mozilla/5.0")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if session != nil {
		t.Fatal("expected nil session on error")
	}
}

func TestValidateSession_ReturnsSessionWhenItExists(t *testing.T) {
	mockStore := &mocks.MockAuthStore{
		GetSessionFunc: func(sessionID string) (*store.Session, error) {
			return &store.Session{
				ID:        sessionID,
				UserID:    42,
				IPAddress: "10.0.0.1",
				UserAgent: "TestAgent",
			}, nil
		},
	}

	session, err := ValidateSession(mockStore, "test-session-id")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if session == nil {
		t.Fatal("expected session, got nil")
	}
	if session.ID != "test-session-id" {
		t.Errorf("expected ID 'test-session-id', got %s", session.ID)
	}
	if session.UserID != 42 {
		t.Errorf("expected UserID 42, got %d", session.UserID)
	}
}

func TestValidateSession_ReturnsErrorWhenSessionNotFound(t *testing.T) {
	mockStore := &mocks.MockAuthStore{
		GetSessionFunc: func(sessionID string) (*store.Session, error) {
			return nil, errors.New("session not found")
		},
	}

	session, err := ValidateSession(mockStore, "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if session != nil {
		t.Fatal("expected nil session")
	}
}

func TestInvalidateSession_DeletesSessionFromStore(t *testing.T) {
	deletedSessionID := ""
	mockStore := &mocks.MockAuthStore{
		DeleteSessionFunc: func(sessionID string) error {
			deletedSessionID = sessionID
			return nil
		},
	}

	err := InvalidateSession(mockStore, "session-to-delete")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if deletedSessionID != "session-to-delete" {
		t.Errorf("expected 'session-to-delete', got %s", deletedSessionID)
	}
}

func TestInvalidateAllUserSessions_DeletesAllSessionsForUser(t *testing.T) {
	deletedUserID := 0
	mockStore := &mocks.MockAuthStore{
		DeleteUserSessionsFunc: func(userID int) error {
			deletedUserID = userID
			return nil
		},
	}

	err := InvalidateAllUserSessions(mockStore, 123)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if deletedUserID != 123 {
		t.Errorf("expected userID 123, got %d", deletedUserID)
	}
}

func TestCleanupExpiredSessions_RemovesExpiredSessionsFromStore(t *testing.T) {
	t.Run("successfully removes expired sessions", func(t *testing.T) {
		mockStore := &mocks.MockAuthStore{
			DeleteExpiredSessionsFunc: func() (int64, error) {
				return 5, nil
			},
		}

		err := CleanupExpiredSessions(mockStore)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	})

	t.Run("returns error when store fails", func(t *testing.T) {
		mockStore := &mocks.MockAuthStore{
			DeleteExpiredSessionsFunc: func() (int64, error) {
				return 0, errors.New("database error")
			},
		}

		err := CleanupExpiredSessions(mockStore)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestExtendSession_ExtendsSessionExpiryInStore(t *testing.T) {
	extendedSessionID := ""
	mockStore := &mocks.MockAuthStore{
		ExtendSessionFunc: func(sessionID string) error {
			extendedSessionID = sessionID
			return nil
		},
	}

	err := ExtendSession(mockStore, "extend-me")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if extendedSessionID != "extend-me" {
		t.Errorf("expected 'extend-me', got %s", extendedSessionID)
	}
}

func TestGetActiveSessionCount_ReturnsCountOfActiveSessionsForUser(t *testing.T) {
	mockStore := &mocks.MockAuthStore{
		GetActiveSessionCountFunc: func(userID int) (int, error) {
			if userID == 1 {
				return 3, nil
			}
			return 0, nil
		},
	}

	count, err := GetActiveSessionCount(mockStore, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if count != 3 {
		t.Errorf("expected count 3, got %d", count)
	}
}

func TestSetSecureSessionCookie_SetsHttpOnlyStrictCookie(t *testing.T) {
	rec := httptest.NewRecorder()
	SetSecureSessionCookie(rec, "test-session-id")

	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected cookie to be set")
	}

	cookie := cookies[0]
	if cookie.Value != "test-session-id" {
		t.Errorf("expected value 'test-session-id', got %s", cookie.Value)
	}
	if !cookie.HttpOnly {
		t.Error("expected HttpOnly to be true")
	}
	if cookie.SameSite != http.SameSiteStrictMode {
		t.Error("expected SameSite Strict")
	}
}

func TestClearSessionCookie_SetsExpiredCookieToRemoveIt(t *testing.T) {
	rec := httptest.NewRecorder()
	ClearSessionCookie(rec)

	cookies := rec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected cookie to be set")
	}

	cookie := cookies[0]
	if cookie.Value != "" {
		t.Errorf("expected empty value, got %s", cookie.Value)
	}
	if cookie.MaxAge != -1 {
		t.Errorf("expected MaxAge -1, got %d", cookie.MaxAge)
	}
}

func TestGetSessionFromRequest_ExtractsSessionIDFromCookie(t *testing.T) {
	t.Run("returns session ID when cookie is present", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.AddCookie(&http.Cookie{Name: "session", Value: "my-session-id"})

		sessionID, err := GetSessionFromRequest(req)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if sessionID != "my-session-id" {
			t.Errorf("expected 'my-session-id', got %s", sessionID)
		}
	})

	t.Run("returns error when cookie is missing", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)

		_, err := GetSessionFromRequest(req)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestGetClientIP_ExtractsIPFromHeadersOrRemoteAddr(t *testing.T) {
	tests := []struct {
		name       string
		headers    map[string]string
		remoteAddr string
		expected   string
	}{
		{
			name:       "returns IP from X-Forwarded-For header",
			headers:    map[string]string{"X-Forwarded-For": "203.0.113.195"},
			remoteAddr: "192.168.1.1:12345",
			expected:   "203.0.113.195",
		},
		{
			name:       "returns IP from X-Real-IP header",
			headers:    map[string]string{"X-Real-IP": "198.51.100.178"},
			remoteAddr: "192.168.1.1:12345",
			expected:   "198.51.100.178",
		},
		{
			name:       "returns IP from RemoteAddr when no headers",
			headers:    map[string]string{},
			remoteAddr: "10.0.0.1:54321",
			expected:   "10.0.0.1",
		},
		{
			name:       "prioritizes X-Forwarded-For over X-Real-IP",
			headers:    map[string]string{"X-Forwarded-For": "1.2.3.4", "X-Real-IP": "5.6.7.8"},
			remoteAddr: "192.168.1.1:12345",
			expected:   "1.2.3.4",
		},
		{
			name:       "strips port from RemoteAddr when used as fallback",
			headers:    map[string]string{},
			remoteAddr: "127.0.0.1",
			expected:   "127.0.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}

			result := GetClientIP(req)
			if result != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestValidateSessionWithContext_ReturnsSessionWhenValid(t *testing.T) {
	mockStore := &mocks.MockAuthStore{
		GetSessionFunc: func(sessionID string) (*store.Session, error) {
			return &store.Session{
				ID:        sessionID,
				UserID:    1,
				IPAddress: "10.0.0.1",
				UserAgent: "TestBrowser",
			}, nil
		},
	}

	session, err := ValidateSessionWithContext(mockStore, "test-session", "10.0.0.1", "TestBrowser")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if session == nil {
		t.Fatal("expected session, got nil")
	}
	if session.ID != "test-session" {
		t.Errorf("expected ID 'test-session', got %s", session.ID)
	}
}
