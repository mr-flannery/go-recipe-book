package postgres

import (
	"context"
	"testing"
	"time"

	"github.com/mr-flannery/go-recipe-book/src/store"
	"github.com/mr-flannery/go-recipe-book/src/testutil"
)

func TestAuthStore_GetUserByEmail_ReturnsUserWhenExists(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	authStore := NewAuthStore(testDB.DB)

	user, passwordHash, err := authStore.GetUserByEmail(context.Background(), "test@example.com")
	if err != nil {
		t.Fatalf("failed to get user by email: %v", err)
	}

	if user.Username != "testuser" {
		t.Errorf("expected username 'testuser', got '%s'", user.Username)
	}
	if user.Email != "test@example.com" {
		t.Errorf("expected email 'test@example.com', got '%s'", user.Email)
	}
	if passwordHash != "hashedpassword" {
		t.Errorf("expected password hash 'hashedpassword', got '%s'", passwordHash)
	}
}

func TestAuthStore_GetUserByEmail_ReturnsErrorWhenNotFound(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	authStore := NewAuthStore(testDB.DB)

	_, _, err := authStore.GetUserByEmail(context.Background(), "nonexistent@example.com")
	if err == nil {
		t.Error("expected error for non-existent user")
	}
}

func TestAuthStore_GetUserByID_ReturnsUserWhenExists(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", true)
	authStore := NewAuthStore(testDB.DB)

	user, err := authStore.GetUserByID(context.Background(), userID)
	if err != nil {
		t.Fatalf("failed to get user by ID: %v", err)
	}

	if user.ID != userID {
		t.Errorf("expected user ID %d, got %d", userID, user.ID)
	}
	if !user.IsAdmin {
		t.Error("expected user to be admin")
	}
}

func TestAuthStore_GetUserIDByUsername_ReturnsIDWhenExists(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	expectedID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	authStore := NewAuthStore(testDB.DB)

	userID, err := authStore.GetUserIDByUsername(context.Background(), "testuser")
	if err != nil {
		t.Fatalf("failed to get user ID by username: %v", err)
	}

	if userID != expectedID {
		t.Errorf("expected user ID %d, got %d", expectedID, userID)
	}
}

func TestAuthStore_CreateSession_CreatesAndRetrievesSession(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	authStore := NewAuthStore(testDB.DB)

	session := &store.Session{
		ID:        "test-session-id",
		UserID:    userID,
		IPAddress: "127.0.0.1",
		UserAgent: "TestAgent/1.0",
	}

	err := authStore.CreateSession(context.Background(), session)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	retrieved, err := authStore.GetSession(context.Background(), "test-session-id")
	if err != nil {
		t.Fatalf("failed to get session: %v", err)
	}

	if retrieved.ID != session.ID {
		t.Errorf("expected session ID '%s', got '%s'", session.ID, retrieved.ID)
	}
	if retrieved.UserID != session.UserID {
		t.Errorf("expected user ID %d, got %d", session.UserID, retrieved.UserID)
	}
}

func TestAuthStore_GetSession_ReturnsErrorForEmptyID(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	authStore := NewAuthStore(testDB.DB)

	_, err := authStore.GetSession(context.Background(), "")
	if err == nil {
		t.Error("expected error for empty session ID")
	}
}

func TestAuthStore_GetSession_ReturnsErrorForExpiredSession(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	testDB.SeedSession(t, "expired-session", userID, time.Now().Add(-1*time.Hour))
	authStore := NewAuthStore(testDB.DB)

	_, err := authStore.GetSession(context.Background(), "expired-session")
	if err == nil {
		t.Error("expected error for expired session")
	}
}

func TestAuthStore_DeleteSession_RemovesSession(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	testDB.SeedSession(t, "to-delete", userID, time.Now().Add(24*time.Hour))
	authStore := NewAuthStore(testDB.DB)

	err := authStore.DeleteSession(context.Background(), "to-delete")
	if err != nil {
		t.Fatalf("failed to delete session: %v", err)
	}

	_, err = authStore.GetSession(context.Background(), "to-delete")
	if err == nil {
		t.Error("expected error after deleting session")
	}
}

func TestAuthStore_DeleteExpiredSessions_RemovesOnlyExpired(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	testDB.SeedSession(t, "expired-1", userID, time.Now().Add(-1*time.Hour))
	testDB.SeedSession(t, "expired-2", userID, time.Now().Add(-2*time.Hour))
	testDB.SeedSession(t, "valid", userID, time.Now().Add(24*time.Hour))
	authStore := NewAuthStore(testDB.DB)

	deleted, err := authStore.DeleteExpiredSessions(context.Background())
	if err != nil {
		t.Fatalf("failed to delete expired sessions: %v", err)
	}

	if deleted != 2 {
		t.Errorf("expected 2 deleted sessions, got %d", deleted)
	}

	_, err = authStore.GetSession(context.Background(), "valid")
	if err != nil {
		t.Error("valid session should still exist")
	}
}

