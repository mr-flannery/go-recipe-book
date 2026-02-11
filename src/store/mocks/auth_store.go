package mocks

import "github.com/mr-flannery/go-recipe-book/src/store"

type MockAuthStore struct {
	GetUserByEmailFunc            func(email string) (*store.AuthUser, string, error)
	UpdateLastLoginFunc           func(userID int) error
	GetUserByIDFunc               func(userID int) (*store.AuthUser, error)
	GetUserIDByUsernameFunc       func(username string) (int, error)
	CreateSessionFunc             func(session *store.Session) error
	GetSessionFunc                func(sessionID string) (*store.Session, error)
	DeleteSessionFunc             func(sessionID string) error
	DeleteExpiredSessionsFunc     func() (int64, error)
	DeleteUserSessionsFunc        func(userID int) error
	GetActiveSessionCountFunc     func(userID int) (int, error)
	ExtendSessionFunc             func(sessionID string) error
	CreateRegistrationRequestFunc func(username, email, passwordHash string) error
	GetPendingRegistrationsFunc   func() ([]store.RegistrationRequest, error)
	ApproveRegistrationFunc       func(requestID, adminID int) error
	RejectRegistrationFunc        func(requestID, adminID int, reason string) error
	CreateUserFunc                func(username, email, passwordHash string, isAdmin bool) error
	UserExistsFunc                func(username string) (bool, error)
}

func (m *MockAuthStore) GetUserByEmail(email string) (*store.AuthUser, string, error) {
	if m.GetUserByEmailFunc != nil {
		return m.GetUserByEmailFunc(email)
	}
	return nil, "", nil
}

func (m *MockAuthStore) UpdateLastLogin(userID int) error {
	if m.UpdateLastLoginFunc != nil {
		return m.UpdateLastLoginFunc(userID)
	}
	return nil
}

func (m *MockAuthStore) GetUserByID(userID int) (*store.AuthUser, error) {
	if m.GetUserByIDFunc != nil {
		return m.GetUserByIDFunc(userID)
	}
	return nil, nil
}

func (m *MockAuthStore) GetUserIDByUsername(username string) (int, error) {
	if m.GetUserIDByUsernameFunc != nil {
		return m.GetUserIDByUsernameFunc(username)
	}
	return 0, nil
}

func (m *MockAuthStore) CreateSession(session *store.Session) error {
	if m.CreateSessionFunc != nil {
		return m.CreateSessionFunc(session)
	}
	return nil
}

func (m *MockAuthStore) GetSession(sessionID string) (*store.Session, error) {
	if m.GetSessionFunc != nil {
		return m.GetSessionFunc(sessionID)
	}
	return nil, nil
}

func (m *MockAuthStore) DeleteSession(sessionID string) error {
	if m.DeleteSessionFunc != nil {
		return m.DeleteSessionFunc(sessionID)
	}
	return nil
}

func (m *MockAuthStore) DeleteExpiredSessions() (int64, error) {
	if m.DeleteExpiredSessionsFunc != nil {
		return m.DeleteExpiredSessionsFunc()
	}
	return 0, nil
}

func (m *MockAuthStore) DeleteUserSessions(userID int) error {
	if m.DeleteUserSessionsFunc != nil {
		return m.DeleteUserSessionsFunc(userID)
	}
	return nil
}

func (m *MockAuthStore) GetActiveSessionCount(userID int) (int, error) {
	if m.GetActiveSessionCountFunc != nil {
		return m.GetActiveSessionCountFunc(userID)
	}
	return 0, nil
}

func (m *MockAuthStore) ExtendSession(sessionID string) error {
	if m.ExtendSessionFunc != nil {
		return m.ExtendSessionFunc(sessionID)
	}
	return nil
}

func (m *MockAuthStore) CreateRegistrationRequest(username, email, passwordHash string) error {
	if m.CreateRegistrationRequestFunc != nil {
		return m.CreateRegistrationRequestFunc(username, email, passwordHash)
	}
	return nil
}

func (m *MockAuthStore) GetPendingRegistrations() ([]store.RegistrationRequest, error) {
	if m.GetPendingRegistrationsFunc != nil {
		return m.GetPendingRegistrationsFunc()
	}
	return nil, nil
}

func (m *MockAuthStore) ApproveRegistration(requestID, adminID int) error {
	if m.ApproveRegistrationFunc != nil {
		return m.ApproveRegistrationFunc(requestID, adminID)
	}
	return nil
}

func (m *MockAuthStore) RejectRegistration(requestID, adminID int, reason string) error {
	if m.RejectRegistrationFunc != nil {
		return m.RejectRegistrationFunc(requestID, adminID, reason)
	}
	return nil
}

func (m *MockAuthStore) CreateUser(username, email, passwordHash string, isAdmin bool) error {
	if m.CreateUserFunc != nil {
		return m.CreateUserFunc(username, email, passwordHash, isAdmin)
	}
	return nil
}

func (m *MockAuthStore) UserExists(username string) (bool, error) {
	if m.UserExistsFunc != nil {
		return m.UserExistsFunc(username)
	}
	return false, nil
}
