package mocks

import (
	"context"
	"time"

	"github.com/mr-flannery/go-recipe-book/src/store"
)

type MockAuthStore struct {
	GetUserByEmailFunc                   func(ctx context.Context, email string) (*store.AuthUser, string, error)
	UpdateLastLoginFunc                  func(ctx context.Context, userID int) error
	GetUserByIDFunc                      func(ctx context.Context, userID int) (*store.AuthUser, error)
	GetFullUserByIDFunc                  func(ctx context.Context, userID int) (*store.FullAuthUser, error)
	GetUserIDByUsernameFunc              func(ctx context.Context, username string) (int, error)
	CreateSessionFunc                    func(ctx context.Context, session *store.Session) error
	GetSessionFunc                       func(ctx context.Context, sessionID string) (*store.Session, error)
	DeleteSessionFunc                    func(ctx context.Context, sessionID string) error
	DeleteExpiredSessionsFunc            func(ctx context.Context) (int64, error)
	DeleteUserSessionsFunc               func(ctx context.Context, userID int) error
	GetActiveSessionCountFunc            func(ctx context.Context, userID int) (int, error)
	ExtendSessionFunc                    func(ctx context.Context, sessionID string) error
	CreateRegistrationRequestFunc        func(ctx context.Context, username, email, passwordHash string) error
	GetPendingRegistrationsFunc          func(ctx context.Context) ([]store.RegistrationRequest, error)
	GetAllRegistrationsFunc              func(ctx context.Context) ([]store.RegistrationRequest, error)
	GetAllRegistrationsPaginatedFunc     func(ctx context.Context, limit, offset int) ([]store.RegistrationRequest, error)
	CountAllRegistrationsFunc            func(ctx context.Context) (int, error)
	ApproveRegistrationFunc              func(ctx context.Context, requestID, adminID int) error
	RejectRegistrationFunc               func(ctx context.Context, requestID, adminID int) error
	CreateUserFunc                       func(ctx context.Context, username, email, passwordHash string, isAdmin bool) error
	UserExistsFunc                       func(ctx context.Context, username string) (bool, error)
	GetAllUsersFunc                      func(ctx context.Context) ([]store.AuthUser, error)
	DeleteUserFunc                       func(ctx context.Context, userID int) error
	CreatePasswordResetTokenFunc         func(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) error
	GetPasswordResetTokenFunc            func(ctx context.Context, tokenHash string) (*store.PasswordResetToken, error)
	MarkPasswordResetTokenUsedFunc       func(ctx context.Context, tokenHash string) error
	DeleteExpiredPasswordResetTokensFunc func(ctx context.Context) (int64, error)
	UpdateUserPasswordFunc               func(ctx context.Context, userID int, passwordHash string) error
	ResetPasswordWithTokenFunc           func(ctx context.Context, tokenHash string, newPasswordHash string) (int, error)
}

func (m *MockAuthStore) GetUserByEmail(ctx context.Context, email string) (*store.AuthUser, string, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(ctx, email)
	}
	return nil, "", nil
}

func (m *MockAuthStore) UpdateLastLogin(ctx context.Context, userID int) error {
	if m.UpdateLastLoginFunc != nil {
		return m.UpdateLastLoginFunc(ctx, userID)
	}
	return nil
}

