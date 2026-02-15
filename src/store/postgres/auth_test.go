package postgres

import (
	"testing"
	"time"

	"github.com/mr-flannery/go-recipe-book/src/store"
	"github.com/mr-flannery/go-recipe-book/src/testutil"
)

func TestAuthStore_GetUserByEmail_ReturnsUserWhenExists(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	authStore := NewAuthStore(testDB.DB)

	user, passwordHash, err := authStore.GetUserByEmail("test@example.com")
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

	_, _, err := authStore.GetUserByEmail("nonexistent@example.com")
	if err == nil {
		t.Error("expected error for non-existent user")
	}
}

func TestAuthStore_GetUserByID_ReturnsUserWhenExists(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", true)
	authStore := NewAuthStore(testDB.DB)

	user, err := authStore.GetUserByID(userID)
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

	userID, err := authStore.GetUserIDByUsername("testuser")
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

	err := authStore.CreateSession(session)
	if err != nil {
		t.Fatalf("failed to create session: %v", err)
	}

	retrieved, err := authStore.GetSession("test-session-id")
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

	_, err := authStore.GetSession("")
	if err == nil {
		t.Error("expected error for empty session ID")
	}
}

func TestAuthStore_GetSession_ReturnsErrorForExpiredSession(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	testDB.SeedSession(t, "expired-session", userID, time.Now().Add(-1*time.Hour))
	authStore := NewAuthStore(testDB.DB)

	_, err := authStore.GetSession("expired-session")
	if err == nil {
		t.Error("expected error for expired session")
	}
}

func TestAuthStore_DeleteSession_RemovesSession(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	testDB.SeedSession(t, "to-delete", userID, time.Now().Add(24*time.Hour))
	authStore := NewAuthStore(testDB.DB)

	err := authStore.DeleteSession("to-delete")
	if err != nil {
		t.Fatalf("failed to delete session: %v", err)
	}

	_, err = authStore.GetSession("to-delete")
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

	deleted, err := authStore.DeleteExpiredSessions()
	if err != nil {
		t.Fatalf("failed to delete expired sessions: %v", err)
	}

	if deleted != 2 {
		t.Errorf("expected 2 deleted sessions, got %d", deleted)
	}

	_, err = authStore.GetSession("valid")
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

	err := authStore.DeleteUserSessions(userID)
	if err != nil {
		t.Fatalf("failed to delete user sessions: %v", err)
	}

	count, _ := authStore.GetActiveSessionCount(userID)
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

	count, err := authStore.GetActiveSessionCount(userID)
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

	err := authStore.ExtendSession("to-extend")
	if err != nil {
		t.Fatalf("failed to extend session: %v", err)
	}
}

func TestAuthStore_ExtendSession_ReturnsErrorForExpiredSession(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	testDB.SeedSession(t, "expired", userID, time.Now().Add(-1*time.Hour))
	authStore := NewAuthStore(testDB.DB)

	err := authStore.ExtendSession("expired")
	if err == nil {
		t.Error("expected error when extending expired session")
	}
}

func TestAuthStore_CreateRegistrationRequest_CreatesRequest(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	authStore := NewAuthStore(testDB.DB)

	err := authStore.CreateRegistrationRequest("newuser", "new@example.com", "hashedpassword")
	if err != nil {
		t.Fatalf("failed to create registration request: %v", err)
	}

	requests, _ := authStore.GetPendingRegistrations()
	if len(requests) != 1 {
		t.Errorf("expected 1 pending registration, got %d", len(requests))
	}
}

func TestAuthStore_CreateRegistrationRequest_RejectsExistingUser(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	testDB.SeedUser(t, "existinguser", "existing@example.com", "hashedpassword", false)
	authStore := NewAuthStore(testDB.DB)

	err := authStore.CreateRegistrationRequest("existinguser", "new@example.com", "hashedpassword")
	if err == nil {
		t.Error("expected error for duplicate username")
	}
}

func TestAuthStore_CreateRegistrationRequest_RejectsDuplicatePendingRequest(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	authStore := NewAuthStore(testDB.DB)

	err := authStore.CreateRegistrationRequest("newuser", "new@example.com", "hashedpassword")
	if err != nil {
		t.Fatalf("failed to create first registration request: %v", err)
	}

	err = authStore.CreateRegistrationRequest("newuser", "different@example.com", "hashedpassword")
	if err == nil {
		t.Error("expected error for duplicate pending request")
	}
}

func TestAuthStore_ApproveRegistration_CreatesUserAndUpdatesRequest(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	adminID := testDB.SeedUser(t, "admin", "admin@example.com", "hashedpassword", true)
	authStore := NewAuthStore(testDB.DB)

	authStore.CreateRegistrationRequest("newuser", "new@example.com", "hashedpassword")
	requests, _ := authStore.GetPendingRegistrations()
	requestID := requests[0].ID

	err := authStore.ApproveRegistration(requestID, adminID)
	if err != nil {
		t.Fatalf("failed to approve registration: %v", err)
	}

	user, _, _ := authStore.GetUserByEmail("new@example.com")
	if user == nil {
		t.Error("expected user to be created after approval")
	}

	pendingRequests, _ := authStore.GetPendingRegistrations()
	if len(pendingRequests) != 0 {
		t.Error("expected no pending requests after approval")
	}
}

func TestAuthStore_RejectRegistration_UpdatesRequestStatus(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	adminID := testDB.SeedUser(t, "admin", "admin@example.com", "hashedpassword", true)
	authStore := NewAuthStore(testDB.DB)

	authStore.CreateRegistrationRequest("newuser", "new@example.com", "hashedpassword")
	requests, _ := authStore.GetPendingRegistrations()
	requestID := requests[0].ID

	err := authStore.RejectRegistration(requestID, adminID)
	if err != nil {
		t.Fatalf("failed to reject registration: %v", err)
	}

	pendingRequests, _ := authStore.GetPendingRegistrations()
	if len(pendingRequests) != 0 {
		t.Error("expected no pending requests after rejection")
	}
}

func TestAuthStore_CreateUser_CreatesNewUser(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	authStore := NewAuthStore(testDB.DB)

	err := authStore.CreateUser("newuser", "new@example.com", "hashedpassword", false)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	user, _, _ := authStore.GetUserByEmail("new@example.com")
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

	exists, err := authStore.UserExists("existinguser")
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

	exists, err := authStore.UserExists("nonexistent")
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

	users, err := authStore.GetAllUsers()
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

	err := authStore.DeleteUser(userID)
	if err != nil {
		t.Fatalf("failed to delete user: %v", err)
	}

	_, _, err = authStore.GetUserByEmail("todelete@example.com")
	if err == nil {
		t.Error("expected error after deleting user")
	}
}

func TestAuthStore_DeleteUser_AlsoDeletesSessions(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "todelete", "todelete@example.com", "hashedpassword", false)
	testDB.SeedSession(t, "user-session", userID, time.Now().Add(24*time.Hour))
	authStore := NewAuthStore(testDB.DB)

	err := authStore.DeleteUser(userID)
	if err != nil {
		t.Fatalf("failed to delete user: %v", err)
	}

	_, err = authStore.GetSession("user-session")
	if err == nil {
		t.Error("expected session to be deleted with user")
	}
}

func TestAuthStore_UpdateLastLogin_UpdatesTimestamp(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	authStore := NewAuthStore(testDB.DB)

	err := authStore.UpdateLastLogin(userID)
	if err != nil {
		t.Fatalf("failed to update last login: %v", err)
	}
}