func TestAuthStore_DeleteUserSessions_RemovesAllUserSessions(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	testDB.SeedSession(t, "session-1", userID, time.Now().Add(24*time.Hour))
	testDB.SeedSession(t, "session-2", userID, time.Now().Add(24*time.Hour))
	authStore := NewAuthStore(testDB.DB)

	err := authStore.DeleteUserSessions(context.Background(), userID)
	if err != nil {
		t.Fatalf("failed to delete user sessions: %v", err)
	}

	count, _ := authStore.GetActiveSessionCount(context.Background(), userID)
	if count != 0 {
		t.Errorf("expected 0 sessions after deletion, got %d", count)
	}
}

func TestAuthStore_GetActiveSessionCount_ReturnsCorrectCount(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	testDB.SeedSession(t, "active-1", userID, time.Now().Add(24*time.Hour))
	testDB.SeedSession(t, "active-2", userID, time.Now().Add(24*time.Hour))
	testDB.SeedSession(t, "expired", userID, time.Now().Add(-1*time.Hour))
	authStore := NewAuthStore(testDB.DB)

	count, err := authStore.GetActiveSessionCount(context.Background(), userID)
	if err != nil {
		t.Fatalf("failed to get session count: %v", err)
	}

	if count != 2 {
		t.Errorf("expected 2 active sessions, got %d", count)
	}
}

func TestAuthStore_ExtendSession_ExtendsValidSession(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	testDB.SeedSession(t, "to-extend", userID, time.Now().Add(1*time.Hour))
	authStore := NewAuthStore(testDB.DB)

	err := authStore.ExtendSession(context.Background(), "to-extend")
	if err != nil {
		t.Fatalf("failed to extend session: %v", err)
	}
}

func TestAuthStore_ExtendSession_ReturnsErrorForExpiredSession(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	testDB.SeedSession(t, "expired", userID, time.Now().Add(-1*time.Hour))
	authStore := NewAuthStore(testDB.DB)

	err := authStore.ExtendSession(context.Background(), "expired")
	if err == nil {
		t.Error("expected error when extending expired session")
	}
}

func TestAuthStore_CreateRegistrationRequest_CreatesRequest(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	authStore := NewAuthStore(testDB.DB)

	err := authStore.CreateRegistrationRequest(context.Background(), "newuser", "new@example.com", "hashedpassword")
	if err != nil {
		t.Fatalf("failed to create registration request: %v", err)
	}

	requests, _ := authStore.GetPendingRegistrations(context.Background())
	if len(requests) != 1 {
		t.Errorf("expected 1 pending registration, got %d", len(requests))
	}
}

func TestAuthStore_CreateRegistrationRequest_RejectsExistingUser(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	testDB.SeedUser(t, "existinguser", "existing@example.com", "hashedpassword", false)
	authStore := NewAuthStore(testDB.DB)

	err := authStore.CreateRegistrationRequest(context.Background(), "existinguser", "new@example.com", "hashedpassword")
	if err == nil {
		t.Error("expected error for duplicate username")
	}
}

func TestAuthStore_CreateRegistrationRequest_RejectsDuplicatePendingRequest(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	authStore := NewAuthStore(testDB.DB)

	err := authStore.CreateRegistrationRequest(context.Background(), "newuser", "new@example.com", "hashedpassword")
	if err != nil {
		t.Fatalf("failed to create first registration request: %v", err)
	}

	err = authStore.CreateRegistrationRequest(context.Background(), "newuser", "different@example.com", "hashedpassword")
	if err == nil {
		t.Error("expected error for duplicate pending request")
	}
}

func TestAuthStore_ApproveRegistration_CreatesUserAndUpdatesRequest(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	adminID := testDB.SeedUser(t, "admin", "admin@example.com", "hashedpassword", true)
	authStore := NewAuthStore(testDB.DB)

	authStore.CreateRegistrationRequest(context.Background(), "newuser", "new@example.com", "hashedpassword")
	requests, _ := authStore.GetPendingRegistrations(context.Background())
	requestID := requests[0].ID

	err := authStore.ApproveRegistration(context.Background(), requestID, adminID)
	if err != nil {
		t.Fatalf("failed to approve registration: %v", err)
	}

	user, _, _ := authStore.GetUserByEmail(context.Background(), "new@example.com")
	if user == nil {
		t.Error("expected user to be created after approval")
	}

	pendingRequests, _ := authStore.GetPendingRegistrations(context.Background())
	if len(pendingRequests) != 0 {
		t.Error("expected no pending requests after approval")
	}
}

