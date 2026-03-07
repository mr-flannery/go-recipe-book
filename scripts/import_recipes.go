// Script to import recipes from recipes2.json (old system export) to the new recipe API.
// Usage: go run scripts/import_recipes.go -api-key=<key> [-url=http://localhost:8080] [-file=recipes2.json] [-dry-run]
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

type OldRecipeExport struct {
	Count   int         `json:"count"`
	Results []OldRecipe `json:"results"`
}

type OldRecipe struct {
	ID               int                  `json:"id"`
	Title            string               `json:"title"`
	Info             string               `json:"info"`
	Directions       string               `json:"directions"`
	Source           string               `json:"source"`
	PrepTime         int                  `json:"prep_time"`
	CookTime         int                  `json:"cook_time"`
	Servings         int                  `json:"servings"`
	Rating           int                  `json:"rating"`
	IngredientGroups []OldIngredientGroup `json:"ingredient_groups"`
	Tags             []OldTag             `json:"tags"`
	Course           *OldCategory         `json:"course"`
	Cuisine          *OldCategory         `json:"cuisine"`
	Username         string               `json:"username"`
	PubDate          string               `json:"pub_date"`
}

type OldIngredientGroup struct {
	Title       string          `json:"title"`
	Ingredients []OldIngredient `json:"ingredients"`
}

type OldIngredient struct {
	Numerator   float64 `json:"numerator"`
	Denominator float64 `json:"denominator"`
	Measurement *string `json:"measurement"`
	Title       string  `json:"title"`
}

type OldTag struct {
	Title string `json:"title"`
}

type OldCategory struct {
	Title string `json:"title"`
}

type APIRecipeRequest struct {
	Title          string `json:"title"`
	IngredientsMD  string `json:"ingredients_md"`
	InstructionsMD string `json:"instructions_md"`
	PrepTime       int    `json:"prep_time"`
	CookTime       int    `json:"cook_time"`
	Calories       int    `json:"calories"`
}

type APIResponse struct {
	Success  bool   `json:"success"`
	Message  string `json:"message"`
	RecipeID int    `json:"recipe_id"`
	Error    string `json:"error"`
}

func main() {
	apiKey := flag.String("api-key", "", "API key for authentication (required)")
	baseURL := flag.String("url", "http://localhost:8080", "Base URL of the recipe API")
	inputFile := flag.String("file", "recipes2.json", "Path to the recipes2.json export file")
	dryRun := flag.Bool("dry-run", false, "Print what would be imported without actually importing")
	flag.Parse()

	if *apiKey == "" {
		log.Fatal("Error: -api-key is required")
	}

	data, err := os.ReadFile(*inputFile)
	if err != nil {
		log.Fatalf("Error reading input file: %v", err)
	}

	var export OldRecipeExport
	if err := json.Unmarshal(data, &export); err != nil {
		log.Fatalf("Error parsing JSON: %v", err)
	}

	log.Printf("Found %d recipes to import", len(export.Results))

	client := &http.Client{Timeout: 30 * time.Second}
	successCount := 0
	failCount := 0

	for i, recipe := range export.Results {
		apiReq := convertRecipe(recipe)

		if *dryRun {
			log.Printf("[DRY RUN] %d/%d: Would import: %s", i+1, len(export.Results), recipe.Title)
			continue
		}

		log.Printf("Importing %d/%d: %s", i+1, len(export.Results), recipe.Title)

		recipeID, err := createRecipe(client, *baseURL, *apiKey, apiReq)
		if err != nil {
			log.Printf("  ERROR: %v", err)
			failCount++
			continue
		}

		log.Printf("  SUCCESS: created recipe ID %d", recipeID)
		successCount++

		time.Sleep(100 * time.Millisecond)
	}

	if !*dryRun {
		log.Printf("\nImport complete: %d succeeded, %d failed", successCount, failCount)
	}
}

