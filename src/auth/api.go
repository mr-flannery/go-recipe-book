package auth

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/mr-flannery/go-recipe-book/src/config"
)

// RequireAPIKey creates middleware to enforce API key authentication
func RequireAPIKey() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get the Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				slog.Warn("API request missing Authorization header", "path", r.URL.Path, "method", r.Method)
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			// Check if it's a Bearer token
			if !strings.HasPrefix(authHeader, "Bearer ") {
				slog.Warn("API request with invalid Authorization header format", "path", r.URL.Path, "method", r.Method)
				http.Error(w, "Authorization header must be in Bearer token format", http.StatusUnauthorized)
				return
			}

			// Extract the token
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == "" {
				slog.Warn("API request with empty Bearer token", "path", r.URL.Path, "method", r.Method)
				http.Error(w, "Bearer token cannot be empty", http.StatusUnauthorized)
				return
			}

			// Get configuration to check valid API keys
			cfg, err := config.GetConfig()
			if err != nil {
				slog.Error("Failed to load configuration for API key validation", "error", err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			// Validate the token against configured API keys
			validToken := false
			for _, validKey := range cfg.Api.Keys {
				if token == validKey {
					validToken = true
					break
				}
			}

			if !validToken {
				slog.Warn("API request with invalid API key", "path", r.URL.Path, "method", r.Method)
				http.Error(w, "Invalid API key", http.StatusUnauthorized)
				return
			}

			slog.Info("API request authenticated successfully", "path", r.URL.Path, "method", r.Method)
			next.ServeHTTP(w, r)
		})
	}
}

// GetAdminUserID returns the admin user ID from the database
func GetAdminUserID() (int, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return 0, fmt.Errorf("failed to load configuration: %w", err)
	}

	// Use the existing function to get user ID by username
	adminID, err := GetUserIDByUsername(cfg.DB.Admin.Username)
	if err != nil {
		return 0, fmt.Errorf("failed to get admin user ID: %w", err)
	}

	return adminID, nil
}
