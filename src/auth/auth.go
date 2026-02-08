package auth

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/mr-flannery/go-recipe-book/src/store"
)

type User struct {
	ID        int
	Username  string
	Email     string
	IsAdmin   bool
	IsActive  bool
	LastLogin *time.Time
}

func userFromAuthUser(au *store.AuthUser) *User {
	if au == nil {
		return nil
	}
	return &User{
		ID:       au.ID,
		Username: au.Username,
		Email:    au.Email,
		IsAdmin:  au.IsAdmin,
		IsActive: au.IsActive,
	}
}

func Authenticate(authStore store.AuthStore, email string, password string) (*User, error) {
	authUser, passwordHash, err := authStore.GetUserByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	if err := VerifyPassword(password, passwordHash); err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	if err := authStore.UpdateLastLogin(authUser.ID); err != nil {
		fmt.Printf("Warning: failed to update last login for user %d: %v\n", authUser.ID, err)
	}

	return userFromAuthUser(authUser), nil
}

func GetUserBySession(authStore store.AuthStore, r *http.Request) (*User, error) {
	sessionID, err := GetSessionFromRequest(r)
	if err != nil {
		return nil, fmt.Errorf("no valid session: %w", err)
	}

	session, err := authStore.GetSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session: %w", err)
	}

	authUser, err := authStore.GetUserByID(session.UserID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return userFromAuthUser(authUser), nil
}

func IsSessionValid(authStore store.AuthStore, r *http.Request) bool {
	_, err := GetUserBySession(authStore, r)
	return err == nil
}

func RequireAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetUserFromContext(r.Context())
			if user == nil {
				currentURL := r.URL.Path
				if r.URL.RawQuery != "" {
					currentURL += "?" + r.URL.RawQuery
				}
				redirectURL := "/login?redirect=" + currentURL
				http.Redirect(w, r, redirectURL, http.StatusSeeOther)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func GetUserIDByUsername(authStore store.AuthStore, username string) (int, error) {
	return authStore.GetUserIDByUsername(username)
}

// Legacy functions that accept *sql.DB for backward compatibility during transition
// These will be deprecated once all callers are updated

func GetUserBySessionLegacy(db *sql.DB, r *http.Request) (*User, error) {
	sessionID, err := GetSessionFromRequest(r)
	if err != nil {
		return nil, fmt.Errorf("no valid session: %w", err)
	}

	session, err := ValidateSessionLegacy(db, sessionID)
	if err != nil {
		return nil, fmt.Errorf("invalid session: %w", err)
	}

	var user User
	query := `
		SELECT id, username, email, is_admin, is_active, last_login
		FROM users 
		WHERE id = $1 AND is_active = true`

	err = db.QueryRow(query, session.UserID).Scan(
		&user.ID, &user.Username, &user.Email,
		&user.IsAdmin, &user.IsActive, &user.LastLogin)

	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return &user, nil
}

func ValidateSessionLegacy(db *sql.DB, sessionID string) (*Session, error) {
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
