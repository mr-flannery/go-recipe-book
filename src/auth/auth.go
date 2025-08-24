package auth

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/mr-flannery/go-recipe-book/src/db"
)

// User represents a user in the system
type User struct {
	ID        int
	Username  string
	Email     string
	IsAdmin   bool
	IsActive  bool
	LastLogin *time.Time
}

// Authenticate verifies email and password against the database
func Authenticate(email string, password string) (*User, error) {
	var user User
	var passwordHash string

	query := `
		SELECT id, username, email, password_hash, is_admin, is_active, last_login
		FROM users 
		WHERE email = $1 AND is_active = true`

	// Get database connection
	database, err := db.GetConnection()
	if err != nil {
		return &user, err
	}
	defer database.Close()

	err = database.QueryRow(query, email).Scan(
		&user.ID, &user.Username, &user.Email, &passwordHash,
		&user.IsAdmin, &user.IsActive, &user.LastLogin)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("invalid email or password")
		}
		return nil, fmt.Errorf("authentication error: %w", err)
	}

	// Verify password using Argon2id
	if err := VerifyPassword(password, passwordHash); err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Update last login time
	updateQuery := `UPDATE users SET last_login = NOW() WHERE id = $1`
	_, err = database.Exec(updateQuery, user.ID)
	if err != nil {
		// Log error but don't fail authentication
		fmt.Printf("Warning: failed to update last login for user %d: %v\n", user.ID, err)
	}

	return &user, nil
}

// GetUserBySession retrieves user information from a session
func GetUserBySession(db *sql.DB, r *http.Request) (*User, error) {
	sessionID, err := GetSessionFromRequest(r)
	if err != nil {
		return nil, fmt.Errorf("no valid session: %w", err)
	}

	session, err := ValidateSession(db, sessionID)
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

// IsSessionValid checks if the current request has a valid session
func IsSessionValid(db *sql.DB, r *http.Request) bool {
	_, err := GetUserBySession(db, r)
	return err == nil
}

// RequireAuth creates middleware to enforce authentication
func RequireAuth(db *sql.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !IsSessionValid(db, r) {
				// Include the current URL as redirect parameter
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

// GetUserIDByUsername fetches the user ID for a given username from the database
func GetUserIDByUsername(username string) (int, error) {
	// Replace with actual database connection
	db, err := sql.Open("postgres", "host=localhost port=5432 user=local-recipe-user password=local-recipe-password dbname=recipe-book sslmode=disable")
	if err != nil {
		return 0, fmt.Errorf("failed to connect to database: %v", err)
	}
	defer db.Close()

	var userID int
	err = db.QueryRow("SELECT id FROM users WHERE username = $1", username).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("user not found")
		}
		return 0, fmt.Errorf("failed to fetch user ID: %v", err)
	}

	return userID, nil
}
