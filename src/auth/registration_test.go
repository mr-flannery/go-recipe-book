package auth

import (
	"errors"
	"testing"

	"github.com/mr-flannery/go-recipe-book/src/store"
	"github.com/mr-flannery/go-recipe-book/src/store/mocks"
)

func TestCreateRegistrationRequest_CreatesRequestWhenPasswordIsStrong(t *testing.T) {
	var capturedUsername, capturedEmail, capturedHash string
	mockStore := &mocks.MockAuthStore{
		CreateRegistrationRequestFunc: func(username, email, passwordHash string) error {
			capturedUsername = username
			capturedEmail = email
			capturedHash = passwordHash
			return nil
		},
	}

	err := CreateRegistrationRequest(mockStore, "newuser", "new@example.com", "ValidPassword123!")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if capturedUsername != "newuser" {
		t.Errorf("expected username 'newuser', got %s", capturedUsername)
	}
	if capturedEmail != "new@example.com" {
		t.Errorf("expected email 'new@example.com', got %s", capturedEmail)
	}
	if capturedHash == "" {
		t.Error("expected password hash to be set")
	}
}

func TestCreateRegistrationRequest_ReturnsErrorWhenPasswordIsWeak(t *testing.T) {
	mockStore := &mocks.MockAuthStore{}

	err := CreateRegistrationRequest(mockStore, "user", "user@example.com", "weak")
	if err == nil {
		t.Fatal("expected error for weak password, got nil")
	}
}

