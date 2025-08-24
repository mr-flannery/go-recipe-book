package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/mr-flannery/go-recipe-book/src/config"
	"github.com/mr-flannery/go-recipe-book/src/db"
)

const (
	sessionIDLength = 32             // 32 bytes = 256 bits
	sessionDuration = 24 * time.Hour // 24 hours default
)

func cookieSettings() (string, bool) {
	config, err := config.GetConfig()
	if err != nil {
		slog.Error("Failed to load config")
		return "__Secure-session", true
	}

	if config.Environment.Mode == "development" {
		return "session", false
	} else {
		return "__Secure-session", true
	}
}

// Session represents a user session
type Session struct {
	ID        string
	UserID    int
	CreatedAt time.Time
	ExpiresAt time.Time
	IPAddress string
	UserAgent string
}

// CreateSession creates a new secure session for the user
func CreateSession(userID int, ipAddress, userAgent string) (*Session, error) {
	// Generate cryptographically secure session ID
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

	// Store session in database
	query := `
		INSERT INTO sessions (id, user_id, created_at, expires_at, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6)`

	// Get database connection
	database, err := db.GetConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}
	defer database.Close()

	_, err = database.Exec(query, session.ID, session.UserID, session.CreatedAt,
		session.ExpiresAt, session.IPAddress, session.UserAgent)
	if err != nil {
		return nil, fmt.Errorf("failed to store session: %w", err)
	}

	return session, nil
}

// ValidateSession validates a session ID and returns the session if valid
func ValidateSession(db *sql.DB, sessionID string) (*Session, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("empty session ID")
	}

	var session Session
	query := `
		SELECT id, user_id, created_at, expires_at, ip_address, user_agent
		FROM sessions 
		WHERE id = $1 AND expires_at > NOW()`

	err := db.QueryRow(query, sessionID).Scan(
		&session.ID, &session.UserID, &session.CreatedAt,
		&session.ExpiresAt, &session.IPAddress, &session.UserAgent)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found or expired")
		}
		return nil, fmt.Errorf("failed to validate session: %w", err)
	}

	return &session, nil
}

// InvalidateSession removes a session from the database
func InvalidateSession(sessionID string) error {
	// Get database connection
	database, err := db.GetConnection()
	if err != nil {
		return fmt.Errorf("failed to get database connection: %w", err)
	}
	defer database.Close()

	if sessionID == "" {
		return nil // Nothing to invalidate
	}

	query := `DELETE FROM sessions WHERE id = $1`
	_, err = database.Exec(query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to invalidate session: %w", err)
	}

	return nil
}

// InvalidateAllUserSessions removes all sessions for a specific user
func InvalidateAllUserSessions(db *sql.DB, userID int) error {
	query := `DELETE FROM sessions WHERE user_id = $1`
	_, err := db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to invalidate user sessions: %w", err)
	}

	return nil
}

// CleanupExpiredSessions removes expired sessions from the database
func CleanupExpiredSessions(db *sql.DB) error {
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

// SetSecureSessionCookie sets a secure session cookie
func SetSecureSessionCookie(w http.ResponseWriter, sessionID string) {
	cookieName, secure := cookieSettings()
	cookie := &http.Cookie{
		Name:     cookieName,
		Value:    sessionID,
		Path:     "/",
		MaxAge:   int(sessionDuration.Seconds()),
		HttpOnly: true,
		Secure:   secure, // Set to true in production with HTTPS
		SameSite: http.SameSiteStrictMode,
	}

	http.SetCookie(w, cookie)
}

// ClearSessionCookie clears the session cookie
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

// GetSessionFromRequest extracts session ID from request cookie
func GetSessionFromRequest(r *http.Request) (string, error) {
	cookieName, _ := cookieSettings()
	cookie, err := r.Cookie(cookieName)
	if err != nil {
		return "", fmt.Errorf("session cookie not found: %w", err)
	}

	return cookie.Value, nil
}

// GetClientIP extracts the real client IP address from the request
func GetClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first (for proxies/load balancers)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP in the chain
		if ip := net.ParseIP(xff); ip != nil {
			return ip.String()
		}
	}

	// Check X-Real-IP header
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		if ip := net.ParseIP(xri); ip != nil {
			return ip.String()
		}
	}

	// Fall back to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}

// ValidateSessionWithContext validates session and checks IP/User-Agent for additional security
func ValidateSessionWithContext(db *sql.DB, sessionID, currentIP, currentUserAgent string) (*Session, error) {
	session, err := ValidateSession(db, sessionID)
	if err != nil {
		return nil, err
	}

	// Optional: Check if IP address matches (can be disabled for mobile users)
	// This is commented out as it can cause issues with mobile networks
	// if session.IPAddress != currentIP {
	//     return nil, fmt.Errorf("session IP mismatch")
	// }

	// Optional: Check if User-Agent matches (basic session hijacking protection)
	// This is also commented out as it can cause issues with browser updates
	// if session.UserAgent != currentUserAgent {
	//     return nil, fmt.Errorf("session user agent mismatch")
	// }

	return session, nil
}

// ExtendSession extends the expiration time of a session
func ExtendSession(db *sql.DB, sessionID string) error {
	newExpiresAt := time.Now().Add(sessionDuration)

	query := `UPDATE sessions SET expires_at = $1 WHERE id = $2 AND expires_at > NOW()`
	result, err := db.Exec(query, newExpiresAt, sessionID)
	if err != nil {
		return fmt.Errorf("failed to extend session: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("session not found or already expired")
	}

	return nil
}

// generateSecureSessionID generates a cryptographically secure random session ID
func generateSecureSessionID() (string, error) {
	bytes := make([]byte, sessionIDLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return hex.EncodeToString(bytes), nil
}

// GetActiveSessionCount returns the number of active sessions for a user
func GetActiveSessionCount(db *sql.DB, userID int) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM sessions WHERE user_id = $1 AND expires_at > NOW()`
	err := db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get session count: %w", err)
	}

	return count, nil
}
