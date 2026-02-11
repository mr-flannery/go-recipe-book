package models

import (
	"encoding/base64"
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
