package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/mr-flannery/go-recipe-book/src/store"
	"github.com/mr-flannery/go-recipe-book/src/store/mocks"
)

func TestGenerateResetToken_ReturnsUniqueTokens(t *testing.T) {
	token1, _, err := GenerateResetToken()
	if err != nil {
		t.Fatalf("failed to generate first token: %v", err)
	}

	token2, _, err := GenerateResetToken()
	if err != nil {
		t.Fatalf("failed to generate second token: %v", err)
	}

	if token1 == token2 {
		t.Error("expected unique tokens, got identical tokens")
	}
}

func TestGenerateResetToken_ReturnsCorrectLength(t *testing.T) {
	plainToken, hashedToken, err := GenerateResetToken()
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	expectedPlainLength := resetTokenLength * 2
	if len(plainToken) != expectedPlainLength {
		t.Errorf("expected plain token length %d, got %d", expectedPlainLength, len(plainToken))
	}

	expectedHashLength := 64
	if len(hashedToken) != expectedHashLength {
		t.Errorf("expected hashed token length %d, got %d", expectedHashLength, len(hashedToken))
	}
}

func TestHashResetToken_ProducesConsistentHash(t *testing.T) {
	token := "test-token-12345"

	hash1 := HashResetToken(token)
	hash2 := HashResetToken(token)

	if hash1 != hash2 {
		t.Error("expected consistent hash for same input")
	}
}

func TestHashResetToken_ProducesDifferentHashesForDifferentInputs(t *testing.T) {
	hash1 := HashResetToken("token1")
	hash2 := HashResetToken("token2")

	if hash1 == hash2 {
		t.Error("expected different hashes for different inputs")
	}
}

func TestCreatePasswordResetToken_StoresTokenAndReturnsPlaintext(t *testing.T) {
	var storedUserID int
	var storedTokenHash string
	var storedExpiresAt time.Time

	mockStore := &mocks.MockAuthStore{
		CreatePasswordResetTokenFunc: func(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) error {
			storedUserID = userID
			storedTokenHash = tokenHash
			storedExpiresAt = expiresAt
			return nil
		},
	}

	plainToken, err := CreatePasswordResetToken(context.Background(), mockStore, 42)
	if err != nil {
		t.Fatalf("failed to create password reset token: %v", err)
	}

	if plainToken == "" {
		t.Error("expected non-empty plain token")
	}

	if storedUserID != 42 {
		t.Errorf("expected user ID 42, got %d", storedUserID)
	}

	if storedTokenHash == "" {
		t.Error("expected non-empty stored token hash")
	}

	expectedHash := HashResetToken(plainToken)
	if storedTokenHash != expectedHash {
		t.Error("stored hash doesn't match hash of plain token")
	}

	if storedExpiresAt.Before(time.Now().Add(23 * time.Hour)) {
		t.Error("expected expiration to be approximately 24 hours from now")
	}
}

func TestCreatePasswordResetToken_ReturnsErrorOnStoreFailure(t *testing.T) {
	mockStore := &mocks.MockAuthStore{
		CreatePasswordResetTokenFunc: func(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) error {
			return errors.New("database error")
		},
	}

	_, err := CreatePasswordResetToken(context.Background(), mockStore, 42)
	if err == nil {
		t.Error("expected error when store fails")
	}
}

func TestValidateResetToken_ReturnsUserIDWhenValid(t *testing.T) {
	plainToken := "test-plain-token"
	hashedToken := HashResetToken(plainToken)

	mockStore := &mocks.MockAuthStore{
		GetPasswordResetTokenFunc: func(ctx context.Context, tokenHash string) (*store.PasswordResetToken, error) {
			if tokenHash == hashedToken {
				return &store.PasswordResetToken{
					ID:        1,
					UserID:    42,
					ExpiresAt: time.Now().Add(1 * time.Hour),
					UsedAt:    nil,
				}, nil
			}
			return nil, errors.New("not found")
		},
	}

	userID, err := ValidateResetToken(context.Background(), mockStore, plainToken)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if userID != 42 {
		t.Errorf("expected user ID 42, got %d", userID)
	}
}