func convertRecipe(old OldRecipe) APIRecipeRequest {
	ingredientsMD := buildIngredientsMD(old)
	instructionsMD := buildInstructionsMD(old)

	return APIRecipeRequest{
		Title:          old.Title,
		IngredientsMD:  ingredientsMD,
		InstructionsMD: instructionsMD,
		PrepTime:       old.PrepTime,
		CookTime:       old.CookTime,
		Calories:       0,
	}
}

func buildIngredientsMD(old OldRecipe) string {
	var sb strings.Builder

	for _, group := range old.IngredientGroups {
		if group.Title != "" {
			sb.WriteString("## ")
			sb.WriteString(group.Title)
			sb.WriteString("\n\n")
		}

		for _, ing := range group.Ingredients {
			sb.WriteString("- ")
			sb.WriteString(formatIngredient(ing))
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	return strings.TrimSpace(sb.String())
}

func formatIngredient(ing OldIngredient) string {
	var parts []string

	if ing.Numerator > 0 {
		amount := formatAmount(ing.Numerator, ing.Denominator)
		parts = append(parts, amount)
	}

	if ing.Measurement != nil && *ing.Measurement != "" {
		parts = append(parts, *ing.Measurement)
	}

	parts = append(parts, ing.Title)

	return strings.Join(parts, " ")
}

func formatAmount(num, denom float64) string {
	if denom == 1 {
		if num == float64(int(num)) {
			return fmt.Sprintf("%d", int(num))
		}
		return fmt.Sprintf("%.1f", num)
	}

	whole := int(num / denom)
	remainder := int(num) % int(denom)

	if whole > 0 && remainder > 0 {
		return fmt.Sprintf("%d %d/%d", whole, remainder, int(denom))
	} else if whole > 0 {
		return fmt.Sprintf("%d", whole)
	}
	return fmt.Sprintf("%d/%d", int(num), int(denom))
}

func buildInstructionsMD(old OldRecipe) string {
	var sb strings.Builder

	if old.Info != "" {
		sb.WriteString("**Info:** ")
		sb.WriteString(old.Info)
		sb.WriteString("\n\n")
	}

	if old.Course != nil && old.Course.Title != "" && old.Course.Title != "beliebig" {
		sb.WriteString("**Gang:** ")
		sb.WriteString(old.Course.Title)
		sb.WriteString("\n\n")
	}

	if old.Cuisine != nil && old.Cuisine.Title != "" && old.Cuisine.Title != "beliebig" {
		sb.WriteString("**Küche:** ")
		sb.WriteString(old.Cuisine.Title)
		sb.WriteString("\n\n")
	}

	if old.Servings > 0 && old.Servings < 999999999 {
		sb.WriteString(fmt.Sprintf("**Portionen:** %d\n\n", old.Servings))
	}

	if old.Directions != "" {
		sb.WriteString("## Zubereitung\n\n")
		sb.WriteString(old.Directions)
		sb.WriteString("\n")
	}

	if old.Source != "" {
		sb.WriteString("\n---\n")
		sb.WriteString("**Quelle:** ")
		sb.WriteString(old.Source)
		sb.WriteString("\n")
	}

	if old.Rating > 0 {
		sb.WriteString(fmt.Sprintf("\n**Bewertung:** %d/5\n", old.Rating))
	}

	return strings.TrimSpace(sb.String())
}

func createRecipe(client *http.Client, baseURL, apiKey string, req APIRecipeRequest) (int, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", baseURL+"/api/recipe/upload", bytes.NewReader(body))
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(httpReq)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp APIResponse
	if err := json.Unmarshal(respBody, &apiResp); err != nil {
		return 0, fmt.Errorf("failed to parse response (status %d): %s", resp.StatusCode, string(respBody))
	}

	if !apiResp.Success {
		return 0, fmt.Errorf("API error: %s", apiResp.Error)
	}

	return apiResp.RecipeID, nil
}
