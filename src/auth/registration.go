package auth

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"
)

// RegistrationRequest represents a user registration request
type RegistrationRequest struct {
	ID           int
	Username     string
	Email        string
	PasswordHash string
	RequestedAt  time.Time
	Status       string // pending, approved, rejected
	ReviewedBy   *int
	ReviewedAt   *time.Time
	Notes        string
}

// CreateRegistrationRequest creates a new registration request
func CreateRegistrationRequest(db *sql.DB, username, email, password string) error {
	// Validate password strength
	if err := ValidatePasswordStrength(password); err != nil {
		return fmt.Errorf("password validation failed: %w", err)
	}

	// Hash the password
	passwordHash, err := HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Check if username or email already exists in users table
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM users WHERE username = $1 OR email = $2", username, email).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing users: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("username or email already exists")
	}

	// Check if there's already a pending request for this username/email
	err = db.QueryRow("SELECT COUNT(*) FROM registration_requests WHERE (username = $1 OR email = $2) AND status = 'pending'", username, email).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing requests: %w", err)
	}
	if count > 0 {
		return fmt.Errorf("registration request already pending for this username or email")
	}

	// Insert registration request
	query := `
		INSERT INTO registration_requests (username, email, password_hash, status)
		VALUES ($1, $2, $3, 'pending')`

	_, err = db.Exec(query, username, email, passwordHash)
	if err != nil {
		return fmt.Errorf("failed to create registration request: %w", err)
	}

	return nil
}

// GetPendingRegistrations returns all pending registration requests
func GetPendingRegistrations(db *sql.DB) ([]RegistrationRequest, error) {
	query := `
		SELECT id, username, email, password_hash, requested_at, status, reviewed_by, reviewed_at, notes
		FROM registration_requests 
		WHERE status = 'pending'
		ORDER BY requested_at ASC`

	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending registrations: %w", err)
	}
	defer rows.Close()

	var requests []RegistrationRequest
	for rows.Next() {
		var req RegistrationRequest
		err := rows.Scan(&req.ID, &req.Username, &req.Email, &req.PasswordHash,
			&req.RequestedAt, &req.Status, &req.ReviewedBy, &req.ReviewedAt, &req.Notes)
		if err != nil {
			return nil, fmt.Errorf("failed to scan registration request: %w", err)
		}
		requests = append(requests, req)
	}

	return requests, nil
}

// ApproveRegistration approves a registration request and creates the user account
func ApproveRegistration(db *sql.DB, requestID int, adminID int) error {
	// Start transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Get the registration request
	var req RegistrationRequest
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

	// Create the user account
	userQuery := `
		INSERT INTO users (username, email, password_hash, is_admin, is_active)
		VALUES ($1, $2, $3, false, true)`

	_, err = tx.Exec(userQuery, req.Username, req.Email, req.PasswordHash)
	if err != nil {
		return fmt.Errorf("failed to create user account: %w", err)
	}

	// Update registration request status
	updateQuery := `
		UPDATE registration_requests 
		SET status = 'approved', reviewed_by = $1, reviewed_at = NOW()
		WHERE id = $2`

	_, err = tx.Exec(updateQuery, adminID, requestID)
	if err != nil {
		return fmt.Errorf("failed to update registration request: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// RejectRegistration rejects a registration request
func RejectRegistration(db *sql.DB, requestID int, adminID int, reason string) error {
	query := `
		UPDATE registration_requests 
		SET status = 'rejected', reviewed_by = $1, reviewed_at = NOW(), notes = $2
		WHERE id = $3 AND status = 'pending'`

	result, err := db.Exec(query, adminID, reason, requestID)
	if err != nil {
		return fmt.Errorf("failed to reject registration request: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("registration request not found or already processed")
	}

	return nil
}

// GetRegistrationRequestByID returns a specific registration request
func GetRegistrationRequestByID(db *sql.DB, requestID int) (*RegistrationRequest, error) {
	var req RegistrationRequest
	query := `
		SELECT id, username, email, password_hash, requested_at, status, reviewed_by, reviewed_at, notes
		FROM registration_requests 
		WHERE id = $1`

	err := db.QueryRow(query, requestID).Scan(&req.ID, &req.Username, &req.Email, &req.PasswordHash,
		&req.RequestedAt, &req.Status, &req.ReviewedBy, &req.ReviewedAt, &req.Notes)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("registration request not found")
		}
		return nil, fmt.Errorf("failed to get registration request: %w", err)
	}

	return &req, nil
}

// CreateSeedAdmin creates the initial admin account if it doesn't exist
func CreateSeedAdmin(db *sql.DB, username, email, password string) error {
	// Check if admin already exists
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = $1", username).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing admin: %w", err)
	}
	if count > 0 {
		slog.Info("Seed admin already exists, skipping creation\n", "username", username, "email", email)
		return nil
	}

	// Hash the password
	passwordHash, err := HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

	// Create admin account
	query := `
		INSERT INTO users (username, email, password_hash, is_admin, is_active)
		VALUES ($1, $2, $3, true, true)`

	_, err = db.Exec(query, username, email, passwordHash)
	if err != nil {
		return fmt.Errorf("failed to create seed admin: %w", err)
	}

	fmt.Printf("Created seed admin account: %s\n", username)
	return nil
}