func TestCreateRegistrationRequest_ReturnsErrorWhenStoreFails(t *testing.T) {
	mockStore := &mocks.MockAuthStore{
		CreateRegistrationRequestFunc: func(username, email, passwordHash string) error {
			return errors.New("duplicate email")
		},
	}

	err := CreateRegistrationRequest(mockStore, "user", "existing@example.com", "StrongPassword123!")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetPendingRegistrations_ReturnsListOfPendingRequests(t *testing.T) {
	mockStore := &mocks.MockAuthStore{
		GetPendingRegistrationsFunc: func() ([]store.RegistrationRequest, error) {
			return []store.RegistrationRequest{
				{ID: 1, Username: "user1", Email: "user1@example.com", Status: "pending"},
				{ID: 2, Username: "user2", Email: "user2@example.com", Status: "pending"},
			}, nil
		},
	}

	reqs, err := GetPendingRegistrations(mockStore)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(reqs) != 2 {
		t.Fatalf("expected 2 registrations, got %d", len(reqs))
	}
	if reqs[0].Username != "user1" {
		t.Errorf("expected username 'user1', got %s", reqs[0].Username)
	}
	if reqs[1].Email != "user2@example.com" {
		t.Errorf("expected email 'user2@example.com', got %s", reqs[1].Email)
	}
}

func TestGetPendingRegistrations_ReturnsEmptyListWhenNoPendingRequests(t *testing.T) {
	mockStore := &mocks.MockAuthStore{
		GetPendingRegistrationsFunc: func() ([]store.RegistrationRequest, error) {
			return []store.RegistrationRequest{}, nil
		},
	}

	reqs, err := GetPendingRegistrations(mockStore)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(reqs) != 0 {
		t.Errorf("expected 0 registrations, got %d", len(reqs))
	}
}

func TestGetPendingRegistrations_ReturnsErrorWhenStoreFails(t *testing.T) {
	mockStore := &mocks.MockAuthStore{
		GetPendingRegistrationsFunc: func() ([]store.RegistrationRequest, error) {
			return nil, errors.New("database error")
		},
	}

	reqs, err := GetPendingRegistrations(mockStore)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if reqs != nil {
		t.Error("expected nil registrations on error")
	}
}

func TestApproveRegistration_ApprovesRequestAndRecordsAdminID(t *testing.T) {
	var approvedID, adminID int
	mockStore := &mocks.MockAuthStore{
		ApproveRegistrationFunc: func(requestID, aID int) error {
			approvedID = requestID
			adminID = aID
			return nil
		},
	}

	err := ApproveRegistration(mockStore, 5, 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if approvedID != 5 {
		t.Errorf("expected requestID 5, got %d", approvedID)
	}
	if adminID != 1 {
		t.Errorf("expected adminID 1, got %d", adminID)
	}
}

func TestRejectRegistration_RejectsRequestWithReasonAndAdminID(t *testing.T) {
	var rejectedID, adminID int
	var capturedReason string
	mockStore := &mocks.MockAuthStore{
		RejectRegistrationFunc: func(requestID, aID int, reason string) error {
			rejectedID = requestID
			adminID = aID
			capturedReason = reason
			return nil
		},
	}

	err := RejectRegistration(mockStore, 10, 2, "spam account")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if rejectedID != 10 {
		t.Errorf("expected requestID 10, got %d", rejectedID)
	}
	if adminID != 2 {
		t.Errorf("expected adminID 2, got %d", adminID)
	}
	if capturedReason != "spam account" {
		t.Errorf("expected reason 'spam account', got %s", capturedReason)
	}
}

func TestGetRegistrationRequestByID_ReturnsRequestWhenItExists(t *testing.T) {
	mockStore := &mocks.MockAuthStore{
		GetPendingRegistrationsFunc: func() ([]store.RegistrationRequest, error) {
			return []store.RegistrationRequest{
				{ID: 1, Username: "user1", Email: "user1@example.com", Status: "pending"},
				{ID: 5, Username: "target", Email: "target@example.com", Status: "pending"},
				{ID: 10, Username: "user3", Email: "user3@example.com", Status: "pending"},
			}, nil
		},
	}

	req, err := GetRegistrationRequestByID(mockStore, 5)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if req == nil {
		t.Fatal("expected request, got nil")
	}
	if req.ID != 5 {
		t.Errorf("expected ID 5, got %d", req.ID)
	}
	if req.Username != "target" {
		t.Errorf("expected username 'target', got %s", req.Username)
	}
}

func TestGetRegistrationRequestByID_ReturnsErrorWhenRequestNotFound(t *testing.T) {
	mockStore := &mocks.MockAuthStore{
		GetPendingRegistrationsFunc: func() ([]store.RegistrationRequest, error) {
			return []store.RegistrationRequest{
				{ID: 1, Username: "user1", Email: "user1@example.com", Status: "pending"},
			}, nil
		},
	}

	req, err := GetRegistrationRequestByID(mockStore, 999)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if req != nil {
		t.Error("expected nil request")
	}
	if err.Error() != "registration request not found" {
		t.Errorf("expected 'registration request not found', got %s", err.Error())
	}
}

func TestCreateSeedAdmin_CreatesAdminUserWhenNotExists(t *testing.T) {
	var createdUsername, createdEmail string
	var createdAsAdmin bool
	mockStore := &mocks.MockAuthStore{
		UserExistsFunc: func(username string) (bool, error) {
			return false, nil
		},
		CreateUserFunc: func(username, email, passwordHash string, isAdmin bool) error {
			createdUsername = username
			createdEmail = email
			createdAsAdmin = isAdmin
			return nil
		},
	}

	err := CreateSeedAdmin(mockStore, "admin", "admin@example.com", "AdminPass123!")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if createdUsername != "admin" {
		t.Errorf("expected username 'admin', got %s", createdUsername)
	}
	if createdEmail != "admin@example.com" {
		t.Errorf("expected email 'admin@example.com', got %s", createdEmail)
	}
	if !createdAsAdmin {
		t.Error("expected user to be created as admin")
	}
}

func TestCreateSeedAdmin_SkipsCreationWhenUserAlreadyExists(t *testing.T) {
	createUserCalled := false
	mockStore := &mocks.MockAuthStore{
		UserExistsFunc: func(username string) (bool, error) {
			return true, nil
		},
		CreateUserFunc: func(username, email, passwordHash string, isAdmin bool) error {
			createUserCalled = true
			return nil
		},
	}

	err := CreateSeedAdmin(mockStore, "admin", "admin@example.com", "AdminPass123!")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if createUserCalled {
		t.Error("CreateUser should not be called when user exists")
	}
}

func TestCreateSeedAdmin_ReturnsErrorWhenUserExistsCheckFails(t *testing.T) {
	mockStore := &mocks.MockAuthStore{
		UserExistsFunc: func(username string) (bool, error) {
			return false, errors.New("database error")
		},
	}

	err := CreateSeedAdmin(mockStore, "admin", "admin@example.com", "AdminPass123!")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateSeedAdmin_ReturnsErrorWhenCreateUserFails(t *testing.T) {
	mockStore := &mocks.MockAuthStore{
		UserExistsFunc: func(username string) (bool, error) {
			return false, nil
		},
		CreateUserFunc: func(username, email, passwordHash string, isAdmin bool) error {
			return errors.New("failed to create user")
		},
	}

	err := CreateSeedAdmin(mockStore, "admin", "admin@example.com", "AdminPass123!")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
