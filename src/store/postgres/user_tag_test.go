package postgres

import (
	"testing"

	"github.com/mr-flannery/go-recipe-book/src/testutil"
)

func TestUserTagStore_GetOrCreate_CreatesNewUserTag(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	recipeID := testDB.SeedRecipe(t, "Test Recipe", "- flour", "Mix it", userID)
	store := NewUserTagStore(testDB.DB)

	tag, err := store.GetOrCreate(userID, recipeID, "favorite")
	if err != nil {
		t.Fatalf("failed to create user tag: %v", err)
	}

	if tag.Name != "favorite" {
		t.Errorf("expected tag name 'favorite', got '%s'", tag.Name)
	}
	if tag.UserID != userID {
		t.Errorf("expected user ID %d, got %d", userID, tag.UserID)
	}
	if tag.RecipeID != recipeID {
		t.Errorf("expected recipe ID %d, got %d", recipeID, tag.RecipeID)
	}
}

func TestUserTagStore_GetOrCreate_ReturnsExistingUserTag(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	recipeID := testDB.SeedRecipe(t, "Test Recipe", "- flour", "Mix it", userID)
	existingTagID := testDB.SeedUserTag(t, userID, recipeID, "favorite")
	store := NewUserTagStore(testDB.DB)

	tag, err := store.GetOrCreate(userID, recipeID, "favorite")
	if err != nil {
		t.Fatalf("failed to get existing user tag: %v", err)
	}

	if tag.ID != existingTagID {
		t.Errorf("expected existing tag ID %d, got %d", existingTagID, tag.ID)
	}
}

func TestUserTagStore_GetOrCreate_NormalizesTagName(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	recipeID := testDB.SeedRecipe(t, "Test Recipe", "- flour", "Mix it", userID)
	store := NewUserTagStore(testDB.DB)

	tag, err := store.GetOrCreate(userID, recipeID, "  FAVORITE  ")
	if err != nil {
		t.Fatalf("failed to create user tag: %v", err)
	}

	if tag.Name != "favorite" {
		t.Errorf("expected normalized tag name 'favorite', got '%s'", tag.Name)
	}
}

func TestUserTagStore_GetOrCreate_ReturnsErrorForEmptyName(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	recipeID := testDB.SeedRecipe(t, "Test Recipe", "- flour", "Mix it", userID)
	store := NewUserTagStore(testDB.DB)

	_, err := store.GetOrCreate(userID, recipeID, "   ")
	if err == nil {
		t.Error("expected error for empty tag name")
	}
}

func TestUserTagStore_Search_FindsMatchingTags(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	recipeID := testDB.SeedRecipe(t, "Test Recipe", "- flour", "Mix it", userID)
	testDB.SeedUserTag(t, userID, recipeID, "favorite")
	testDB.SeedUserTag(t, userID, recipeID, "fast")
	testDB.SeedUserTag(t, userID, recipeID, "healthy")
	store := NewUserTagStore(testDB.DB)

	tags, err := store.Search(userID, "fa")
	if err != nil {
		t.Fatalf("failed to search user tags: %v", err)
	}

	if len(tags) != 2 {
		t.Errorf("expected 2 matching tags (favorite, fast), got %d", len(tags))
	}
}

func TestUserTagStore_Search_ReturnsDistinctTags(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	recipe1ID := testDB.SeedRecipe(t, "Recipe 1", "- flour", "Mix it", userID)
	recipe2ID := testDB.SeedRecipe(t, "Recipe 2", "- sugar", "Stir it", userID)
	testDB.SeedUserTag(t, userID, recipe1ID, "favorite")
	testDB.SeedUserTag(t, userID, recipe2ID, "favorite")
	store := NewUserTagStore(testDB.DB)

	tags, err := store.Search(userID, "fav")
	if err != nil {
		t.Fatalf("failed to search user tags: %v", err)
	}

	if len(tags) != 1 {
		t.Errorf("expected 1 distinct tag, got %d", len(tags))
	}
}

func TestUserTagStore_GetByRecipeID_ReturnsUserTagsForRecipe(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	recipeID := testDB.SeedRecipe(t, "Test Recipe", "- flour", "Mix it", userID)
	testDB.SeedUserTag(t, userID, recipeID, "favorite")
	testDB.SeedUserTag(t, userID, recipeID, "to-try")
	store := NewUserTagStore(testDB.DB)

	tags, err := store.GetByRecipeID(userID, recipeID)
	if err != nil {
		t.Fatalf("failed to get user tags: %v", err)
	}

	if len(tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(tags))
	}
}

