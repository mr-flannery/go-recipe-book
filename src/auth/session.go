package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/mr-flannery/go-recipe-book/src/config"
	"github.com/mr-flannery/go-recipe-book/src/store"
)

const (
	sessionIDLength = 32
	sessionDuration = 24 * time.Hour
)

func cookieSettings() (string, bool) {
	config := config.GetConfig()

	if config.Environment.Mode == "development" {
		return "session", false
	} else {
		return "__Secure-session", true
	}
}

type Session struct {
	ID        string
	UserID    int
	CreatedAt time.Time
	ExpiresAt time.Time
	IPAddress string
	UserAgent string
}

func CreateSession(authStore store.AuthStore, userID int, ipAddress, userAgent string) (*Session, error) {
	sessionID, err := generateSecureSessionID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate session ID: %w", err)
	}

	now := time.Now()
	expiresAt := now.Add(sessionDuration)

	session := &Session{
		ID:        sessionID,
		UserID:    userID,
		CreatedAt: now,
		ExpiresAt: expiresAt,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	storeSession := &store.Session{
		ID:        sessionID,
		UserID:    userID,
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}

	err = authStore.CreateSession(storeSession)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func ValidateSession(authStore store.AuthStore, sessionID string) (*Session, error) {
	storeSession, err := authStore.GetSession(sessionID)
	if err != nil {
		return nil, err
	}

	return &Session{
		ID:        storeSession.ID,
		UserID:    storeSession.UserID,
		IPAddress: storeSession.IPAddress,
		UserAgent: storeSession.UserAgent,
	}, nil
}

func InvalidateSession(authStore store.AuthStore, sessionID string) error {
	return authStore.DeleteSession(sessionID)
}

func InvalidateAllUserSessions(authStore store.AuthStore, userID int) error {
	return authStore.DeleteUserSessions(userID)
}

func CleanupExpiredSessions(authStore store.AuthStore) error {
	rowsAffected, err := authStore.DeleteExpiredSessions()
	if err != nil {
		return err
	}

	if rowsAffected > 0 {
		fmt.Printf("Cleaned up %d expired sessions\n", rowsAffected)
	}

	return nil
}

func SetSecureSessionCookie(w http.ResponseWriter, sessionID string) {
	cookieName, secure := cookieSettings()
	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    sessionID,
		Path:     "/",
		MaxAge:   int(sessionDuration.Seconds()),
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(w, cookie)
}

func ClearSessionCookie(w http.ResponseWriter) {
	cookieName, secure := cookieSettings()
	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC),
	}

	http.SetCookie(w, cookie)
}

func GetSessionFromRequest(r *http.Request) (string, error) {
	cookieName, _ := cookieSettings()
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return "", fmt.Errorf("session cookie not found: %w", err)
	}

	return cookie.Value, nil
}

func GetClientIP(r *http.Request) string {
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		if ip := net.ParseIP(xff); ip != nil {
			return ip.String()
		}
	}

	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		if ip := net.ParseIP(xri); ip != nil {
			return ip.String()
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}

func ValidateSessionWithContext(authStore store.AuthStore, sessionID, currentIP, currentUserAgent string) (*Session, error) {
	session, err := ValidateSession(authStore, sessionID)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func ExtendSession(authStore store.AuthStore, sessionID string) error {
	return authStore.ExtendSession(sessionID)
}

func generateSecureSessionID() (string, error) {
	bytes := make([]byte, sessionIDLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

func GetActiveSessionCount(authStore store.AuthStore, userID int) (int, error) {
	return authStore.GetActiveSessionCount(userID)
}

// Legacy functions for backward compatibility during transition

func CleanupExpiredSessionsLegacy(db *sql.DB) error {
	query := `DELETE FROM sessions WHERE expires_at <= NOW()`
	result, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		fmt.Printf("Cleaned up %d expired sessions\n", rowsAffected)
	}

	return nil
}
