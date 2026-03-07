package auth

import (
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"net/http"
	"strings"

	"github.com/mr-flannery/go-recipe-book/src/config"
	"github.com/mr-flannery/go-recipe-book/src/logging"
	"github.com/mr-flannery/go-recipe-book/src/store"
)

func RequireAPIKey(apiKeyStore store.APIKeyStore, authStore store.AuthStore) func(http.Handler) http.Handler {
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

			apiKey, user, err := validateAPIKey(ctx, apiKeyStore, authStore, token)
			if err != nil {
				logging.AddError(ctx, err, "API authentication failed")
				http.Error(w, "Invalid API key", http.StatusUnauthorized)
				return
			}

			go func() {
				_ = apiKeyStore.UpdateLastUsed(context.Background(), apiKey.ID)
			}()

			userInfo := &UserInfo{
				IsLoggedIn: true,
				IsAdmin:    user.IsAdmin,
				Username:   user.Username,
				UserID:     user.ID,
			}
			ctx = ContextWithUserInfo(ctx, userInfo)

			logging.AddMany(ctx, map[string]any{
				"api_key.id":    apiKey.ID,
				"api_key.name":  apiKey.Name,
				"user.id":       user.ID,
				"user.username": user.Username,
			})

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func validateAPIKey(ctx context.Context, apiKeyStore store.APIKeyStore, authStore store.AuthStore, token string) (*store.APIKey, *store.AuthUser, error) {
	cfg := config.GetConfig()
	for _, validKey := range cfg.Api.Keys {
		if validKey != "" && subtle.ConstantTimeCompare([]byte(token), []byte(validKey)) == 1 {
			adminUser, err := getAdminUser(ctx, authStore)
			if err != nil {
				return nil, nil, err
			}
			return &store.APIKey{ID: 0, Name: "legacy-config-key", UserID: adminUser.ID}, adminUser, nil
		}
	}

	keyHash := HashAPIKey(token)
	apiKey, err := apiKeyStore.GetByKeyHash(ctx, keyHash)
	if err != nil {
		return nil, nil, err
	}
	if apiKey == nil {
		return nil, nil, errors.New("API key not found")
	}

	user, err := authStore.GetUserByID(ctx, apiKey.UserID)
	if err != nil {
		return nil, nil, err
	}

	return apiKey, user, nil
}

func getAdminUser(ctx context.Context, authStore store.AuthStore) (*store.AuthUser, error) {
	cfg := config.GetConfig()
	adminID, err := authStore.GetUserIDByUsername(ctx, cfg.DB.Admin.Username)
	if err != nil {
		return nil, err
	}
	return authStore.GetUserByID(ctx, adminID)
}

func HashAPIKey(key string) string {
	hash := sha256.Sum256([]byte(key))
	return hex.EncodeToString(hash[:])
}
