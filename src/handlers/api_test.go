package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mr-flannery/go-recipe-book/src/models"
	"github.com/mr-flannery/go-recipe-book/src/store/mocks"
)

func TestValidateRecipeRequest_ValidatesRequiredFieldsAndConstraints(t *testing.T) {
	tests := []struct {
		name    string
		req     APIRecipeRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: APIRecipeRequest{
				Title:          "Test Recipe",
				IngredientsMD:  "- 1 cup flour",
				InstructionsMD: "Mix and bake",
				PrepTime:       10,
				CookTime:       20,
				Calories:       300,
			},
			wantErr: false,
		},
		{
			name: "missing title",
			req: APIRecipeRequest{
				Title:          "",
				IngredientsMD:  "- 1 cup flour",
				InstructionsMD: "Mix and bake",
			},
			wantErr: true,
			errMsg:  "title is required",
		},
		{
			name: "whitespace only title",
			req: APIRecipeRequest{
				Title:          "   ",
				IngredientsMD:  "- 1 cup flour",
				InstructionsMD: "Mix and bake",
			},
			wantErr: true,
			errMsg:  "title is required",
		},
		{
			name: "missing ingredients",
			req: APIRecipeRequest{
				Title:          "Test Recipe",
				IngredientsMD:  "",
				InstructionsMD: "Mix and bake",
			},
			wantErr: true,
			errMsg:  "ingredients are required",
		},
		{
			name: "missing instructions",
			req: APIRecipeRequest{
				Title:          "Test Recipe",
				IngredientsMD:  "- 1 cup flour",
				InstructionsMD: "",
			},
			wantErr: true,
			errMsg:  "instructions are required",
		},
		{
			name: "negative prep time",
			req: APIRecipeRequest{
				Title:          "Test Recipe",
				IngredientsMD:  "- 1 cup flour",
				InstructionsMD: "Mix and bake",
				PrepTime:       -5,
			},
			wantErr: true,
			errMsg:  "prep time cannot be negative",
		},
		{
			name: "negative cook time",
			req: APIRecipeRequest{
				Title:          "Test Recipe",
				IngredientsMD:  "- 1 cup flour",
				InstructionsMD: "Mix and bake",
				CookTime:       -10,
			},
			wantErr: true,
			errMsg:  "cook time cannot be negative",
		},
		{
			name: "negative calories",
			req: APIRecipeRequest{
				Title:          "Test Recipe",
				IngredientsMD:  "- 1 cup flour",
				InstructionsMD: "Mix and bake",
				Calories:       -100,
			},
			wantErr: true,
			errMsg:  "calories cannot be negative",
		},
		{
			name: "zero values are valid",
			req: APIRecipeRequest{
				Title:          "Test Recipe",
				IngredientsMD:  "- 1 cup flour",
				InstructionsMD: "Mix and bake",
				PrepTime:       0,
				CookTime:       0,
				Calories:       0,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateRecipeRequest(tt.req)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if err.Error() != tt.errMsg {
					t.Errorf("expected error '%s', got '%s'", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
			}
		})
	}
}

func TestSendJSONError_SendsErrorResponseWithStatusCode(t *testing.T) {
	rec := httptest.NewRecorder()
	sendJSONError(rec, "test error message", http.StatusBadRequest)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
	}

	var response APIErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if response.Success {
		t.Error("expected Success to be false")
	}
	if response.Error != "test error message" {
		t.Errorf("expected error 'test error message', got '%s'", response.Error)
	}
}

func TestSendJSONResponse_SendsSuccessResponseWithRecipeID(t *testing.T) {
	rec := httptest.NewRecorder()
	sendJSONResponse(rec, "Recipe created", 42)

	if rec.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
	}

	var response APIRecipeResponse
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !response.Success {
		t.Error("expected Success to be true")
	}
	if response.Message != "Recipe created" {
		t.Errorf("expected message 'Recipe created', got '%s'", response.Message)
	}
	if response.RecipeID != 42 {
		t.Errorf("expected RecipeID 42, got %d", response.RecipeID)
	}
}

