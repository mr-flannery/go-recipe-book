package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/mr-flannery/go-recipe-book/src/store"
	"github.com/mr-flannery/go-recipe-book/src/utils"
)

const (
	resetTokenLength   = 32
	ResetTokenDuration = 24 * time.Hour
)

func GenerateResetToken() (string, string, error) {
	randomBytes := make([]byte, resetTokenLength)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	plainToken := hex.EncodeToString(randomBytes)
	hashedToken := HashResetToken(plainToken)

	return plainToken, hashedToken, nil
}

func HashResetToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func CreatePasswordResetToken(ctx context.Context, authStore store.AuthStore, userID int) (string, error) {
	plainToken, hashedToken, err := GenerateResetToken()
	if err != nil {
		return "", err
	}

	expiresAt := time.Now().Add(ResetTokenDuration)

	err = authStore.CreatePasswordResetToken(ctx, userID, hashedToken, expiresAt)
	if err != nil {
		return "", fmt.Errorf("failed to store password reset token: %w", err)
	}

	return plainToken, nil
}

func ValidateResetToken(ctx context.Context, authStore store.AuthStore, token string) (int, error) {
	hashedToken := HashResetToken(token)

	resetToken, err := authStore.GetPasswordResetToken(ctx, hashedToken)
	if err != nil {
		return 0, fmt.Errorf("invalid or expired reset token")
	}

	if resetToken.UsedAt != nil {
		return 0, fmt.Errorf("reset token has already been used")
	}

	if time.Now().After(resetToken.ExpiresAt) {
		return 0, fmt.Errorf("reset token has expired")
	}

	return resetToken.UserID, nil
}

func ValidateAndUseResetToken(ctx context.Context, authStore store.AuthStore, token string) (int, error) {
	hashedToken := HashResetToken(token)

	resetToken, err := authStore.GetPasswordResetToken(ctx, hashedToken)
	if err != nil {
		return 0, fmt.Errorf("invalid or expired reset token")
	}

	if resetToken.UsedAt != nil {
		return 0, fmt.Errorf("reset token has already been used")
	}

	if time.Now().After(resetToken.ExpiresAt) {
		return 0, fmt.Errorf("reset token has expired")
	}

	err = authStore.MarkPasswordResetTokenUsed(ctx, hashedToken)
	if err != nil {
		return 0, fmt.Errorf("failed to mark token as used: %w", err)
	}

	return resetToken.UserID, nil
}

func ResetPassword(ctx context.Context, authStore store.AuthStore, userID int, newPassword string) error {
	passwordHash, err := HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	err = authStore.UpdateUserPassword(ctx, userID, passwordHash)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	err = authStore.DeleteUserSessions(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to invalidate sessions: %w", err)
	}

	return nil
}

func ResetPasswordWithToken(ctx context.Context, authStore store.AuthStore, token string, newPassword string) (int, error) {
	passwordHash, err := HashPassword(newPassword)
	if err != nil {
		return 0, fmt.Errorf("failed to hash password: %w", err)
	}

	tokenHash := HashResetToken(token)
	userID, err := authStore.ResetPasswordWithToken(ctx, tokenHash, passwordHash)
	if err != nil {
		return 0, err
	}

	return userID, nil
}

func GetResetURL(token string) string {
	baseURL := utils.GetAppBaseURL()
	return fmt.Sprintf("%s/reset-password?token=%s", baseURL, token)
}

func CleanupExpiredPasswordResetTokens(ctx context.Context, authStore store.AuthStore) error {
	rowsAffected, err := authStore.DeleteExpiredPasswordResetTokens(ctx)
	if err != nil {
		return err
	}

	if rowsAffected > 0 {
		fmt.Printf("Cleaned up %d expired/used password reset tokens\n", rowsAffected)
	}

	return nil
}
