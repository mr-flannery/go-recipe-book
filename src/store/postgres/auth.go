package postgres

import (
	"context"
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

func (s *AuthStore) GetUserByEmail(ctx context.Context, email string) (*store.AuthUser, string, error) {
	var user store.AuthUser
	var passwordHash string

	query := `
		SELECT id, username, email, password_hash, is_admin, is_active
		FROM users 
		WHERE email = $1 AND is_active = true`

	err := s.db.QueryRowContext(ctx, query, email).Scan(
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

func (s *AuthStore) UpdateLastLogin(ctx context.Context, userID int) error {
	query := `UPDATE users SET last_login = NOW() WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, userID)
	return err
}

func (s *AuthStore) GetUserByID(ctx context.Context, userID int) (*store.AuthUser, error) {
	var user store.AuthUser
	query := `
		SELECT id, username, email, is_admin, is_active
		FROM users 
		WHERE id = $1 AND is_active = true`

	err := s.db.QueryRowContext(ctx, query, userID).Scan(
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

func (s *AuthStore) GetFullUserByID(ctx context.Context, userID int) (*store.FullAuthUser, error) {
	var user store.FullAuthUser
	query := `
		SELECT id, username, email, is_admin, is_active, created_at, last_login
		FROM users 
		WHERE id = $1 AND is_active = true`

	err := s.db.QueryRowContext(ctx, query, userID).Scan(
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

func (s *AuthStore) GetUserIDByUsername(ctx context.Context, username string) (int, error) {
	var userID int
	err := s.db.QueryRowContext(ctx, "SELECT id FROM users WHERE username = $1", username).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("user not found")
		}
		return 0, fmt.Errorf("failed to fetch user ID: %w", err)
	}
	return userID, nil
}

func (s *AuthStore) CreateSession(ctx context.Context, session *store.Session) error {
	query := `
		INSERT INTO sessions (id, user_id, created_at, expires_at, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6)`

	now := time.Now()
	expiresAt := now.Add(sessionDuration)

	_, err := s.db.ExecContext(ctx, query, session.ID, session.UserID, now, expiresAt, session.IPAddress, session.UserAgent)
	if err != nil {
		return fmt.Errorf("failed to store session: %w", err)
	}

	return nil
}

func (s *AuthStore) GetSession(ctx context.Context, sessionID string) (*store.Session, error) {
	if sessionID == "" {
		return nil, fmt.Errorf("empty session ID")
	}

	var session store.Session
	query := `
		SELECT id, user_id, ip_address, user_agent
		FROM sessions 
		WHERE id = $1 AND expires_at > NOW()`

	err := s.db.QueryRowContext(ctx, query, sessionID).Scan(
		&session.ID, &session.UserID, &session.IPAddress, &session.UserAgent)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("session not found or expired")
		}
		return nil, fmt.Errorf("failed to validate session: %w", err)
	}

	return &session, nil
}

func (s *AuthStore) DeleteSession(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return nil
	}

	query := `DELETE FROM sessions WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to invalidate session: %w", err)
	}

	return nil
}

func (s *AuthStore) DeleteExpiredSessions(ctx context.Context) (int64, error) {
	query := `DELETE FROM sessions WHERE expires_at <= NOW()`
	result, err := s.db.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired sessions: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected, nil
}

func (s *AuthStore) DeleteUserSessions(ctx context.Context, userID int) error {
	query := `DELETE FROM sessions WHERE user_id = $1`
	_, err := s.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to invalidate user sessions: %w", err)
	}

	return nil
}

func (s *AuthStore) GetActiveSessionCount(ctx context.Context, userID int) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM sessions WHERE user_id = $1 AND expires_at > NOW()`
	err := s.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get session count: %w", err)
	}

	return count, nil
}

func (s *AuthStore) ExtendSession(ctx context.Context, sessionID string) error {
	newExpiresAt := time.Now().Add(sessionDuration)

	query := `UPDATE sessions SET expires_at = $1 WHERE id = $2 AND expires_at > NOW()`
	result, err := s.db.ExecContext(ctx, query, newExpiresAt, sessionID)
	if err != nil {
		return fmt.Errorf("failed to extend session: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("session not found or already expired")
	}

	return nil
}

func (s *AuthStore) CreateRegistrationRequest(ctx context.Context, username, email, passwordHash string) error {
	var count int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE username = $1 OR email = $2", username, email).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing users: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("username or email already exists")
	}

	err = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM registration_requests WHERE (username = $1 OR email = $2) AND status = 'pending'", username, email).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing requests: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("registration request already pending for this username or email")
	}

	query := `
		INSERT INTO registration_requests (username, email, password_hash, status)
		VALUES ($1, $2, $3, 'pending')`

	_, err = s.db.ExecContext(ctx, query, username, email, passwordHash)
	if err != nil {
		return fmt.Errorf("failed to create registration request: %w", err)
	}

	return nil
}

func (s *AuthStore) GetPendingRegistrations(ctx context.Context) ([]store.RegistrationRequest, error) {
	query := `
		SELECT id, username, email, password_hash, requested_at, status
		FROM registration_requests 
		WHERE status = 'pending'
		ORDER BY requested_at ASC`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending registrations: %w", err)
	}
	defer rows.Close()

	var requests []store.RegistrationRequest
	for rows.Next() {
		var req store.RegistrationRequest
		err := rows.Scan(&req.ID, &req.Username, &req.Email, &req.PasswordHash, &req.RequestedAt, &req.Status)
		if err != nil {
			return nil, fmt.Errorf("failed to scan registration request: %w", err)
		}
		requests = append(requests, req)
	}

	return requests, nil
}

func (s *AuthStore) GetAllRegistrations(ctx context.Context) ([]store.RegistrationRequest, error) {
	query := `
		SELECT id, username, email, password_hash, requested_at, status
		FROM registration_requests 
		ORDER BY requested_at DESC`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query registrations: %w", err)
	}
	defer rows.Close()

	var requests []store.RegistrationRequest
	for rows.Next() {
		var req store.RegistrationRequest
		err := rows.Scan(&req.ID, &req.Username, &req.Email, &req.PasswordHash, &req.RequestedAt, &req.Status)
		if err != nil {
			return nil, fmt.Errorf("failed to scan registration request: %w", err)
		}
		requests = append(requests, req)
	}

	return requests, nil
}

func (s *AuthStore) GetAllRegistrationsPaginated(ctx context.Context, limit, offset int) ([]store.RegistrationRequest, error) {
	query := `
		SELECT id, username, email, password_hash, requested_at, status
		FROM registration_requests 
		ORDER BY requested_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query registrations: %w", err)
	}
	defer rows.Close()

	var requests []store.RegistrationRequest
	for rows.Next() {
		var req store.RegistrationRequest
		err := rows.Scan(&req.ID, &req.Username, &req.Email, &req.PasswordHash, &req.RequestedAt, &req.Status)
		if err != nil {
			return nil, fmt.Errorf("failed to scan registration request: %w", err)
		}
		requests = append(requests, req)
	}

	return requests, nil
}