func TestUserTagStore_GetByRecipeID_ReturnsOnlyUserOwnTags(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	user1ID := testDB.SeedUser(t, "user1", "user1@example.com", "hashedpassword", false)
	user2ID := testDB.SeedUser(t, "user2", "user2@example.com", "hashedpassword", false)
	recipeID := testDB.SeedRecipe(t, "Test Recipe", "- flour", "Mix it", user1ID)
	testDB.SeedUserTag(t, user1ID, recipeID, "user1-tag")
	testDB.SeedUserTag(t, user2ID, recipeID, "user2-tag")
	store := NewUserTagStore(testDB.DB)

	tags, err := store.GetByRecipeID(user1ID, recipeID)
	if err != nil {
		t.Fatalf("failed to get user tags: %v", err)
	}

	if len(tags) != 1 {
		t.Errorf("expected 1 tag (only user1's), got %d", len(tags))
	}
	if tags[0].Name != "user1-tag" {
		t.Errorf("expected 'user1-tag', got '%s'", tags[0].Name)
	}
}

func TestUserTagStore_GetForRecipes_ReturnsTagsForMultipleRecipes(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	recipe1ID := testDB.SeedRecipe(t, "Recipe 1", "- flour", "Mix it", userID)
	recipe2ID := testDB.SeedRecipe(t, "Recipe 2", "- sugar", "Stir it", userID)
	testDB.SeedUserTag(t, userID, recipe1ID, "favorite")
	testDB.SeedUserTag(t, userID, recipe2ID, "to-try")
	store := NewUserTagStore(testDB.DB)

	tagsMap, err := store.GetForRecipes(userID, []int{recipe1ID, recipe2ID})
	if err != nil {
		t.Fatalf("failed to get user tags for recipes: %v", err)
	}

	if len(tagsMap[recipe1ID]) != 1 {
		t.Errorf("expected 1 tag for recipe 1, got %d", len(tagsMap[recipe1ID]))
	}
	if len(tagsMap[recipe2ID]) != 1 {
		t.Errorf("expected 1 tag for recipe 2, got %d", len(tagsMap[recipe2ID]))
	}
}

func TestUserTagStore_GetForRecipes_ReturnsEmptyMapForEmptyInput(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	store := NewUserTagStore(testDB.DB)

	tagsMap, err := store.GetForRecipes(userID, []int{})
	if err != nil {
		t.Fatalf("failed to get user tags for empty recipes: %v", err)
	}

	if len(tagsMap) != 0 {
		t.Errorf("expected empty map, got %d entries", len(tagsMap))
	}
}

func TestUserTagStore_Remove_RemovesUserTag(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	recipeID := testDB.SeedRecipe(t, "Test Recipe", "- flour", "Mix it", userID)
	tagID := testDB.SeedUserTag(t, userID, recipeID, "to-remove")
	store := NewUserTagStore(testDB.DB)

	err := store.Remove(userID, tagID)
	if err != nil {
		t.Fatalf("failed to remove user tag: %v", err)
	}

	tags, _ := store.GetByRecipeID(userID, recipeID)
	if len(tags) != 0 {
		t.Errorf("expected 0 tags after removal, got %d", len(tags))
	}
}

func TestUserTagStore_Remove_DoesNotRemoveOtherUserTags(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	user1ID := testDB.SeedUser(t, "user1", "user1@example.com", "hashedpassword", false)
	user2ID := testDB.SeedUser(t, "user2", "user2@example.com", "hashedpassword", false)
	recipeID := testDB.SeedRecipe(t, "Test Recipe", "- flour", "Mix it", user1ID)
	tagID := testDB.SeedUserTag(t, user1ID, recipeID, "user1-tag")
	store := NewUserTagStore(testDB.DB)

	err := store.Remove(user2ID, tagID)
	if err != nil {
		t.Fatalf("failed to remove user tag: %v", err)
	}

	tags, _ := store.GetByRecipeID(user1ID, recipeID)
	if len(tags) != 1 {
		t.Errorf("expected user1's tag to still exist, got %d tags", len(tags))
	}
}
