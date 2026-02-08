package auth

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/mr-flannery/go-recipe-book/src/store"
)

type RegistrationRequest struct {
	ID           int
	Username     string
	Email        string
	PasswordHash string
	RequestedAt  time.Time
	Status       string
	ReviewedBy   *int
	ReviewedAt   *time.Time
}

func CreateRegistrationRequest(authStore store.AuthStore, username, email, password string) error {
	if err := ValidatePasswordStrength(password); err != nil {
		return fmt.Errorf("password validation failed: %w", err)
	}

	passwordHash, err := HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	return authStore.CreateRegistrationRequest(username, email, passwordHash)
}

func GetPendingRegistrations(authStore store.AuthStore) ([]RegistrationRequest, error) {
	storeReqs, err := authStore.GetPendingRegistrations()
	if err != nil {
		return nil, err
	}

	reqs := make([]RegistrationRequest, len(storeReqs))
	for i, r := range storeReqs {
		reqs[i] = RegistrationRequest{
			ID:           r.ID,
			Username:     r.Username,
			Email:        r.Email,
			PasswordHash: r.PasswordHash,
			Status:       r.Status,
		}
	}
	return reqs, nil
}

func ApproveRegistration(authStore store.AuthStore, requestID int, adminID int) error {
	return authStore.ApproveRegistration(requestID, adminID)
}

func RejectRegistration(authStore store.AuthStore, requestID int, adminID int, reason string) error {
	return authStore.RejectRegistration(requestID, adminID, reason)
}

func GetRegistrationRequestByID(authStore store.AuthStore, requestID int) (*RegistrationRequest, error) {
	reqs, err := authStore.GetPendingRegistrations()
	if err != nil {
		return nil, err
	}

	for _, r := range reqs {
		if r.ID == requestID {
			return &RegistrationRequest{
				ID:           r.ID,
				Username:     r.Username,
				Email:        r.Email,
				PasswordHash: r.PasswordHash,
				Status:       r.Status,
			}, nil
		}
	}

	return nil, fmt.Errorf("registration request not found")
}

func CreateSeedAdmin(authStore store.AuthStore, username, email, password string) error {
	exists, err := authStore.UserExists(username)
	if err != nil {
		return fmt.Errorf("failed to check existing admin: %w", err)
	}
	if exists {
		slog.Info("Seed admin already exists, skipping creation", "username", username, "email", email)
		return nil
	}

	passwordHash, err := HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

	err = authStore.CreateUser(username, email, passwordHash, true)
	if err != nil {
		return fmt.Errorf("failed to create seed admin: %w", err)
	}

	fmt.Printf("Created seed admin account: %s\n", username)
	return nil
}

// Legacy function for backward compatibility
func CreateSeedAdminLegacy(db *sql.DB, username, email, password string) error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE username = $1", username).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing admin: %w", err)
	}
	if count > 0 {
		slog.Info("Seed admin already exists, skipping creation", "username", username, "email", email)
		return nil
	}

	passwordHash, err := HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

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
