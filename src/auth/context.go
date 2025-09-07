package auth

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/mr-flannery/go-recipe-book/src/db"
)

// Context keys for storing user information
type contextKey string

const userInfoKey contextKey = "userInfo"

// UserInfo contains the essential user information needed by templates
type UserInfo struct {
	IsLoggedIn bool
	IsAdmin    bool
	Username   string
}

// UserContextMiddleware extracts user information from session and stores UserInfo in request context
// This middleware runs on ALL requests and never blocks - it just enriches the context
func UserContextMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get fresh DB connection using existing pattern
			database, err := db.GetConnection()
			if err != nil {
				slog.Error("Failed to get database connection in UserContextMiddleware", "error", err)
				// If DB fails, continue with anonymous user
				userInfo := &UserInfo{
					IsLoggedIn: false,
					IsAdmin:    false,
					Username:   "",
				}
				ctx := context.WithValue(r.Context(), userInfoKey, userInfo)
				r = r.WithContext(ctx)
				next.ServeHTTP(w, r)
				return
			}
			defer database.Close()

			user, err := GetUserBySession(database, r)
			if err != nil {
				// No valid session - store anonymous user info
				userInfo := &UserInfo{
					IsLoggedIn: false,
					IsAdmin:    false,
					Username:   "",
				}
				ctx := context.WithValue(r.Context(), userInfoKey, userInfo)
				r = r.WithContext(ctx)
			} else {
				// Valid session - store user info
				slog.Debug("User context middleware found valid session", "username", user.Username, "userID", user.ID)
				userInfo := &UserInfo{
					IsLoggedIn: true,
					IsAdmin:    user.IsAdmin,
					Username:   user.Username,
				}
				ctx := context.WithValue(r.Context(), userInfoKey, userInfo)
				r = r.WithContext(ctx)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// GetUserInfoFromContext extracts UserInfo from the request context
// Returns UserInfo with IsLoggedIn=false if no user info is found
func GetUserInfoFromContext(ctx context.Context) *UserInfo {
	userInfo, ok := ctx.Value(userInfoKey).(*UserInfo)
	if !ok {
		// Return anonymous user info as fallback
		return &UserInfo{
			IsLoggedIn: false,
			IsAdmin:    false,
			Username:   "",
		}
	}
	return userInfo
}

// Legacy helper functions for backward compatibility with existing RequireAuth middleware
// These functions still work with the new UserInfo approach

// GetUserFromContext extracts user information from the request context
// Returns nil if no user is logged in (anonymous user)
// NOTE: This is kept for backward compatibility with RequireAuth middleware
func GetUserFromContext(ctx context.Context) *User {
	userInfo := GetUserInfoFromContext(ctx)
	if !userInfo.IsLoggedIn {
		return nil
	}
	// We can't reconstruct the full User object from UserInfo alone
	// RequireAuth middleware only needs to check if user exists (non-nil)
	// So we return a minimal User object for compatibility
	return &User{
		Username: userInfo.Username,
		IsAdmin:  userInfo.IsAdmin,
	}
}

// IsUserLoggedIn checks if there's a valid user in the context
func IsUserLoggedIn(ctx context.Context) bool {
	userInfo := GetUserInfoFromContext(ctx)
	return userInfo.IsLoggedIn
}

// IsUserAdmin checks if the current user is an admin
func IsUserAdmin(ctx context.Context) bool {
	userInfo := GetUserInfoFromContext(ctx)
	return userInfo.IsLoggedIn && userInfo.IsAdmin
}

// GetUsernameFromContext returns the username of the current user, or empty string if not logged in
func GetUsernameFromContext(ctx context.Context) string {
	userInfo := GetUserInfoFromContext(ctx)
	return userInfo.Username
}

// GetUserIDFromContext returns the user ID of the current user, or 0 if not logged in
// NOTE: UserInfo doesn't contain UserID, so this function is deprecated
// Use GetUserInfoFromContext() instead for new code
func GetUserIDFromContext(ctx context.Context) int {
	userInfo := GetUserInfoFromContext(ctx)
	if !userInfo.IsLoggedIn {
		return 0
	}
	// We can't get UserID from UserInfo, so return 0
	// This function is kept for backward compatibility only
	return 0
}
