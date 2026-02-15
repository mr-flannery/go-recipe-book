package postgres

import (
	"testing"

	"github.com/mr-flannery/go-recipe-book/src/testutil"
)

func TestTagStore_GetOrCreate_CreatesNewTag(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	store := NewTagStore(testDB.DB)

	tag, err := store.GetOrCreate("dessert")
	if err != nil {
		t.Fatalf("failed to create tag: %v", err)
	}

	if tag.Name != "dessert" {
		t.Errorf("expected tag name 'dessert', got '%s'", tag.Name)
	}
	if tag.ID == 0 {
		t.Error("expected non-zero tag ID")
	}
}

func TestTagStore_GetOrCreate_ReturnsExistingTag(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	existingTagID := testDB.SeedTag(t, "dessert")
	store := NewTagStore(testDB.DB)

	tag, err := store.GetOrCreate("dessert")
	if err != nil {
		t.Fatalf("failed to get existing tag: %v", err)
	}

	if tag.ID != existingTagID {
		t.Errorf("expected existing tag ID %d, got %d", existingTagID, tag.ID)
	}
}

func TestTagStore_GetOrCreate_NormalizesTagName(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	store := NewTagStore(testDB.DB)

	tag, err := store.GetOrCreate("  DESSERT  ")
	if err != nil {
		t.Fatalf("failed to create tag: %v", err)
	}

	if tag.Name != "dessert" {
		t.Errorf("expected normalized tag name 'dessert', got '%s'", tag.Name)
	}
}

func TestTagStore_GetOrCreate_ReturnsErrorForEmptyName(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	store := NewTagStore(testDB.DB)

	_, err := store.GetOrCreate("   ")
	if err == nil {
		t.Error("expected error for empty tag name")
	}
}

func TestTagStore_Search_FindsMatchingTags(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	testDB.SeedTag(t, "dessert")
	testDB.SeedTag(t, "desert")
	testDB.SeedTag(t, "main course")
	store := NewTagStore(testDB.DB)

	tags, err := store.Search("des")
	if err != nil {
		t.Fatalf("failed to search tags: %v", err)
	}

	if len(tags) != 2 {
		t.Errorf("expected 2 matching tags, got %d", len(tags))
	}
}

func TestTagStore_Search_ReturnsEmptyForNoMatches(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	testDB.SeedTag(t, "dessert")
	store := NewTagStore(testDB.DB)

	tags, err := store.Search("xyz")
	if err != nil {
		t.Fatalf("failed to search tags: %v", err)
	}

	if len(tags) != 0 {
		t.Errorf("expected 0 matching tags, got %d", len(tags))
	}
}

func TestTagStore_GetByRecipeID_ReturnsTagsForRecipe(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	recipeID := testDB.SeedRecipe(t, "Test Recipe", "- flour", "Mix it", userID)
	tagID1 := testDB.SeedTag(t, "dessert")
	tagID2 := testDB.SeedTag(t, "chocolate")
	testDB.SeedRecipeTag(t, recipeID, tagID1)
	testDB.SeedRecipeTag(t, recipeID, tagID2)
	store := NewTagStore(testDB.DB)

	tags, err := store.GetByRecipeID(recipeID)
	if err != nil {
		t.Fatalf("failed to get tags by recipe ID: %v", err)
	}

	if len(tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(tags))
	}
}

func TestTagStore_GetForRecipes_ReturnsTagsForMultipleRecipes(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	recipe1ID := testDB.SeedRecipe(t, "Recipe 1", "- flour", "Mix it", userID)
	recipe2ID := testDB.SeedRecipe(t, "Recipe 2", "- sugar", "Stir it", userID)
	tagID1 := testDB.SeedTag(t, "dessert")
	tagID2 := testDB.SeedTag(t, "main")
	testDB.SeedRecipeTag(t, recipe1ID, tagID1)
	testDB.SeedRecipeTag(t, recipe2ID, tagID2)
	store := NewTagStore(testDB.DB)

	tagsMap, err := store.GetForRecipes([]int{recipe1ID, recipe2ID})
	if err != nil {
		t.Fatalf("failed to get tags for recipes: %v", err)
	}

	if len(tagsMap[recipe1ID]) != 1 {
		t.Errorf("expected 1 tag for recipe 1, got %d", len(tagsMap[recipe1ID]))
	}
	if len(tagsMap[recipe2ID]) != 1 {
		t.Errorf("expected 1 tag for recipe 2, got %d", len(tagsMap[recipe2ID]))
	}
}

