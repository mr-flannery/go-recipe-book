package models

import (
	"encoding/base64"
	"testing"
)

func TestRecipe_ImageBase64(t *testing.T) {
	t.Run("empty image returns empty string", func(t *testing.T) {
		recipe := Recipe{
			ID:    1,
			Title: "Test Recipe",
			Image: nil,
		}

		result := recipe.ImageBase64()
		if result != "" {
			t.Errorf("expected empty string, got '%s'", result)
		}
	})

	t.Run("empty slice returns empty string", func(t *testing.T) {
		recipe := Recipe{
			ID:    1,
			Title: "Test Recipe",
			Image: []byte{},
		}

		result := recipe.ImageBase64()
		if result != "" {
			t.Errorf("expected empty string, got '%s'", result)
		}
	})

	t.Run("image data returns base64 string", func(t *testing.T) {
		imageData := []byte("Hello World")
		recipe := Recipe{
			ID:    1,
			Title: "Test Recipe",
			Image: imageData,
		}

		result := recipe.ImageBase64()
		expected := base64.StdEncoding.EncodeToString(imageData)

		if result != expected {
			t.Errorf("expected '%s', got '%s'", expected, result)
		}
	})

	t.Run("binary data encodes correctly", func(t *testing.T) {
		imageData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A} // PNG header
		recipe := Recipe{
			ID:    1,
			Title: "Test Recipe",
			Image: imageData,
		}

		result := recipe.ImageBase64()

		decoded, err := base64.StdEncoding.DecodeString(result)
		if err != nil {
			t.Fatalf("failed to decode result: %v", err)
		}

		if string(decoded) != string(imageData) {
			t.Error("decoded data doesn't match original")
		}
	})
}

func TestFilterParams(t *testing.T) {
	t.Run("default values", func(t *testing.T) {
		params := FilterParams{}

		if params.Search != "" {
			t.Error("expected empty search")
		}
		if params.CaloriesOp != "" {
			t.Error("expected empty calories op")
		}
		if params.CaloriesValue != 0 {
			t.Error("expected zero calories value")
		}
		if len(params.Tags) != 0 {
			t.Error("expected empty tags slice")
		}
	})

	t.Run("can set all fields", func(t *testing.T) {
		params := FilterParams{
			Search:        "pasta",
			CaloriesOp:    "lt",
			CaloriesValue: 500,
			PrepTimeOp:    "lte",
			PrepTimeValue: 30,
			CookTimeOp:    "gt",
			CookTimeValue: 15,
			Tags:          []string{"italian", "dinner"},
			UserID:        42,
			UserTags:      []string{"favorite"},
		}

		if params.Search != "pasta" {
			t.Errorf("expected search 'pasta', got '%s'", params.Search)
		}
		if params.CaloriesOp != "lt" {
			t.Errorf("expected calories op 'lt', got '%s'", params.CaloriesOp)
		}
		if params.CaloriesValue != 500 {
			t.Errorf("expected calories value 500, got %d", params.CaloriesValue)
		}
		if len(params.Tags) != 2 {
			t.Errorf("expected 2 tags, got %d", len(params.Tags))
		}
		if params.UserID != 42 {
			t.Errorf("expected user ID 42, got %d", params.UserID)
		}
	})
}

func TestTag(t *testing.T) {
	tag := Tag{
		ID:   1,
		Name: "breakfast",
	}

	if tag.ID != 1 {
		t.Errorf("expected ID 1, got %d", tag.ID)
	}
	if tag.Name != "breakfast" {
		t.Errorf("expected name 'breakfast', got '%s'", tag.Name)
	}
}

func TestUserTag(t *testing.T) {
	userTag := UserTag{
		ID:       10,
		UserID:   1,
		RecipeID: 5,
		Name:     "must-try",
	}

	if userTag.ID != 10 {
		t.Errorf("expected ID 10, got %d", userTag.ID)
	}
	if userTag.UserID != 1 {
		t.Errorf("expected UserID 1, got %d", userTag.UserID)
	}
	if userTag.RecipeID != 5 {
		t.Errorf("expected RecipeID 5, got %d", userTag.RecipeID)
	}
	if userTag.Name != "must-try" {
		t.Errorf("expected name 'must-try', got '%s'", userTag.Name)
	}
}

func TestComment(t *testing.T) {
	comment := Comment{
		ID:        1,
		RecipeID:  5,
		AuthorID:  2,
		ContentMD: "Great recipe!",
	}

	if comment.ID != 1 {
		t.Errorf("expected ID 1, got %d", comment.ID)
	}
	if comment.RecipeID != 5 {
		t.Errorf("expected RecipeID 5, got %d", comment.RecipeID)
	}
	if comment.AuthorID != 2 {
		t.Errorf("expected AuthorID 2, got %d", comment.AuthorID)
	}
	if comment.ContentMD != "Great recipe!" {
		t.Errorf("expected content 'Great recipe!', got '%s'", comment.ContentMD)
	}
}

func TestUser(t *testing.T) {
	user := User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		IsAdmin:  true,
		IsActive: true,
	}

	if user.ID != 1 {
		t.Errorf("expected ID 1, got %d", user.ID)
	}
	if user.Username != "testuser" {
		t.Errorf("expected username 'testuser', got '%s'", user.Username)
	}
	if !user.IsAdmin {
		t.Error("expected IsAdmin to be true")
	}
	if !user.IsActive {
		t.Error("expected IsActive to be true")
	}
}

func TestRecipe(t *testing.T) {
	parentID := 10
	recipe := Recipe{
		ID:             1,
		Title:          "Test Recipe",
		IngredientsMD:  "- 1 cup flour",
		InstructionsMD: "Mix and bake",
		PrepTime:       15,
		CookTime:       30,
		Calories:       250,
		AuthorID:       1,
		ParentID:       &parentID,
		Tags:           []Tag{{ID: 1, Name: "breakfast"}},
	}

	if recipe.ID != 1 {
		t.Errorf("expected ID 1, got %d", recipe.ID)
	}
	if recipe.Title != "Test Recipe" {
		t.Errorf("expected title 'Test Recipe', got '%s'", recipe.Title)
	}
	if recipe.PrepTime != 15 {
		t.Errorf("expected PrepTime 15, got %d", recipe.PrepTime)
	}
	if recipe.CookTime != 30 {
		t.Errorf("expected CookTime 30, got %d", recipe.CookTime)
	}
	if recipe.ParentID == nil || *recipe.ParentID != 10 {
		t.Error("expected ParentID to be 10")
	}
	if len(recipe.Tags) != 1 {
		t.Errorf("expected 1 tag, got %d", len(recipe.Tags))
	}
}

func TestProposedChange(t *testing.T) {
	change := ProposedChange{
		ID:             1,
		RecipeID:       5,
		ProposerID:     2,
		Title:          "Updated Recipe",
		IngredientsMD:  "- 2 cups flour",
		InstructionsMD: "Mix well",
		Status:         "pending",
	}

	if change.ID != 1 {
		t.Errorf("expected ID 1, got %d", change.ID)
	}
	if change.RecipeID != 5 {
		t.Errorf("expected RecipeID 5, got %d", change.RecipeID)
	}
	if change.Status != "pending" {
		t.Errorf("expected status 'pending', got '%s'", change.Status)
	}
}
