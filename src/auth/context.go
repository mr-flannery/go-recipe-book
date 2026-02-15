package auth

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/mr-flannery/go-recipe-book/src/store"
)

type contextKey string

const userInfoKey contextKey = "userInfo"

type UserInfo struct {
	IsLoggedIn bool
	IsAdmin    bool
	Username   string
	UserID     int
}

func UserContextMiddleware(authStore store.AuthStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := GetUserBySession(authStore, r)
			if err != nil {
				userInfo := &UserInfo{
					IsLoggedIn: false,
					IsAdmin:    false,
					Username:   "",
					UserID:     0,
				}
				ctx := context.WithValue(r.Context(), userInfoKey, userInfo)
				r = r.WithContext(ctx)
			} else {
				slog.Debug("User context middleware found valid session", "username", user.Username, "userID", user.ID)
				userInfo := &UserInfo{
					IsLoggedIn: true,
					IsAdmin:    user.IsAdmin,
					Username:   user.Username,
					UserID:     user.ID,
				}
				ctx := context.WithValue(r.Context(), userInfoKey, userInfo)
				r = r.WithContext(ctx)
			}
			next.ServeHTTP(w, r)
		})
	}
}

func GetUserInfoFromContext(ctx context.Context) *UserInfo {
	userInfo, ok := ctx.Value(userInfoKey).(*UserInfo)
	if !ok {
		return &UserInfo{
			IsLoggedIn: false,
			IsAdmin:    false,
			Username:   "",
			UserID:     0,
		}
	}
	return userInfo
}

func GetUserFromContext(ctx context.Context) *User {
	userInfo := GetUserInfoFromContext(ctx)
	if !userInfo.IsLoggedIn {
		return nil
	}
	return &User{
		ID:       userInfo.UserID,
		Username: userInfo.Username,
		IsAdmin:  userInfo.IsAdmin,
	}
}

func IsUserAdmin(ctx context.Context) bool {
	userInfo := GetUserInfoFromContext(ctx)
	return userInfo.IsLoggedIn && userInfo.IsAdmin
}

func GetUsernameFromContext(ctx context.Context) string {
	userInfo := GetUserInfoFromContext(ctx)
	return userInfo.Username
}

func GetUserIDFromContext(ctx context.Context) int {
	userInfo := GetUserInfoFromContext(ctx)
	return userInfo.UserID
}

func ContextWithUserInfo(ctx context.Context, userInfo *UserInfo) context.Context {
	return context.WithValue(ctx, userInfoKey, userInfo)
}
