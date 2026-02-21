package postgres

import (
	"strconv"
	"testing"

	"github.com/mr-flannery/go-recipe-book/src/models"
	"github.com/mr-flannery/go-recipe-book/src/testutil"
)

func TestRecipeStore_Save_ReturnsIDWhenRecipeIsSaved(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpass", false)
	store := NewRecipeStore(testDB.DB)

	recipe := models.Recipe{
		Title:          "Test Recipe",
		IngredientsMD:  "- 1 cup flour",
		InstructionsMD: "Mix everything",
		PrepTime:       10,
		CookTime:       20,
		Calories:       300,
		AuthorID:       userID,
	}

	id, err := store.Save(recipe)
	if err != nil {
		t.Fatalf("failed to save recipe: %v", err)
	}

	if id == 0 {
		t.Error("expected non-zero recipe ID")
	}
}

func TestRecipeStore_GetByID_ReturnsRecipeWhenExists(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpass", false)
	recipeID := testDB.SeedRecipe(t, "Test Recipe", "- flour", "Mix it", userID)
	store := NewRecipeStore(testDB.DB)

	recipe, err := store.GetByID(itoa(recipeID))
	if err != nil {
		t.Fatalf("failed to get recipe: %v", err)
	}

	if recipe.Title != "Test Recipe" {
		t.Errorf("expected title 'Test Recipe', got '%s'", recipe.Title)
	}

	if recipe.AuthorID != userID {
		t.Errorf("expected author ID %d, got %d", userID, recipe.AuthorID)
	}
}

func TestRecipeStore_GetByID_ReturnsErrorWhenNotFound(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	store := NewRecipeStore(testDB.DB)

	_, err := store.GetByID("99999")
	if err == nil {
		t.Error("expected error for non-existent recipe")
	}
}

func TestRecipeStore_Update_ModifiesExistingRecipe(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpass", false)
	recipeID := testDB.SeedRecipe(t, "Original Title", "- flour", "Mix it", userID)
	store := NewRecipeStore(testDB.DB)

	recipe, _ := store.GetByID(itoa(recipeID))
	recipe.Title = "Updated Title"
	recipe.Calories = 500

	err := store.Update(recipe)
	if err != nil {
		t.Fatalf("failed to update recipe: %v", err)
	}

	updated, _ := store.GetByID(itoa(recipeID))
	if updated.Title != "Updated Title" {
		t.Errorf("expected title 'Updated Title', got '%s'", updated.Title)
	}
	if updated.Calories != 500 {
		t.Errorf("expected calories 500, got %d", updated.Calories)
	}
}

func TestRecipeStore_Delete_RemovesRecipe(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpass", false)
	recipeID := testDB.SeedRecipe(t, "To Delete", "- flour", "Mix it", userID)
	store := NewRecipeStore(testDB.DB)

	err := store.Delete(itoa(recipeID))
	if err != nil {
		t.Fatalf("failed to delete recipe: %v", err)
	}

	_, err = store.GetByID(itoa(recipeID))
	if err == nil {
		t.Error("expected error after deleting recipe")
	}
}

func TestRecipeStore_GetAll_ReturnsAllRecipes(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpass", false)
	testDB.SeedRecipe(t, "Recipe 1", "- flour", "Mix it", userID)
	testDB.SeedRecipe(t, "Recipe 2", "- sugar", "Stir it", userID)
	store := NewRecipeStore(testDB.DB)

	recipes, err := store.GetAll()
	if err != nil {
		t.Fatalf("failed to get all recipes: %v", err)
	}

	if len(recipes) != 2 {
		t.Errorf("expected 2 recipes, got %d", len(recipes))
	}
}

func TestRecipeStore_GetFiltered_FiltersBySearch(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpass", false)
	testDB.SeedRecipe(t, "Chocolate Cake", "- cocoa", "Bake it", userID)
	testDB.SeedRecipe(t, "Vanilla Ice Cream", "- vanilla", "Freeze it", userID)
	store := NewRecipeStore(testDB.DB)

	params := models.FilterParams{Search: "chocolate"}
	recipes, err := store.GetFiltered(params)
	if err != nil {
		t.Fatalf("failed to filter recipes: %v", err)
	}

	if len(recipes) != 1 {
		t.Errorf("expected 1 recipe, got %d", len(recipes))
	}

	if len(recipes) > 0 && recipes[0].Title != "Chocolate Cake" {
		t.Errorf("expected 'Chocolate Cake', got '%s'", recipes[0].Title)
	}
}

func TestRecipeStore_GetFiltered_FiltersByCalories(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpass", false)

	var lowCalID int
	err := testDB.DB.QueryRow(`
		INSERT INTO recipes (title, ingredients_md, instructions_md, author_id, prep_time, cook_time, calories, created_at, updated_at)
		VALUES ('Low Cal', '- lettuce', 'Toss it', $1, 5, 0, 100, NOW(), NOW())
		RETURNING id
	`, userID).Scan(&lowCalID)
	if err != nil {
		t.Fatalf("failed to seed low cal recipe: %v", err)
	}

	var highCalID int
	err = testDB.DB.QueryRow(`
		INSERT INTO recipes (title, ingredients_md, instructions_md, author_id, prep_time, cook_time, calories, created_at, updated_at)
		VALUES ('High Cal', '- butter', 'Fry it', $1, 10, 30, 800, NOW(), NOW())
		RETURNING id
	`, userID).Scan(&highCalID)
	if err != nil {
		t.Fatalf("failed to seed high cal recipe: %v", err)
	}

	store := NewRecipeStore(testDB.DB)

	params := models.FilterParams{CaloriesOp: "lt", CaloriesValue: 500}
	recipes, err := store.GetFiltered(params)
	if err != nil {
		t.Fatalf("failed to filter recipes: %v", err)
	}

	if len(recipes) != 1 {
		t.Errorf("expected 1 recipe with calories < 500, got %d", len(recipes))
	}
}

