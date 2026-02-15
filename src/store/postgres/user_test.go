package postgres

import (
	"testing"

	"github.com/mr-flannery/go-recipe-book/src/testutil"
)

func TestUserStore_GetUsernameByID_ReturnsUsernameWhenExists(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	store := NewUserStore(testDB.DB)

	username, err := store.GetUsernameByID(userID)
	if err != nil {
		t.Fatalf("failed to get username by ID: %v", err)
	}

	if username != "testuser" {
		t.Errorf("expected username 'testuser', got '%s'", username)
	}
}

func TestUserStore_GetUsernameByID_ReturnsErrorWhenNotFound(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	store := NewUserStore(testDB.DB)

	_, err := store.GetUsernameByID(99999)
	if err == nil {
		t.Error("expected error for non-existent user")
	}
}