func (m *MockAuthStore) GetUserByID(ctx context.Context, userID int) (*store.AuthUser, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockAuthStore) GetFullUserByID(ctx context.Context, userID int) (*store.FullAuthUser, error) {
	if m.GetFullUserByIDFunc != nil {
		return m.GetFullUserByIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockAuthStore) GetUserIDByUsername(ctx context.Context, username string) (int, error) {
	if m.GetUserIDByUsernameFunc != nil {
		return m.GetUserIDByUsernameFunc(ctx, username)
	}
	return 0, nil
}

func (m *MockAuthStore) CreateSession(ctx context.Context, session *store.Session) error {
	if m.CreateSessionFunc != nil {
		return m.CreateSessionFunc(ctx, session)
	}
	return nil
}

func (m *MockAuthStore) GetSession(ctx context.Context, sessionID string) (*store.Session, error) {
	if m.GetSessionFunc != nil {
		return m.GetSessionFunc(ctx, sessionID)
	}
	return nil, nil
}

func (m *MockAuthStore) DeleteSession(ctx context.Context, sessionID string) error {
	if m.DeleteSessionFunc != nil {
		return m.DeleteSessionFunc(ctx, sessionID)
	}
	return nil
}

func (m *MockAuthStore) DeleteExpiredSessions(ctx context.Context) (int64, error) {
	if m.DeleteExpiredSessionsFunc != nil {
		return m.DeleteExpiredSessionsFunc(ctx)
	}
	return 0, nil
}

func (m *MockAuthStore) DeleteUserSessions(ctx context.Context, userID int) error {
	if m.DeleteUserSessionsFunc != nil {
		return m.DeleteUserSessionsFunc(ctx, userID)
	}
	return nil
}

func (m *MockAuthStore) GetActiveSessionCount(ctx context.Context, userID int) (int, error) {
	if m.GetActiveSessionCountFunc != nil {
		return m.GetActiveSessionCountFunc(ctx, userID)
	}
	return 0, nil
}

func (m *MockAuthStore) ExtendSession(ctx context.Context, sessionID string) error {
	if m.ExtendSessionFunc != nil {
		return m.ExtendSessionFunc(ctx, sessionID)
	}
	return nil
}

func (m *MockAuthStore) CreateRegistrationRequest(ctx context.Context, username, email, passwordHash string) error {
	if m.CreateRegistrationRequestFunc != nil {
		return m.CreateRegistrationRequestFunc(ctx, username, email, passwordHash)
	}
	return nil
}

func (m *MockAuthStore) GetPendingRegistrations(ctx context.Context) ([]store.RegistrationRequest, error) {
	if m.GetPendingRegistrationsFunc != nil {
		return m.GetPendingRegistrationsFunc(ctx)
	}
	return nil, nil
}

func (m *MockAuthStore) GetAllRegistrations(ctx context.Context) ([]store.RegistrationRequest, error) {
	if m.GetAllRegistrationsFunc != nil {
		return m.GetAllRegistrationsFunc(ctx)
	}
	return nil, nil
}

func (m *MockAuthStore) GetAllRegistrationsPaginated(ctx context.Context, limit, offset int) ([]store.RegistrationRequest, error) {
	if m.GetAllRegistrationsPaginatedFunc != nil {
		return m.GetAllRegistrationsPaginatedFunc(ctx, limit, offset)
	}
	return nil, nil
}

func (m *MockAuthStore) CountAllRegistrations(ctx context.Context) (int, error) {
	if m.CountAllRegistrationsFunc != nil {
		return m.CountAllRegistrationsFunc(ctx)
	}
	return 0, nil
}

func (m *MockAuthStore) ApproveRegistration(ctx context.Context, requestID, adminID int) error {
	if m.ApproveRegistrationFunc != nil {
		return m.ApproveRegistrationFunc(ctx, requestID, adminID)
	}
	return nil
}

func (m *MockAuthStore) RejectRegistration(ctx context.Context, requestID, adminID int) error {
	if m.RejectRegistrationFunc != nil {
		return m.RejectRegistrationFunc(ctx, requestID, adminID)
	}
	return nil
}

func (m *MockAuthStore) CreateUser(ctx context.Context, username, email, passwordHash string, isAdmin bool) error {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(ctx, username, email, passwordHash, isAdmin)
	}
	return nil
}

func (m *MockAuthStore) UserExists(ctx context.Context, username string) (bool, error) {
	if m.UserExistsFunc != nil {
		return m.UserExistsFunc(ctx, username)
	}
	return false, nil
}

func (m *MockAuthStore) GetAllUsers(ctx context.Context) ([]store.AuthUser, error) {
	if m.GetAllUsersFunc != nil {
		return m.GetAllUsersFunc(ctx)
	}
	return nil, nil
}

func (m *MockAuthStore) DeleteUser(ctx context.Context, userID int) error {
	if m.DeleteUserFunc != nil {
		return m.DeleteUserFunc(ctx, userID)
	}
	return nil
}

func (m *MockAuthStore) CreatePasswordResetToken(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) error {
	if m.CreatePasswordResetTokenFunc != nil {
		return m.CreatePasswordResetTokenFunc(ctx, userID, tokenHash, expiresAt)
	}
	return nil
}

func (m *MockAuthStore) GetPasswordResetToken(ctx context.Context, tokenHash string) (*store.PasswordResetToken, error) {
	if m.GetPasswordResetTokenFunc != nil {
		return m.GetPasswordResetTokenFunc(ctx, tokenHash)
	}
	return nil, nil
}

func (m *MockAuthStore) MarkPasswordResetTokenUsed(ctx context.Context, tokenHash string) error {
	if m.MarkPasswordResetTokenUsedFunc != nil {
		return m.MarkPasswordResetTokenUsedFunc(ctx, tokenHash)
	}
	return nil
}

func (m *MockAuthStore) DeleteExpiredPasswordResetTokens(ctx context.Context) (int64, error) {
	if m.DeleteExpiredPasswordResetTokensFunc != nil {
		return m.DeleteExpiredPasswordResetTokensFunc(ctx)
	}
	return 0, nil
}

func (m *MockAuthStore) UpdateUserPassword(ctx context.Context, userID int, passwordHash string) error {
	if m.UpdateUserPasswordFunc != nil {
		return m.UpdateUserPasswordFunc(ctx, userID, passwordHash)
	}
	return nil
}

func (m *MockAuthStore) ResetPasswordWithToken(ctx context.Context, tokenHash string, newPasswordHash string) (int, error) {
	if m.ResetPasswordWithTokenFunc != nil {
		return m.ResetPasswordWithTokenFunc(ctx, tokenHash, newPasswordHash)
	}
	return 0, nil
}