func TestRecipeStore_GetFiltered_FiltersByTags(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpass", false)
	recipeID := testDB.SeedRecipe(t, "Tagged Recipe", "- flour", "Mix it", userID)
	testDB.SeedRecipe(t, "Untagged Recipe", "- sugar", "Stir it", userID)
	tagID := testDB.SeedTag(t, "dessert")
	testDB.SeedRecipeTag(t, recipeID, tagID)
	store := NewRecipeStore(testDB.DB)

	params := models.FilterParams{Tags: []string{"dessert"}}
	recipes, err := store.GetFiltered(params)
	if err != nil {
		t.Fatalf("failed to filter recipes: %v", err)
	}

	if len(recipes) != 1 {
		t.Errorf("expected 1 recipe with tag, got %d", len(recipes))
	}

	if len(recipes) > 0 && recipes[0].Title != "Tagged Recipe" {
		t.Errorf("expected 'Tagged Recipe', got '%s'", recipes[0].Title)
	}
}

func TestRecipeStore_GetFiltered_AppliesPagination(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpass", false)
	for i := 0; i < 5; i++ {
		testDB.SeedRecipe(t, "Recipe", "- flour", "Mix it", userID)
	}
	store := NewRecipeStore(testDB.DB)

	params := models.FilterParams{Limit: 2, Offset: 2}
	recipes, err := store.GetFiltered(params)
	if err != nil {
		t.Fatalf("failed to filter recipes: %v", err)
	}

	if len(recipes) != 2 {
		t.Errorf("expected 2 recipes with pagination, got %d", len(recipes))
	}
}

func TestRecipeStore_CountFiltered_ReturnsTotalCount(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpass", false)
	testDB.SeedRecipe(t, "Recipe 1", "- flour", "Mix it", userID)
	testDB.SeedRecipe(t, "Recipe 2", "- sugar", "Stir it", userID)
	testDB.SeedRecipe(t, "Recipe 3", "- eggs", "Beat it", userID)
	store := NewRecipeStore(testDB.DB)

	count, err := store.CountFiltered(models.FilterParams{})
	if err != nil {
		t.Fatalf("failed to count recipes: %v", err)
	}

	if count != 3 {
		t.Errorf("expected count 3, got %d", count)
	}
}

func TestRecipeStore_GetRandomID_ReturnsValidID(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpass", false)
	testDB.SeedRecipe(t, "Recipe 1", "- flour", "Mix it", userID)
	store := NewRecipeStore(testDB.DB)

	id, err := store.GetRandomID()
	if err != nil {
		t.Fatalf("failed to get random ID: %v", err)
	}

	if id == 0 {
		t.Error("expected non-zero random ID")
	}
}

func TestRecipeStore_GetRandomID_ReturnsErrorWhenNoRecipes(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	store := NewRecipeStore(testDB.DB)

	_, err := store.GetRandomID()
	if err == nil {
		t.Error("expected error when no recipes exist")
	}
}

func TestRecipeStore_SearchByTitle_ReturnsMatchingRecipes(t *testing.T) {

	testDB := testutil.GetTestDatabase(t)

	userID := testDB.SeedUser(t, "testuser", "test@example.com", "hashedpass", false)
	testDB.SeedRecipe(t, "Chocolate Cake", "- cocoa", "Bake it", userID)
	testDB.SeedRecipe(t, "Carrot Cake", "- carrots", "Bake it too", userID)
	testDB.SeedRecipe(t, "Vanilla Ice Cream", "- vanilla", "Freeze it", userID)
	store := NewRecipeStore(testDB.DB)

	t.Run("finds recipes matching query", func(t *testing.T) {
		results, err := store.SearchByTitle("cake", 10)
		if err != nil {
			t.Fatalf("failed to search recipes: %v", err)
		}

		if len(results) != 2 {
			t.Errorf("expected 2 recipes matching 'cake', got %d", len(results))
		}
	})

	t.Run("search is case-insensitive", func(t *testing.T) {
		results, err := store.SearchByTitle("CHOCOLATE", 10)
		if err != nil {
			t.Fatalf("failed to search recipes: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("expected 1 recipe matching 'CHOCOLATE', got %d", len(results))
		}
		if len(results) > 0 && results[0].Title != "Chocolate Cake" {
			t.Errorf("expected 'Chocolate Cake', got '%s'", results[0].Title)
		}
	})

	t.Run("respects limit parameter", func(t *testing.T) {
		results, err := store.SearchByTitle("cake", 1)
		if err != nil {
			t.Fatalf("failed to search recipes: %v", err)
		}

		if len(results) != 1 {
			t.Errorf("expected 1 recipe with limit 1, got %d", len(results))
		}
	})

	t.Run("returns empty slice when no matches", func(t *testing.T) {
		results, err := store.SearchByTitle("pizza", 10)
		if err != nil {
			t.Fatalf("failed to search recipes: %v", err)
		}

		if len(results) != 0 {
			t.Errorf("expected 0 recipes matching 'pizza', got %d", len(results))
		}
	})
}

func itoa(i int) string {
	return strconv.Itoa(i)
}