func TestValidateResetToken_ReturnsErrorWhenTokenNotFound(t *testing.T) {
	mockStore := &mocks.MockAuthStore{
		GetPasswordResetTokenFunc: func(ctx context.Context, tokenHash string) (*store.PasswordResetToken, error) {
			return nil, errors.New("not found")
		},
	}

	_, err := ValidateResetToken(context.Background(), mockStore, "invalid-token")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestValidateResetToken_ReturnsErrorWhenTokenExpired(t *testing.T) {
	plainToken := "expired-token"
	hashedToken := HashResetToken(plainToken)

	mockStore := &mocks.MockAuthStore{
		GetPasswordResetTokenFunc: func(ctx context.Context, tokenHash string) (*store.PasswordResetToken, error) {
			if tokenHash == hashedToken {
				return &store.PasswordResetToken{
					ID:        1,
					UserID:    42,
					ExpiresAt: time.Now().Add(-1 * time.Hour),
					UsedAt:    nil,
				}, nil
			}
			return nil, errors.New("not found")
		},
	}

	_, err := ValidateResetToken(context.Background(), mockStore, plainToken)
	if err == nil {
		t.Error("expected error for expired token")
	}
}

func TestValidateResetToken_ReturnsErrorWhenTokenAlreadyUsed(t *testing.T) {
	plainToken := "used-token"
	hashedToken := HashResetToken(plainToken)
	usedAt := time.Now().Add(-30 * time.Minute)

	mockStore := &mocks.MockAuthStore{
		GetPasswordResetTokenFunc: func(ctx context.Context, tokenHash string) (*store.PasswordResetToken, error) {
			if tokenHash == hashedToken {
				return &store.PasswordResetToken{
					ID:        1,
					UserID:    42,
					ExpiresAt: time.Now().Add(1 * time.Hour),
					UsedAt:    &usedAt,
				}, nil
			}
			return nil, errors.New("not found")
		},
	}

	_, err := ValidateResetToken(context.Background(), mockStore, plainToken)
	if err == nil {
		t.Error("expected error for already used token")
	}
}

func TestValidateAndUseResetToken_MarksTokenAsUsed(t *testing.T) {
	plainToken := "test-token"
	hashedToken := HashResetToken(plainToken)
	tokenMarkedUsed := false

	mockStore := &mocks.MockAuthStore{
		GetPasswordResetTokenFunc: func(ctx context.Context, tokenHash string) (*store.PasswordResetToken, error) {
			if tokenHash == hashedToken {
				return &store.PasswordResetToken{
					ID:        1,
					UserID:    42,
					ExpiresAt: time.Now().Add(1 * time.Hour),
					UsedAt:    nil,
				}, nil
			}
			return nil, errors.New("not found")
		},
		MarkPasswordResetTokenUsedFunc: func(ctx context.Context, tokenHash string) error {
			if tokenHash == hashedToken {
				tokenMarkedUsed = true
				return nil
			}
			return errors.New("not found")
		},
	}

	userID, err := ValidateAndUseResetToken(context.Background(), mockStore, plainToken)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if userID != 42 {
		t.Errorf("expected user ID 42, got %d", userID)
	}

	if !tokenMarkedUsed {
		t.Error("expected token to be marked as used")
	}
}

func TestResetPassword_UpdatesPasswordAndInvalidatesSessions(t *testing.T) {
	passwordUpdated := false
	sessionsInvalidated := false
	var updatedPasswordHash string

	mockStore := &mocks.MockAuthStore{
		UpdateUserPasswordFunc: func(ctx context.Context, userID int, passwordHash string) error {
			if userID == 42 {
				passwordUpdated = true
				updatedPasswordHash = passwordHash
				return nil
			}
			return errors.New("user not found")
		},
		DeleteUserSessionsFunc: func(ctx context.Context, userID int) error {
			if userID == 42 {
				sessionsInvalidated = true
				return nil
			}
			return errors.New("user not found")
		},
	}

	err := ResetPassword(context.Background(), mockStore, 42, "NewSecurePass123!")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if !passwordUpdated {
		t.Error("expected password to be updated")
	}

	if !sessionsInvalidated {
		t.Error("expected all sessions to be invalidated")
	}

	if updatedPasswordHash == "" {
		t.Error("expected password hash to be stored")
	}

	if updatedPasswordHash == "NewSecurePass123!" {
		t.Error("password should be hashed, not stored in plain text")
	}
}

func TestResetPassword_ReturnsErrorWhenPasswordInvalid(t *testing.T) {
	mockStore := &mocks.MockAuthStore{}

	err := ResetPassword(context.Background(), mockStore, 42, "weak")
	if err == nil {
		t.Error("expected error for weak password")
	}
}

func TestResetPassword_ReturnsErrorWhenUpdateFails(t *testing.T) {
	mockStore := &mocks.MockAuthStore{
		UpdateUserPasswordFunc: func(ctx context.Context, userID int, passwordHash string) error {
			return errors.New("database error")
		},
	}

	err := ResetPassword(context.Background(), mockStore, 42, "NewSecurePass123!")
	if err == nil {
		t.Error("expected error when password update fails")
	}
}

func TestResetPassword_ReturnsErrorWhenSessionInvalidationFails(t *testing.T) {
	mockStore := &mocks.MockAuthStore{
		UpdateUserPasswordFunc: func(ctx context.Context, userID int, passwordHash string) error {
			return nil
		},
		DeleteUserSessionsFunc: func(ctx context.Context, userID int) error {
			return errors.New("database error")
		},
	}

	err := ResetPassword(context.Background(), mockStore, 42, "NewSecurePass123!")
	if err == nil {
		t.Error("expected error when session invalidation fails")
	}
}

func TestGetResetURL_ContainsToken(t *testing.T) {
	token := "test-token-abc123"
	url := GetResetURL(token)

	if url == "" {
		t.Error("expected non-empty URL")
	}

	expectedSuffix := "reset-password?token=" + token
	if len(url) < len(expectedSuffix) {
		t.Errorf("URL too short: %s", url)
	}

	if url[len(url)-len(expectedSuffix):] != expectedSuffix {
		t.Errorf("URL doesn't contain expected token parameter: %s", url)
	}
}
