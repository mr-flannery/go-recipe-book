package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/mr-flannery/go-recipe-book/src/store"
)

const sessionDuration = 24 * time.Hour

type AuthStore struct {
	db *sql.DB
}

func NewAuthStore(db *sql.DB) *AuthStore {
	return &AuthStore{db: db}
}

func (s *AuthStore) GetUserByEmail(email string) (*store.AuthUser, string, error) {
	var user store.AuthUser
	var passwordHash string

	query := `
		SELECT id, username, email, password_hash, is_admin, is_active
		FROM users 
		WHERE email = $1 AND is_active = true`

	err := s.db.QueryRow(query, email).Scan(
		&user.ID, &user.Username, &user.Email, &passwordHash,
		&user.IsAdmin, &user.IsActive)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, "", fmt.Errorf("user not found")
		}
		return nil, "", fmt.Errorf("query error: %w", err)
	}

	return &user, passwordHash, nil
}

func (s *AuthStore) UpdateLastLogin(userID int) error {
	query := `UPDATE users SET last_login = NOW() WHERE id = $1`
	_, err := s.db.Exec(query, userID)
	return err
}

func (s *AuthStore) GetUserByID(userID int) (*store.AuthUser, error) {
	var user store.AuthUser
	query := `
		SELECT id, username, email, is_admin, is_active
		FROM users 
		WHERE id = $1 AND is_active = true`

	err := s.db.QueryRow(query, userID).Scan(
		&user.ID, &user.Username, &user.Email,
		&user.IsAdmin, &user.IsActive)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (s *AuthStore) GetFullUserByID(userID int) (*store.FullAuthUser, error) {
	var user store.FullAuthUser
	query := `
		SELECT id, username, email, is_admin, is_active, created_at, last_login
		FROM users 
		WHERE id = $1 AND is_active = true`

	err := s.db.QueryRow(query, userID).Scan(
		&user.ID, &user.Username, &user.Email,
		&user.IsAdmin, &user.IsActive, &user.CreatedAt, &user.LastLogin)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}

func (s *AuthStore) GetUserIDByUsername(username string) (int, error) {
	var userID int
	err := s.db.QueryRow("SELECT id FROM users WHERE username = $1", username).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("user not found")
		}
		return 0, fmt.Errorf("failed to fetch user ID: %w", err)
	}
	return userID, nil
}

func (s *AuthStore) CreateSession(session *store.Session) error {
	query := `
		INSERT INTO sessions (id, user_id, created_at, expires_at, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6)`

	now := time.Now()
	expiresAt := now.Add(sessionDuration)

	_, err := s.db.Exec(query, session.ID, session.UserID, now, expiresAt, session.IPAddress, session.UserAgent)
	if err != nil {
		return fmt.Errorf("failed to store session: %w", err)
	}

	return nil
}

func (s *AuthStore) GetSession(sessionID string) (*store.Session, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("empty session ID")
	}

	var session store.Session
	query := `
		SELECT id, user_id, ip_address, user_agent
		FROM sessions 
		WHERE id = $1 AND expires_at > NOW()`

	err := s.db.QueryRow(query, sessionID).Scan(
		&session.ID, &session.UserID, &session.IPAddress, &session.UserAgent)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found or expired")
		}
		return nil, fmt.Errorf("failed to validate session: %w", err)
	}

	return &session, nil
}

func (s *AuthStore) DeleteSession(sessionID string) error {
	if sessionID == "" {
		return nil
	}

	query := `DELETE FROM sessions WHERE id = $1`
	_, err := s.db.Exec(query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to invalidate session: %w", err)
	}

	return nil
}

func (s *AuthStore) DeleteExpiredSessions() (int64, error) {
	query := `DELETE FROM sessions WHERE expires_at <= NOW()`
	result, err := s.db.Exec(query)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected, nil
}

func (s *AuthStore) DeleteUserSessions(userID int) error {
	query := `DELETE FROM sessions WHERE user_id = $1`
	_, err := s.db.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("failed to invalidate user sessions: %w", err)
	}

	return nil
}