func (s *AuthStore) CountAllRegistrations(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM registration_requests`
	var count int
	err := s.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count registrations: %w", err)
	}
	return count, nil
}

func (s *AuthStore) ApproveRegistration(ctx context.Context, requestID, adminID int) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	var req store.RegistrationRequest
	query := `
		SELECT id, username, email, password_hash, status
		FROM registration_requests 
		WHERE id = $1 AND status = 'pending'`

	err = tx.QueryRowContext(ctx, query, requestID).Scan(&req.ID, &req.Username, &req.Email, &req.PasswordHash, &req.Status)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("registration request not found or already processed")
		}
		return fmt.Errorf("failed to get registration request: %w", err)
	}

	userQuery := `
		INSERT INTO users (username, email, password_hash, is_admin, is_active)
		VALUES ($1, $2, $3, false, true)`

	_, err = tx.ExecContext(ctx, userQuery, req.Username, req.Email, req.PasswordHash)
	if err != nil {
		return fmt.Errorf("failed to create user account: %w", err)
	}

	updateQuery := `
		UPDATE registration_requests 
		SET status = 'approved', reviewed_by = $1, reviewed_at = NOW()
		WHERE id = $2`

	_, err = tx.ExecContext(ctx, updateQuery, adminID, requestID)
	if err != nil {
		return fmt.Errorf("failed to update registration request: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *AuthStore) RejectRegistration(ctx context.Context, requestID, adminID int) error {
	query := `
		UPDATE registration_requests 
		SET status = 'rejected', reviewed_by = $1, reviewed_at = NOW()
		WHERE id = $2 AND status = 'pending'`

	result, err := s.db.ExecContext(ctx, query, adminID, requestID)
	if err != nil {
		return fmt.Errorf("failed to reject registration request: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("registration request not found or already processed")
	}

	return nil
}

func (s *AuthStore) CreateUser(ctx context.Context, username, email, passwordHash string, isAdmin bool) error {
	query := `
		INSERT INTO users (username, email, password_hash, is_admin, is_active)
		VALUES ($1, $2, $3, $4, true)`

	_, err := s.db.ExecContext(ctx, query, username, email, passwordHash, isAdmin)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (s *AuthStore) UserExists(ctx context.Context, username string) (bool, error) {
	var count int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE username = $1", username).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check existing user: %w", err)
	}
	return count > 0, nil
}

func (s *AuthStore) GetAllUsers(ctx context.Context) ([]store.AuthUser, error) {
	query := `
		SELECT id, username, email, COALESCE(is_admin, false), COALESCE(is_active, true)
		FROM users
		ORDER BY username ASC`

	rows, err := s.db.QueryContext(ctx, query)
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

func (s *AuthStore) DeleteUser(ctx context.Context, userID int) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM sessions WHERE user_id = $1", userID)
	if err != nil {
		return fmt.Errorf("failed to delete user sessions: %w", err)
	}

	result, err := s.db.ExecContext(ctx, "DELETE FROM users WHERE id = $1", userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (s *AuthStore) CreatePasswordResetToken(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) error {
	query := `
		INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)`

	_, err := s.db.ExecContext(ctx, query, userID, tokenHash, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to create password reset token: %w", err)
	}

	return nil
}

func (s *AuthStore) GetPasswordResetToken(ctx context.Context, tokenHash string) (*store.PasswordResetToken, error) {
	var token store.PasswordResetToken
	query := `
		SELECT id, user_id, expires_at, used_at
		FROM password_reset_tokens
		WHERE token_hash = $1`

	err := s.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&token.ID, &token.UserID, &token.ExpiresAt, &token.UsedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("password reset token not found")
		}
		return nil, fmt.Errorf("failed to get password reset token: %w", err)
	}

	return &token, nil
}

func (s *AuthStore) MarkPasswordResetTokenUsed(ctx context.Context, tokenHash string) error {
	query := `UPDATE password_reset_tokens SET used_at = NOW() WHERE token_hash = $1`
	result, err := s.db.ExecContext(ctx, query, tokenHash)
	if err != nil {
		return fmt.Errorf("failed to mark password reset token as used: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("password reset token not found")
	}

	return nil
}

func (s *AuthStore) DeleteExpiredPasswordResetTokens(ctx context.Context) (int64, error) {
	query := `DELETE FROM password_reset_tokens WHERE expires_at <= NOW() OR used_at IS NOT NULL`
	result, err := s.db.ExecContext(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired password reset tokens: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	return rowsAffected, nil
}

func (s *AuthStore) UpdateUserPassword(ctx context.Context, userID int, passwordHash string) error {
	query := `UPDATE users SET password_hash = $1 WHERE id = $2`
	result, err := s.db.ExecContext(ctx, query, passwordHash, userID)
	if err != nil {
		return fmt.Errorf("failed to update user password: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (s *AuthStore) ResetPasswordWithToken(ctx context.Context, tokenHash string, newPasswordHash string) (int, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	var token store.PasswordResetToken
	query := `
		SELECT id, user_id, expires_at, used_at
		FROM password_reset_tokens
		WHERE token_hash = $1`

	err = tx.QueryRowContext(ctx, query, tokenHash).Scan(
		&token.ID, &token.UserID, &token.ExpiresAt, &token.UsedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("invalid or expired reset token")
		}
		return 0, fmt.Errorf("failed to get password reset token: %w", err)
	}

	if token.UsedAt != nil {
		return 0, fmt.Errorf("reset token has already been used")
	}

	if time.Now().After(token.ExpiresAt) {
		return 0, fmt.Errorf("reset token has expired")
	}

	_, err = tx.ExecContext(ctx, `UPDATE users SET password_hash = $1 WHERE id = $2`, newPasswordHash, token.UserID)
	if err != nil {
		return 0, fmt.Errorf("failed to update password: %w", err)
	}

	_, err = tx.ExecContext(ctx, `UPDATE password_reset_tokens SET used_at = NOW() WHERE id = $1`, token.ID)
	if err != nil {
		return 0, fmt.Errorf("failed to mark token as used: %w", err)
	}

	_, err = tx.ExecContext(ctx, `DELETE FROM sessions WHERE user_id = $1`, token.UserID)
	if err != nil {
		return 0, fmt.Errorf("failed to invalidate sessions: %w", err)
	}

	if err = tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return token.UserID, nil
}
