package postgres

import (
	"strconv"
	"testing"

	"github.com/mr-flannery/go-recipe-book/src/models"
	"github.com/mr-flannery/go-recipe-book/src/testutil"
)

func TestCommentStore_Save_CreatesComment(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	recipeID := testDB.SeedRecipe(t, "Test Recipe", "- flour", "Mix it", userID)
	store := NewCommentStore(testDB.DB)

	comment := models.Comment{
		RecipeID:  recipeID,
		AuthorID:  userID,
		ContentMD: "This is a test comment",
	}

	err := store.Save(comment)
	if err != nil {
		t.Fatalf("failed to save comment: %v", err)
	}

	comments, _ := store.GetByRecipeID(strconv.Itoa(recipeID))
	if len(comments) != 1 {
		t.Errorf("expected 1 comment, got %d", len(comments))
	}
}

func TestCommentStore_GetByRecipeID_ReturnsCommentsForRecipe(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	recipeID := testDB.SeedRecipe(t, "Test Recipe", "- flour", "Mix it", userID)
	testDB.SeedComment(t, recipeID, userID, "Comment 1")
	testDB.SeedComment(t, recipeID, userID, "Comment 2")
	store := NewCommentStore(testDB.DB)

	comments, err := store.GetByRecipeID(strconv.Itoa(recipeID))
	if err != nil {
		t.Fatalf("failed to get comments: %v", err)
	}

	if len(comments) != 2 {
		t.Errorf("expected 2 comments, got %d", len(comments))
	}
}

func TestCommentStore_GetByRecipeID_ReturnsEmptyForNoComments(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	recipeID := testDB.SeedRecipe(t, "Test Recipe", "- flour", "Mix it", userID)
	store := NewCommentStore(testDB.DB)

	comments, err := store.GetByRecipeID(strconv.Itoa(recipeID))
	if err != nil {
		t.Fatalf("failed to get comments: %v", err)
	}

	if len(comments) != 0 {
		t.Errorf("expected 0 comments, got %d", len(comments))
	}
}

func TestCommentStore_GetByID_ReturnsCommentWhenExists(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	recipeID := testDB.SeedRecipe(t, "Test Recipe", "- flour", "Mix it", userID)
	commentID := testDB.SeedComment(t, recipeID, userID, "Test Comment")
	store := NewCommentStore(testDB.DB)

	comment, err := store.GetByID(commentID)
	if err != nil {
		t.Fatalf("failed to get comment by ID: %v", err)
	}

	if comment.ContentMD != "Test Comment" {
		t.Errorf("expected content 'Test Comment', got '%s'", comment.ContentMD)
	}
}

func TestCommentStore_GetByID_ReturnsErrorWhenNotFound(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	store := NewCommentStore(testDB.DB)

	_, err := store.GetByID(99999)
	if err == nil {
		t.Error("expected error for non-existent comment")
	}
}

func TestCommentStore_GetLatestByUserAndRecipe_ReturnsLatestComment(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	recipeID := testDB.SeedRecipe(t, "Test Recipe", "- flour", "Mix it", userID)
	testDB.SeedComment(t, recipeID, userID, "First comment")
	testDB.SeedComment(t, recipeID, userID, "Latest comment")
	store := NewCommentStore(testDB.DB)

	comment, err := store.GetLatestByUserAndRecipe(userID, recipeID)
	if err != nil {
		t.Fatalf("failed to get latest comment: %v", err)
	}

	if comment.ContentMD != "Latest comment" {
		t.Errorf("expected content 'Latest comment', got '%s'", comment.ContentMD)
	}
}

func TestCommentStore_Update_ModifiesCommentContent(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	recipeID := testDB.SeedRecipe(t, "Test Recipe", "- flour", "Mix it", userID)
	commentID := testDB.SeedComment(t, recipeID, userID, "Original content")
	store := NewCommentStore(testDB.DB)

	err := store.Update(commentID, "Updated content")
	if err != nil {
		t.Fatalf("failed to update comment: %v", err)
	}

	updated, _ := store.GetByID(commentID)
	if updated.ContentMD != "Updated content" {
		t.Errorf("expected content 'Updated content', got '%s'", updated.ContentMD)
	}
}

func TestCommentStore_Delete_RemovesComment(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	recipeID := testDB.SeedRecipe(t, "Test Recipe", "- flour", "Mix it", userID)
	commentID := testDB.SeedComment(t, recipeID, userID, "To delete")
	store := NewCommentStore(testDB.DB)

	err := store.Delete(commentID)
	if err != nil {
		t.Fatalf("failed to delete comment: %v", err)
	}

	_, err = store.GetByID(commentID)
	if err == nil {
		t.Error("expected error after deleting comment")
	}
}