func TestTagStore_GetForRecipes_ReturnsEmptyMapForEmptyInput(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	store := NewTagStore(testDB.DB)

	tagsMap, err := store.GetForRecipes([]int{})
	if err != nil {
		t.Fatalf("failed to get tags for empty recipes: %v", err)
	}

	if len(tagsMap) != 0 {
		t.Errorf("expected empty map, got %d entries", len(tagsMap))
	}
}

func TestTagStore_AddToRecipe_AddsTagToRecipe(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	recipeID := testDB.SeedRecipe(t, "Test Recipe", "- flour", "Mix it", userID)
	tagID := testDB.SeedTag(t, "dessert")
	store := NewTagStore(testDB.DB)

	err := store.AddToRecipe(recipeID, tagID)
	if err != nil {
		t.Fatalf("failed to add tag to recipe: %v", err)
	}

	tags, _ := store.GetByRecipeID(recipeID)
	if len(tags) != 1 {
		t.Errorf("expected 1 tag, got %d", len(tags))
	}
}

func TestTagStore_AddToRecipe_HandlesConflictGracefully(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	recipeID := testDB.SeedRecipe(t, "Test Recipe", "- flour", "Mix it", userID)
	tagID := testDB.SeedTag(t, "dessert")
	testDB.SeedRecipeTag(t, recipeID, tagID)
	store := NewTagStore(testDB.DB)

	err := store.AddToRecipe(recipeID, tagID)
	if err != nil {
		t.Fatalf("expected no error for duplicate add, got: %v", err)
	}
}

func TestTagStore_RemoveFromRecipe_RemovesTagFromRecipe(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	recipeID := testDB.SeedRecipe(t, "Test Recipe", "- flour", "Mix it", userID)
	tagID := testDB.SeedTag(t, "dessert")
	testDB.SeedRecipeTag(t, recipeID, tagID)
	store := NewTagStore(testDB.DB)

	err := store.RemoveFromRecipe(recipeID, tagID)
	if err != nil {
		t.Fatalf("failed to remove tag from recipe: %v", err)
	}

	tags, _ := store.GetByRecipeID(recipeID)
	if len(tags) != 0 {
		t.Errorf("expected 0 tags after removal, got %d", len(tags))
	}
}

func TestTagStore_SetRecipeTags_ReplacesAllTags(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	recipeID := testDB.SeedRecipe(t, "Test Recipe", "- flour", "Mix it", userID)
	tagID := testDB.SeedTag(t, "old-tag")
	testDB.SeedRecipeTag(t, recipeID, tagID)
	store := NewTagStore(testDB.DB)

	err := store.SetRecipeTags(recipeID, []string{"new-tag-1", "new-tag-2"})
	if err != nil {
		t.Fatalf("failed to set recipe tags: %v", err)
	}

	tags, _ := store.GetByRecipeID(recipeID)
	if len(tags) != 2 {
		t.Errorf("expected 2 tags, got %d", len(tags))
	}

	tagNames := make(map[string]bool)
	for _, tag := range tags {
		tagNames[tag.Name] = true
	}
	if tagNames["old-tag"] {
		t.Error("old tag should have been removed")
	}
	if !tagNames["new-tag-1"] || !tagNames["new-tag-2"] {
		t.Error("new tags should have been added")
	}
}

func TestTagStore_SetRecipeTags_SkipsEmptyTagNames(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpassword", false)
	recipeID := testDB.SeedRecipe(t, "Test Recipe", "- flour", "Mix it", userID)
	store := NewTagStore(testDB.DB)

	err := store.SetRecipeTags(recipeID, []string{"valid", "", "   "})
	if err != nil {
		t.Fatalf("failed to set recipe tags: %v", err)
	}

	tags, _ := store.GetByRecipeID(recipeID)
	if len(tags) != 1 {
		t.Errorf("expected 1 tag (empty ones skipped), got %d", len(tags))
	}
}
