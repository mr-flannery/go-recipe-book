package handlers

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/mr-flannery/go-recipe-book/src/auth"
	"github.com/mr-flannery/go-recipe-book/src/logging"
	"github.com/mr-flannery/go-recipe-book/src/models"
)

// Using models.RecipeSearchResult for recipe search results

type APIRecipeRequest struct {
	Title          string `json:"title"`
	Description    string `json:"description,omitempty"`
	IngredientsMD  string `json:"ingredients_md"`
	InstructionsMD string `json:"instructions_md"`
	PrepTime       int    `json:"prep_time"`
	CookTime       int    `json:"cook_time"`
	Calories       int    `json:"calories"`
	Source         string `json:"source,omitempty"`
	ImageBase64    string `json:"image_base64,omitempty"`
}

type APIRecipeResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	RecipeID int    `json:"recipe_id,omitempty"`
}

type APIErrorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

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

func sendJSONError(w http.ResponseWriter, message string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := APIErrorResponse{
		Success: false,
		Error:   message,
	}

	json.NewEncoder(w).Encode(response)
}

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

func (h *Handler) APICreateRecipeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != http.MethodPost {
		sendJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req APIRecipeRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		logging.AddError(ctx, err, "Failed to decode API recipe request")
		sendJSONError(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if err := validateRecipeRequest(req); err != nil {
		logging.AddError(ctx, err, "API recipe validation failed")
		sendJSONError(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := auth.GetUserIDFromContext(ctx)
	if userID == 0 {
		sendJSONError(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var imageData []byte
	if req.ImageBase64 != "" {
		imageStr := req.ImageBase64
		if strings.HasPrefix(imageStr, "data:image/") {
			commaIndex := strings.Index(imageStr, ",")
			if commaIndex != -1 {
				imageStr = imageStr[commaIndex+1:]
			}
		}

		decodedData, err := base64.StdEncoding.DecodeString(imageStr)
		if err != nil {
			logging.AddError(ctx, err, "Failed to decode base64 image data")
			sendJSONError(w, "Invalid base64 image data", http.StatusBadRequest)
			return
		}
		imageData = decodedData
	}

	recipe := models.Recipe{
		Title:          strings.TrimSpace(req.Title),
		Description:    strings.TrimSpace(req.Description),
		IngredientsMD:  strings.TrimSpace(req.IngredientsMD),
		InstructionsMD: strings.TrimSpace(req.InstructionsMD),
		PrepTime:       req.PrepTime,
		CookTime:       req.CookTime,
		Calories:       req.Calories,
		Source:         strings.TrimSpace(req.Source),
		Image:          imageData,
		AuthorID:       userID,
	}

	recipeID, err := h.RecipeStore.Save(ctx, recipe)
	if err != nil {
		logging.AddError(ctx, err, "Failed to save recipe via API")
		sendJSONError(w, "Failed to save recipe", http.StatusInternalServerError)
		return
	}

	logging.AddMany(ctx, map[string]any{
		"action":       "api.recipe.create",
		"recipe.id":    recipeID,
		"recipe.title": recipe.Title,
	})
	sendJSONResponse(w, "Recipe created successfully", recipeID)
}

func APIHealthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	commit := os.Getenv("COMMIT_HASH")
	if commit == "" {
		commit = "dev"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]any{
		"success": true,
		"message": "API is healthy",
		"version": "1.0.0",
		"commit":  commit,
	}

	json.NewEncoder(w).Encode(response)
}

func (h *Handler) APISearchIngredientsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != http.MethodGet {
		sendJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]string{})
		return
	}

	results, err := h.IngredientStore.Search(ctx, query, 10)
	if err != nil {
		logging.AddError(ctx, err, "Failed to search ingredients")
		sendJSONError(w, "Failed to search ingredients", http.StatusInternalServerError)
		return
	}

	if results == nil {
		results = []string{}
	}

	logging.AddMany(ctx, map[string]any{
		"action":       "api.ingredients.search",
		"result.count": len(results),
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (h *Handler) APISearchRecipesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != http.MethodGet {
		sendJSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]models.RecipeSearchResult{})
		return
	}

	results, err := h.RecipeStore.SearchByTitle(ctx, query, 10)
	if err != nil {
		logging.AddError(ctx, err, "Failed to search recipes")
		sendJSONError(w, "Failed to search recipes", http.StatusInternalServerError)
		return
	}

	if results == nil {
		results = []models.RecipeSearchResult{}
	}

	logging.AddMany(ctx, map[string]any{
		"action":       "api.recipes.search",
		"result.count": len(results),
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
