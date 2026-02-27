package models

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestRecipe_ImageBase64_EncodesImageDataToBase64(t *testing.T) {
	t.Run("returns empty string when image is nil", func(t *testing.T) {
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

	t.Run("returns empty string when image is empty slice", func(t *testing.T) {
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

	t.Run("returns base64 encoded string when image has data", func(t *testing.T) {
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

	t.Run("correctly encodes binary data like PNG header", func(t *testing.T) {
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

func TestRecipe_TotalTime_ReturnsSumOfPrepAndCookTime(t *testing.T) {
	tests := []struct {
		name     string
		prepTime int
		cookTime int
		expected int
	}{
		{"both zero", 0, 0, 0},
		{"only prep time", 10, 0, 10},
		{"only cook time", 0, 20, 20},
		{"both set", 15, 30, 45},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recipe := Recipe{PrepTime: tt.prepTime, CookTime: tt.cookTime}
			if got := recipe.TotalTime(); got != tt.expected {
				t.Errorf("TotalTime() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestRecipe_IngredientCount_CountsIngredientsFromMarkdown(t *testing.T) {
	tests := []struct {
		name          string
		ingredientsMD string
		expected      int
	}{
		{"empty string", "", 0},
		{"single dash item", "- 1 cup flour", 1},
		{"multiple dash items", "- 1 cup flour\n- 2 eggs\n- 1 tsp salt", 3},
		{"asterisk items", "* 1 cup flour\n* 2 eggs", 2},
		{"numbered items", "1. 1 cup flour\n2. 2 eggs\n3. 1 tsp salt", 3},
		{"mixed format", "- 1 cup flour\n* 2 eggs\n3. 1 tsp salt", 3},
		{"with blank lines", "- 1 cup flour\n\n- 2 eggs\n\n- 1 tsp salt", 3},
		{"with header", "## Ingredients\n- 1 cup flour\n- 2 eggs", 2},
		{"plain text no numbers", "flour\neggs\nsugar", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recipe := Recipe{IngredientsMD: tt.ingredientsMD}
			if got := recipe.IngredientCount(); got != tt.expected {
				t.Errorf("IngredientCount() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestRecipe_Summary_ReturnsFirstLineOfInstructions(t *testing.T) {
	tests := []struct {
		name           string
		instructionsMD string
		wantContains   string
		maxLen         int
	}{
		{"empty string", "", "", 0},
		{"short first line", "Mix all ingredients together.", "Mix all ingredients together.", 80},
		{"strips leading hash", "# Instructions\nMix ingredients.", "Instructions", 80},
		{"strips leading dash", "- Mix the flour first.", "Mix the flour first.", 80},
		{"strips numbered prefix", "1. Preheat the oven to 350F.", "Preheat the oven to 350F.", 80},
		{"truncates long line", strings.Repeat("a", 100), "...", 83},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recipe := Recipe{InstructionsMD: tt.instructionsMD}
			got := recipe.Summary()

			if tt.wantContains != "" && !strings.Contains(got, tt.wantContains) {
				t.Errorf("Summary() = %q, want to contain %q", got, tt.wantContains)
			}

			if tt.maxLen > 0 && len(got) > tt.maxLen {
				t.Errorf("Summary() length = %d, want <= %d", len(got), tt.maxLen)
			}
		})
	}
}