func TestAPIHealthHandler_RespondsBasedOnHttpMethod(t *testing.T) {
	t.Run("returns success when method is GET", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		rec := httptest.NewRecorder()

		APIHealthHandler(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var response map[string]interface{}
		if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}

		if response["success"] != true {
			t.Error("expected success to be true")
		}
		if response["message"] != "API is healthy" {
			t.Errorf("expected message 'API is healthy', got '%v'", response["message"])
		}
	})

	t.Run("returns method not allowed when method is POST", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/health", nil)
		rec := httptest.NewRecorder()

		APIHealthHandler(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
		}
	})
}

func TestAPICreateRecipeHandler_CreatesRecipeBasedOnInput(t *testing.T) {
	t.Run("returns method not allowed when method is not POST", func(t *testing.T) {
		h := &Handler{}
		req := httptest.NewRequest(http.MethodGet, "/api/recipe/upload", nil)
		rec := httptest.NewRecorder()

		h.APICreateRecipeHandler(rec, req)

		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
		}
	})

	t.Run("returns error when JSON is invalid", func(t *testing.T) {
		h := &Handler{}
		req := httptest.NewRequest(http.MethodPost, "/api/recipe/upload", bytes.NewBufferString("not json"))
		rec := httptest.NewRecorder()

		h.APICreateRecipeHandler(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}

		var response APIErrorResponse
		json.NewDecoder(rec.Body).Decode(&response)
		if response.Error != "Invalid JSON format" {
			t.Errorf("expected error 'Invalid JSON format', got '%s'", response.Error)
		}
	})

	t.Run("returns error when validation fails", func(t *testing.T) {
		h := &Handler{}
		body := `{"title": "", "ingredients_md": "test", "instructions_md": "test"}`
		req := httptest.NewRequest(http.MethodPost, "/api/recipe/upload", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()

		h.APICreateRecipeHandler(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}

		var response APIErrorResponse
		json.NewDecoder(rec.Body).Decode(&response)
		if response.Error != "title is required" {
			t.Errorf("expected error 'title is required', got '%s'", response.Error)
		}
	})

	t.Run("creates recipe and returns ID when input is valid", func(t *testing.T) {
		mockAuthStore := &mocks.MockAuthStore{
			GetUserIDByUsernameFunc: func(username string) (int, error) {
				return 1, nil
			},
		}
		mockRecipeStore := &mocks.MockRecipeStore{
			SaveFunc: func(recipe models.Recipe) (int, error) {
				return 123, nil
			},
		}

		h := &Handler{
			AuthStore:   mockAuthStore,
			RecipeStore: mockRecipeStore,
		}

		body := `{
			"title": "Test Recipe",
			"ingredients_md": "- 1 cup flour",
			"instructions_md": "Mix and bake",
			"prep_time": 10,
			"cook_time": 20,
			"calories": 300
		}`
		req := httptest.NewRequest(http.MethodPost, "/api/recipe/upload", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()

		h.APICreateRecipeHandler(rec, req)

		if rec.Code != http.StatusCreated {
			t.Errorf("expected status %d, got %d", http.StatusCreated, rec.Code)
		}

		var response APIRecipeResponse
		json.NewDecoder(rec.Body).Decode(&response)
		if !response.Success {
			t.Error("expected success to be true")
		}
		if response.RecipeID != 123 {
			t.Errorf("expected RecipeID 123, got %d", response.RecipeID)
		}
	})

	t.Run("decodes and stores base64 image when provided", func(t *testing.T) {
		mockAuthStore := &mocks.MockAuthStore{
			GetUserIDByUsernameFunc: func(username string) (int, error) {
				return 1, nil
			},
		}
		var capturedRecipe models.Recipe
		mockRecipeStore := &mocks.MockRecipeStore{
			SaveFunc: func(recipe models.Recipe) (int, error) {
				capturedRecipe = recipe
				return 124, nil
			},
		}

		h := &Handler{
			AuthStore:   mockAuthStore,
			RecipeStore: mockRecipeStore,
		}

		body := `{
			"title": "Test Recipe",
			"ingredients_md": "- 1 cup flour",
			"instructions_md": "Mix and bake",
			"image_base64": "SGVsbG8gV29ybGQ="
		}`
		req := httptest.NewRequest(http.MethodPost, "/api/recipe/upload", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()

		h.APICreateRecipeHandler(rec, req)

		if rec.Code != http.StatusCreated {
			t.Errorf("expected status %d, got %d", http.StatusCreated, rec.Code)
		}

		if string(capturedRecipe.Image) != "Hello World" {
			t.Errorf("expected decoded image 'Hello World', got '%s'", string(capturedRecipe.Image))
		}
	})

	t.Run("strips data URI prefix before decoding image", func(t *testing.T) {
		mockAuthStore := &mocks.MockAuthStore{
			GetUserIDByUsernameFunc: func(username string) (int, error) {
				return 1, nil
			},
		}
		var capturedRecipe models.Recipe
		mockRecipeStore := &mocks.MockRecipeStore{
			SaveFunc: func(recipe models.Recipe) (int, error) {
				capturedRecipe = recipe
				return 125, nil
			},
		}

		h := &Handler{
			AuthStore:   mockAuthStore,
			RecipeStore: mockRecipeStore,
		}

		body := `{
			"title": "Test Recipe",
			"ingredients_md": "- 1 cup flour",
			"instructions_md": "Mix and bake",
			"image_base64": "data:image/png;base64,SGVsbG8gV29ybGQ="
		}`
		req := httptest.NewRequest(http.MethodPost, "/api/recipe/upload", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()

		h.APICreateRecipeHandler(rec, req)

		if rec.Code != http.StatusCreated {
			t.Errorf("expected status %d, got %d", http.StatusCreated, rec.Code)
		}

		if string(capturedRecipe.Image) != "Hello World" {
			t.Errorf("expected decoded image 'Hello World', got '%s'", string(capturedRecipe.Image))
		}
	})

	t.Run("returns error when base64 image is invalid", func(t *testing.T) {
		mockAuthStore := &mocks.MockAuthStore{
			GetUserIDByUsernameFunc: func(username string) (int, error) {
				return 1, nil
			},
		}

		h := &Handler{
			AuthStore: mockAuthStore,
		}

		body := `{
			"title": "Test Recipe",
			"ingredients_md": "- 1 cup flour",
			"instructions_md": "Mix and bake",
			"image_base64": "not-valid-base64!!!"
		}`
		req := httptest.NewRequest(http.MethodPost, "/api/recipe/upload", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()

		h.APICreateRecipeHandler(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}

		var response APIErrorResponse
		json.NewDecoder(rec.Body).Decode(&response)
		if response.Error != "Invalid base64 image data" {
			t.Errorf("expected error 'Invalid base64 image data', got '%s'", response.Error)
		}
	})

	t.Run("returns error when store fails to save recipe", func(t *testing.T) {
		mockAuthStore := &mocks.MockAuthStore{
			GetUserIDByUsernameFunc: func(username string) (int, error) {
				return 1, nil
			},
		}
		mockRecipeStore := &mocks.MockRecipeStore{
			SaveFunc: func(recipe models.Recipe) (int, error) {
				return 0, errors.New("database error")
			},
		}

		h := &Handler{
			AuthStore:   mockAuthStore,
			RecipeStore: mockRecipeStore,
		}

		body := `{
			"title": "Test Recipe",
			"ingredients_md": "- 1 cup flour",
			"instructions_md": "Mix and bake"
		}`
		req := httptest.NewRequest(http.MethodPost, "/api/recipe/upload", bytes.NewBufferString(body))
		rec := httptest.NewRecorder()

		h.APICreateRecipeHandler(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
		}
	})
}
