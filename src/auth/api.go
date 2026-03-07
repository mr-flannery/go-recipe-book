package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/mr-flannery/go-recipe-book/src/config"
	"github.com/mr-flannery/go-recipe-book/src/logging"
	"github.com/mr-flannery/go-recipe-book/src/store"
)

func RequireAPIKey() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				logging.AddError(ctx, errors.New("missing authorization header"), "API authentication failed")
				http.Error(w, "Authorization header required", http.StatusUnauthorized)
				return
			}

			if !strings.HasPrefix(authHeader, "Bearer ") {
				logging.AddError(ctx, errors.New("invalid authorization header format"), "API authentication failed")
				http.Error(w, "Authorization header must be in Bearer token format", http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == "" {
				logging.AddError(ctx, errors.New("empty bearer token"), "API authentication failed")
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
				logging.AddError(ctx, errors.New("invalid API key"), "API authentication failed")
				http.Error(w, "Invalid API key", http.StatusUnauthorized)
				return
			}

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
