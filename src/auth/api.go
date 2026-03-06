package auth

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/mr-flannery/go-recipe-book/src/config"
	"github.com/mr-flannery/go-recipe-book/src/store"
)

func RequireAPIKey() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				slog.Warn("API request missing Authorization header", "path", r.URL.Path, "method", r.Method)
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			if !strings.HasPrefix(authHeader, "Bearer ") {
				slog.Warn("API request with invalid Authorization header format", "path", r.URL.Path, "method", r.Method)
				http.Error(w, "Authorization header must be in Bearer token format", http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == "" {
				slog.Warn("API request with empty Bearer token", "path", r.URL.Path, "method", r.Method)
				http.Error(w, "Bearer token cannot be empty", http.StatusUnauthorized)
				return
			}

			cfg := config.GetConfig()

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

func GetAdminUserID(ctx context.Context, authStore store.AuthStore) (int, error) {
	cfg := config.GetConfig()

	adminID, err := authStore.GetUserIDByUsername(ctx, cfg.DB.Admin.Username)
	if err != nil {
		return 0, fmt.Errorf("failed to get admin user ID: %w", err)
	}

	return adminID, nil
}
