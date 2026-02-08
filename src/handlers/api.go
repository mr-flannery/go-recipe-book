package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/mr-flannery/go-recipe-book/src/auth"
	"github.com/mr-flannery/go-recipe-book/src/models"
)

// APIRecipeRequest represents the JSON structure for recipe upload via API
type APIRecipeRequest struct {
	Title          string `json:"title"`
	IngredientsMD  string `json:"ingredients_md"`
	InstructionsMD string `json:"instructions_md"`
	PrepTime       int    `json:"prep_time"`
	CookTime       int    `json:"cook_time"`
	Calories       int    `json:"calories"`
	ImageBase64    string `json:"image_base64,omitempty"`
}

// APIRecipeResponse represents the JSON response for recipe operations
type APIRecipeResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	RecipeID int    `json:"recipe_id,omitempty"`
}

// APIErrorResponse represents the JSON structure for API errors
type APIErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

// validateRecipeRequest validates the recipe request data
func validateRecipeRequest(req APIRecipeRequest) error {
	if strings.TrimSpace(req.Title) == "" {
		return fmt.Errorf("title is required")
	}

	if strings.TrimSpace(req.IngredientsMD) == "" {
		return fmt.Errorf("ingredients are required")
	}

	if strings.TrimSpace(req.InstructionsMD) == "" {
		return fmt.Errorf("instructions are required")
	}

	if req.PrepTime < 0 {
		return fmt.Errorf("prep time cannot be negative")
	}

	if req.CookTime < 0 {
		return fmt.Errorf("cook time cannot be negative")
	}

	if req.Calories < 0 {
		return fmt.Errorf("calories cannot be negative")
	}

	return nil
}

// sendJSONError sends a JSON error response
func sendJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := APIErrorResponse{
		Success: false,
		Error:   message,
	}

	json.NewEncoder(w).Encode(response)
}

// sendJSONResponse sends a JSON success response
func sendJSONResponse(w http.ResponseWriter, message string, recipeID int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := APIRecipeResponse{
		Success:  true,
		Message:  message,
		RecipeID: recipeID,
	}

	json.NewEncoder(w).Encode(response)
}

// APICreateRecipeHandler handles recipe creation via API
func APICreateRecipeHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST method
	if r.Method != http.MethodPost {
		sendJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse JSON request body
	var req APIRecipeRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Reject unknown fields

	if err := decoder.Decode(&req); err != nil {
		slog.Error("Failed to decode API recipe request", "error", err)
		sendJSONError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	// Validate the request
	if err := validateRecipeRequest(req); err != nil {
		slog.Warn("API recipe validation failed", "error", err)
		sendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get admin user ID (as specified in requirements)
	adminID, err := auth.GetAdminUserID()
	if err != nil {
		slog.Error("Failed to get admin user ID", "error", err)
		sendJSONError(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Handle image data if provided
	var imageData []byte
	if req.ImageBase64 != "" {
		// Remove data URL prefix if present (e.g., "data:image/jpeg;base64,")
		imageStr := req.ImageBase64
		if strings.HasPrefix(imageStr, "data:image/") {
			commaIndex := strings.Index(imageStr, ",")
			if commaIndex != -1 {
				imageStr = imageStr[commaIndex+1:]
			}
		}

		// Decode base64 image data
		decodedData, err := base64.StdEncoding.DecodeString(imageStr)
		if err != nil {
			slog.Error("Failed to decode base64 image data", "error", err)
			sendJSONError(w, "Invalid base64 image data", http.StatusBadRequest)
			return
		}
		imageData = decodedData
		slog.Info("API recipe image processed", "size", len(imageData))
	}

	// Create recipe model
	recipe := models.Recipe{
		Title:          strings.TrimSpace(req.Title),
		IngredientsMD:  strings.TrimSpace(req.IngredientsMD),
		InstructionsMD: strings.TrimSpace(req.InstructionsMD),
		PrepTime:       req.PrepTime,
		CookTime:       req.CookTime,
		Calories:       req.Calories,
		Image:          imageData,
		AuthorID:       adminID,
	}

	// Save recipe to database
	recipeID, err := models.SaveRecipe(recipe)
	if err != nil {
		slog.Error("Failed to save recipe via API", "error", err)
		sendJSONError(w, "Failed to save recipe", http.StatusInternalServerError)
		return
	}

	slog.Info("Recipe created successfully via API", "title", recipe.Title, "author_id", adminID, "recipe_id", recipeID)
	sendJSONResponse(w, "Recipe created successfully", recipeID)
}

// APIHealthHandler provides a simple health check endpoint
func APIHealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"success": true,
		"message": "API is healthy",
		"version": "1.0.0",
	}

	json.NewEncoder(w).Encode(response)
}