func TestAuthStore_RejectRegistration_UpdatesRequestStatus(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	adminID := testDB.SeedUser(t, "admin", "admin@example.com", "hashedpassword", true)
	authStore := NewAuthStore(testDB.DB)

	authStore.CreateRegistrationRequest(context.Background(), "newuser", "new@example.com", "hashedpassword")
	requests, _ := authStore.GetPendingRegistrations(context.Background())
	requestID := requests[0].ID

	err := authStore.RejectRegistration(context.Background(), requestID, adminID)
	if err != nil {
		t.Fatalf("failed to reject registration: %v", err)
	}

	pendingRequests, _ := authStore.GetPendingRegistrations(context.Background())
	if len(pendingRequests) != 0 {
		t.Error("expected no pending requests after rejection")
	}
}

func TestAuthStore_CreateUser_CreatesNewUser(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	authStore := NewAuthStore(testDB.DB)

	err := authStore.CreateUser(context.Background(), "newuser", "new@example.com", "hashedpassword", false)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	user, _, _ := authStore.GetUserByEmail(context.Background(), "new@example.com")
	if user == nil {
		t.Error("expected user to be created")
	}
	if user.Username != "newuser" {
		t.Errorf("expected username 'newuser', got '%s'", user.Username)
	}
}

func TestAuthStore_UserExists_ReturnsTrueWhenExists(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	testDB.SeedUser(t, "existinguser", "existing@example.com", "hashedpassword", false)
	authStore := NewAuthStore(testDB.DB)

	exists, err := authStore.UserExists(context.Background(), "existinguser")
	if err != nil {
		t.Fatalf("failed to check user exists: %v", err)
	}

	if !exists {
		t.Error("expected user to exist")
	}
}

func TestAuthStore_UserExists_ReturnsFalseWhenNotExists(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	authStore := NewAuthStore(testDB.DB)

	exists, err := authStore.UserExists(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("failed to check user exists: %v", err)
	}

	if exists {
		t.Error("expected user to not exist")
	}
}

func TestAuthStore_GetAllUsers_ReturnsAllUsers(t *testing.T) {
	testDB := testutil.GetTestDatabase(t)

	testDB.SeedUser(t, "user1", "user1@example.com", "hashedpassword", false)
	testDB.SeedUser(t, "user2", "user2@example.com", "hashedpassword", true)
	authStore := NewAuthStore(testDB.DB)

	users, err := authStore.GetAllUsers(context.Background())
	if err != nil {
		t.Fatalf("failed to get all users: %v", err)
	}

	if len(users) != 2 {
		t.Errorf("expected 2 users, got %d", len(users))
	}
}

func TestAuthStore_DeleteUser_RemovesUser(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "todelete", "todelete@example.com", "hashedpassword", false)
	authStore := NewAuthStore(testDB.DB)

	err := authStore.DeleteUser(context.Background(), userID)
	if err != nil {
		t.Fatalf("failed to delete user: %v", err)
	}

	_, _, err = authStore.GetUserByEmail(context.Background(), "todelete@example.com")
	if err == nil {
		t.Error("expected error after deleting user")
	}
}

func TestAuthStore_DeleteUser_AlsoDeletesSessions(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "todelete", "todelete@example.com", "hashedpassword", false)
	testDB.SeedSession(t, "user-session", userID, time.Now().Add(24*time.Hour))
	authStore := NewAuthStore(testDB.DB)

	err := authStore.DeleteUser(context.Background(), userID)
	if err != nil {
		t.Fatalf("failed to delete user: %v", err)
	}

	_, err = authStore.GetSession(context.Background(), "user-session")
	if err == nil {
		t.Error("expected session to be deleted with user")
	}
}

func TestAuthStore_UpdateLastLogin_UpdatesTimestamp(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	authStore := NewAuthStore(testDB.DB)

	err := authStore.UpdateLastLogin(context.Background(), userID)
	if err != nil {
		t.Fatalf("failed to update last login: %v", err)
	}
}

func TestAuthStore_CreatePasswordResetToken_CreatesToken(t *testing.T) {
	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	authStore := NewAuthStore(testDB.DB)

	expiresAt := time.Now().Add(24 * time.Hour)
	err := authStore.CreatePasswordResetToken(context.Background(), userID, "hashed-token-123", expiresAt)
	if err != nil {
		t.Fatalf("failed to create password reset token: %v", err)
	}

	token, err := authStore.GetPasswordResetToken(context.Background(), "hashed-token-123")
	if err != nil {
		t.Fatalf("failed to get password reset token: %v", err)
	}

	if token.UserID != userID {
		t.Errorf("expected user ID %d, got %d", userID, token.UserID)
	}
	if token.UsedAt != nil {
		t.Error("expected used_at to be nil for new token")
	}
}