func (s *AuthStore) GetActiveSessionCount(userID int) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM sessions WHERE user_id = $1 AND expires_at > NOW()`
	err := s.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get session count: %w", err)
	}

	return count, nil
}

func (s *AuthStore) ExtendSession(sessionID string) error {
	newExpiresAt := time.Now().Add(sessionDuration)

	query := `UPDATE sessions SET expires_at = $1 WHERE id = $2 AND expires_at > NOW()`
	result, err := s.db.Exec(query, newExpiresAt, sessionID)
	if err != nil {
		return fmt.Errorf("failed to extend session: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("session not found or already expired")
	}

	return nil
}

func (s *AuthStore) CreateRegistrationRequest(username, email, passwordHash string) error {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users WHERE username = $1 OR email = $2", username, email).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing users: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("username or email already exists")
	}

	err = s.db.QueryRow("SELECT COUNT(*) FROM registration_requests WHERE (username = $1 OR email = $2) AND status = 'pending'", username, email).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing requests: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("registration request already pending for this username or email")
	}

	query := `
		INSERT INTO registration_requests (username, email, password_hash, status)
		VALUES ($1, $2, $3, 'pending')`

	_, err = s.db.Exec(query, username, email, passwordHash)
	if err != nil {
		return fmt.Errorf("failed to create registration request: %w", err)
	}

	return nil
}

func (s *AuthStore) GetPendingRegistrations() ([]store.RegistrationRequest, error) {
	query := `
		SELECT id, username, email, password_hash, status
		FROM registration_requests 
		WHERE status = 'pending'
		ORDER BY requested_at ASC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending registrations: %w", err)
	}
	defer rows.Close()

	var requests []store.RegistrationRequest
	for rows.Next() {
		var req store.RegistrationRequest
		err := rows.Scan(&req.ID, &req.Username, &req.Email, &req.PasswordHash, &req.Status)
		if err != nil {
			return nil, fmt.Errorf("failed to scan registration request: %w", err)
		}
		requests = append(requests, req)
	}

	return requests, nil
}

func (s *AuthStore) ApproveRegistration(requestID, adminID int) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	var req store.RegistrationRequest
	query := `
		SELECT id, username, email, password_hash, status
		FROM registration_requests 
		WHERE id = $1 AND status = 'pending'`

	err = tx.QueryRow(query, requestID).Scan(&req.ID, &req.Username, &req.Email, &req.PasswordHash, &req.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("registration request not found or already processed")
		}
		return fmt.Errorf("failed to get registration request: %w", err)
	}

	userQuery := `
		INSERT INTO users (username, email, password_hash, is_admin, is_active)
		VALUES ($1, $2, $3, false, true)`

	_, err = tx.Exec(userQuery, req.Username, req.Email, req.PasswordHash)
	if err != nil {
		return fmt.Errorf("failed to create user account: %w", err)
	}

	updateQuery := `
		UPDATE registration_requests 
		SET status = 'approved', reviewed_by = $1, reviewed_at = NOW()
		WHERE id = $2`

	_, err = tx.Exec(updateQuery, adminID, requestID)
	if err != nil {
		return fmt.Errorf("failed to update registration request: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *AuthStore) RejectRegistration(requestID, adminID int) error {
	query := `
		UPDATE registration_requests 
		SET status = 'rejected', reviewed_by = $1, reviewed_at = NOW()
		WHERE id = $2 AND status = 'pending'`

	result, err := s.db.Exec(query, adminID, requestID)
	if err != nil {
		return fmt.Errorf("failed to reject registration request: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("registration request not found or already processed")
	}

	return nil
}

func (s *AuthStore) CreateUser(username, email, passwordHash string, isAdmin bool) error {
	query := `
		INSERT INTO users (username, email, password_hash, is_admin, is_active)
		VALUES ($1, $2, $3, $4, true)`

	_, err := s.db.Exec(query, username, email, passwordHash, isAdmin)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (s *AuthStore) UserExists(username string) (bool, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users WHERE username = $1", username).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check existing user: %w", err)
	}
	return count > 0, nil
}

func (s *AuthStore) GetAllUsers() ([]store.AuthUser, error) {
	query := `
		SELECT id, username, email, COALESCE(is_admin, false), COALESCE(is_active, true)
		FROM users
		ORDER BY username ASC`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []store.AuthUser
	for rows.Next() {
		var user store.AuthUser
		err := rows.Scan(&user.ID, &user.Username, &user.Email, &user.IsAdmin, &user.IsActive)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

func (s *AuthStore) DeleteUser(userID int) error {
	_, err := s.db.Exec("DELETE FROM sessions WHERE user_id = $1", userID)
	if err != nil {
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}

	result, err := s.db.Exec("DELETE FROM users WHERE id = $1", userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}