func TestAuthStore_GetPasswordResetToken_ReturnsTokenWhenExists(t *testing.T) {
	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	expiresAt := time.Now().Add(24 * time.Hour)
	testDB.SeedPasswordResetToken(t, userID, "test-token-hash", expiresAt, nil)
	authStore := NewAuthStore(testDB.DB)

	token, err := authStore.GetPasswordResetToken(context.Background(), "test-token-hash")
	if err != nil {
		t.Fatalf("failed to get password reset token: %v", err)
	}

	if token.UserID != userID {
		t.Errorf("expected user ID %d, got %d", userID, token.UserID)
	}
}

func TestAuthStore_GetPasswordResetToken_ReturnsErrorWhenNotFound(t *testing.T) {
	testDB := testutil.GetTestDatabase(t)

	authStore := NewAuthStore(testDB.DB)

	_, err := authStore.GetPasswordResetToken(context.Background(), "nonexistent-token")
	if err == nil {
		t.Error("expected error for non-existent token")
	}
}

func TestAuthStore_MarkPasswordResetTokenUsed_SetsUsedAtTimestamp(t *testing.T) {
	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	expiresAt := time.Now().Add(24 * time.Hour)
	testDB.SeedPasswordResetToken(t, userID, "test-token-hash", expiresAt, nil)
	authStore := NewAuthStore(testDB.DB)

	err := authStore.MarkPasswordResetTokenUsed(context.Background(), "test-token-hash")
	if err != nil {
		t.Fatalf("failed to mark token as used: %v", err)
	}

	token, _ := authStore.GetPasswordResetToken(context.Background(), "test-token-hash")
	if token.UsedAt == nil {
		t.Error("expected used_at to be set after marking as used")
	}
}

func TestAuthStore_MarkPasswordResetTokenUsed_ReturnsErrorWhenNotFound(t *testing.T) {
	testDB := testutil.GetTestDatabase(t)

	authStore := NewAuthStore(testDB.DB)

	err := authStore.MarkPasswordResetTokenUsed(context.Background(), "nonexistent-token")
	if err == nil {
		t.Error("expected error for non-existent token")
	}
}

func TestAuthStore_DeleteExpiredPasswordResetTokens_RemovesExpiredAndUsedTokens(t *testing.T) {
	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	expiredTime := time.Now().Add(-1 * time.Hour)
	validTime := time.Now().Add(24 * time.Hour)
	usedAt := time.Now().Add(-30 * time.Minute)

	testDB.SeedPasswordResetToken(t, userID, "expired-token", expiredTime, nil)
	testDB.SeedPasswordResetToken(t, userID, "used-token", validTime, &usedAt)
	testDB.SeedPasswordResetToken(t, userID, "valid-token", validTime, nil)
	authStore := NewAuthStore(testDB.DB)

	deleted, err := authStore.DeleteExpiredPasswordResetTokens(context.Background())
	if err != nil {
		t.Fatalf("failed to delete expired tokens: %v", err)
	}

	if deleted != 2 {
		t.Errorf("expected 2 deleted tokens, got %d", deleted)
	}

	_, err = authStore.GetPasswordResetToken(context.Background(), "valid-token")
	if err != nil {
		t.Error("valid token should still exist")
	}
}

func TestAuthStore_UpdateUserPassword_UpdatesPasswordHash(t *testing.T) {
	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "oldpasswordhash", false)
	authStore := NewAuthStore(testDB.DB)

	err := authStore.UpdateUserPassword(context.Background(), userID, "newpasswordhash")
	if err != nil {
		t.Fatalf("failed to update user password: %v", err)
	}

	_, passwordHash, _ := authStore.GetUserByEmail(context.Background(), "test@example.com")
	if passwordHash != "newpasswordhash" {
		t.Errorf("expected password hash 'newpasswordhash', got '%s'", passwordHash)
	}
}

func TestAuthStore_UpdateUserPassword_ReturnsErrorWhenUserNotFound(t *testing.T) {
	testDB := testutil.GetTestDatabase(t)

	authStore := NewAuthStore(testDB.DB)

	err := authStore.UpdateUserPassword(context.Background(), 99999, "newpasswordhash")
	if err == nil {
		t.Error("expected error for non-existent user")
	}
}
