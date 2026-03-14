// compare-results.go
//
// Compare extraction results against expected outputs and generate a summary table.
//
// Usage:
//   go run ./scripts/compare-results
//   go run ./scripts/compare-results -verbose    # Show detailed differences
//   go run ./scripts/compare-results -json       # Output as JSON

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var baseDir = "test/llm-extraction"

type Recipe struct {
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	IngredientsMD      string   `json:"ingredients_md"`
	InstructionsMD     string   `json:"instructions_md"`
	PrepTimeMinutes    *int     `json:"prep_time_minutes"`
	CookTimeMinutes    *int     `json:"cook_time_minutes"`
	CaloriesPerServing *int     `json:"calories_per_serving"`
	SuggestedTags      []string `json:"suggested_tags"`
	Confidence         float64  `json:"confidence"`
	ConfidenceNotes    string   `json:"confidence_notes"`
}

type Score struct {
	Title        float64
	Ingredients  float64
	Instructions float64
	Metadata     float64
	Overall      float64
}

type Result struct {
	SampleID string
	Model    string
	Score    Score
	Error    string
}

func loadRecipe(path string) (*Recipe, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var recipe Recipe
	if err := json.Unmarshal(data, &recipe); err != nil {
		return nil, err
	}

	return &recipe, nil
}

func countListItems(md string) int {
	lines := strings.Split(md, "\n")
	count := 0
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "- ") || strings.HasPrefix(line, "* ") {
			count++
		} else if len(line) > 0 && line[0] >= '1' && line[0] <= '9' && strings.Contains(line, ".") {
			count++
		}
	}
	return count
}

func normalizeText(s string) string {
	s = strings.ToLower(s)
	s = strings.TrimSpace(s)
	return s
}

func scoreTitle(expected, actual string) float64 {
	if normalizeText(expected) == normalizeText(actual) {
		return 1.0
	}
	// Partial match - check if one contains the other
	e, a := normalizeText(expected), normalizeText(actual)
	if strings.Contains(e, a) || strings.Contains(a, e) {
		return 0.8
	}
	// Check word overlap
	eWords := strings.Fields(e)
	aWords := strings.Fields(a)
	matches := 0
	for _, ew := range eWords {
		for _, aw := range aWords {
			if ew == aw {
				matches++
				break
			}
		}
	}
	if len(eWords) > 0 {
		return float64(matches) / float64(len(eWords)) * 0.7
	}
	return 0.0
}

func scoreIngredients(expected, actual string) float64 {
	expectedCount := countListItems(expected)
	actualCount := countListItems(actual)

	if expectedCount == 0 {
		if actualCount == 0 {
			return 1.0
		}
		return 0.5 // Can't evaluate, give partial credit
	}

	// Score based on count similarity (rough approximation)
	ratio := float64(actualCount) / float64(expectedCount)
	if ratio > 1.0 {
		ratio = 1.0 / ratio // Penalize over-extraction too
	}

	return math.Min(1.0, ratio)
}

func scoreInstructions(expected, actual string) float64 {
	expectedCount := countListItems(expected)
	actualCount := countListItems(actual)

	if expectedCount == 0 {
		if actualCount == 0 {
			return 1.0
		}
		return 0.5
	}

	ratio := float64(actualCount) / float64(expectedCount)
	if ratio > 1.0 {
		ratio = 1.0 / ratio
	}

	return math.Min(1.0, ratio)
}

func scoreMetadata(expected, actual *Recipe) float64 {
	score := 0.0
	count := 0

	// Prep time
	if expected.PrepTimeMinutes != nil {
		count++
		if actual.PrepTimeMinutes != nil {
			diff := math.Abs(float64(*expected.PrepTimeMinutes - *actual.PrepTimeMinutes))
			if diff == 0 {
				score += 1.0
			} else if diff <= 5 {
				score += 0.8
			} else if diff <= 15 {
				score += 0.5
			}
		}
	}

	// Cook time
	if expected.CookTimeMinutes != nil {
		count++
		if actual.CookTimeMinutes != nil {
			diff := math.Abs(float64(*expected.CookTimeMinutes - *actual.CookTimeMinutes))
			if diff == 0 {
				score += 1.0
			} else if diff <= 5 {
				score += 0.8
			} else if diff <= 15 {
				score += 0.5
			}
		}
	}

	// Calories
	if expected.CaloriesPerServing != nil {
		count++
		if actual.CaloriesPerServing != nil {
			ratio := float64(*actual.CaloriesPerServing) / float64(*expected.CaloriesPerServing)
			if ratio >= 0.9 && ratio <= 1.1 {
				score += 1.0
			} else if ratio >= 0.7 && ratio <= 1.3 {
				score += 0.7
			} else if ratio >= 0.5 && ratio <= 1.5 {
				score += 0.4
			}
		}
	}

	if count == 0 {
		return 1.0 // No metadata to compare
	}

	return score / float64(count)
}

func compareRecipes(expected, actual *Recipe) Score {
	titleScore := scoreTitle(expected.Title, actual.Title)
	ingredientsScore := scoreIngredients(expected.IngredientsMD, actual.IngredientsMD)
	instructionsScore := scoreInstructions(expected.InstructionsMD, actual.InstructionsMD)
	metadataScore := scoreMetadata(expected, actual)

	// Weighted overall score
	overall := titleScore*0.10 +
		ingredientsScore*0.35 +
		instructionsScore*0.35 +
		metadataScore*0.20

	return Score{
		Title:        titleScore,
		Ingredients:  ingredientsScore,
		Instructions: instructionsScore,
		Metadata:     metadataScore,
		Overall:      overall,
	}
}

func getModels() []string {
	resultsDir := filepath.Join(baseDir, "results")
	entries, _ := os.ReadDir(resultsDir)

	var models []string
	for _, e := range entries {
		if e.IsDir() && !strings.HasPrefix(e.Name(), ".") {
			models = append(models, e.Name())
		}
	}
	return models
}

func getSamples() []string {
	expectedDir := filepath.Join(baseDir, "expected")
	entries, _ := os.ReadDir(expectedDir)

	var samples []string
	for _, e := range entries {
		name := e.Name()
		if strings.HasSuffix(name, ".json") && !strings.HasPrefix(name, "_") {
			samples = append(samples, strings.TrimSuffix(name, ".json"))
		}
	}
	return samples
}

func main() {
	verbose := flag.Bool("verbose", false, "Show detailed differences")
	jsonOutput := flag.Bool("json", false, "Output as JSON")
	flag.Parse()

	models := getModels()
	samples := getSamples()

	if len(samples) == 0 {
		fmt.Println("No expected outputs found in test/llm-extraction/expected/")
		fmt.Println("Create ground truth JSON files for your samples first.")
		os.Exit(1)
	}

	var results []Result

	for _, sampleID := range samples {
		expectedPath := filepath.Join(baseDir, "expected", sampleID+".json")
		expected, err := loadRecipe(expectedPath)
		if err != nil {
			fmt.Printf("Warning: Could not load expected for %s: %v\n", sampleID, err)
			continue
		}

		for _, model := range models {
			actualPath := filepath.Join(baseDir, "results", model, sampleID+".json")
			actual, err := loadRecipe(actualPath)

			result := Result{
				SampleID: sampleID,
				Model:    model,
			}

			if err != nil {
				result.Error = err.Error()
			} else {
				result.Score = compareRecipes(expected, actual)
			}

			results = append(results, result)

			if *verbose && err == nil {
				fmt.Printf("\n%s / %s:\n", model, sampleID)
				fmt.Printf("  Title:        %.0f%% (expected: %q, got: %q)\n",
					result.Score.Title*100, expected.Title, actual.Title)
				fmt.Printf("  Ingredients:  %.0f%% (expected: %d items, got: %d items)\n",
					result.Score.Ingredients*100,
					countListItems(expected.IngredientsMD),
					countListItems(actual.IngredientsMD))
				fmt.Printf("  Instructions: %.0f%% (expected: %d steps, got: %d steps)\n",
					result.Score.Instructions*100,
					countListItems(expected.InstructionsMD),
					countListItems(actual.InstructionsMD))
				fmt.Printf("  Metadata:     %.0f%%\n", result.Score.Metadata*100)
				fmt.Printf("  Confidence:   %.0f%%\n", actual.Confidence*100)
			}
		}
	}

	if *jsonOutput {
		output, _ := json.MarshalIndent(results, "", "  ")
		fmt.Println(string(output))
		return
	}

	// Print summary table
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Println("COMPARISON SUMMARY")
	fmt.Println(strings.Repeat("=", 80))

	// Calculate averages per model
	modelScores := make(map[string][]float64)
	for _, r := range results {
		if r.Error == "" {
			modelScores[r.Model] = append(modelScores[r.Model], r.Score.Overall)
		}
	}

	// Header
	fmt.Printf("\n%-30s", "Sample")
	for _, model := range models {
		fmt.Printf(" | %-15s", model)
	}
	fmt.Println()
	fmt.Println(strings.Repeat("-", 30+len(models)*18))

	// Rows by sample
	sort.Strings(samples)
	for _, sampleID := range samples {
		fmt.Printf("%-30s", sampleID)
		for _, model := range models {
			found := false
			for _, r := range results {
				if r.SampleID == sampleID && r.Model == model {
					if r.Error != "" {
						fmt.Printf(" | %-15s", "ERROR")
					} else {
						fmt.Printf(" | %5.0f%%         ", r.Score.Overall*100)
					}
					found = true
					break
				}
			}
			if !found {
				fmt.Printf(" | %-15s", "-")
			}
		}
		fmt.Println()
	}

	// Average row
	fmt.Println(strings.Repeat("-", 30+len(models)*18))
	fmt.Printf("%-30s", "AVERAGE")
	for _, model := range models {
		if scores, ok := modelScores[model]; ok && len(scores) > 0 {
			sum := 0.0
			for _, s := range scores {
				sum += s
			}
			avg := sum / float64(len(scores))
			fmt.Printf(" | %5.0f%%         ", avg*100)
		} else {
			fmt.Printf(" | %-15s", "-")
		}
	}
	fmt.Println()

	fmt.Println("\nScoring weights: Title 10%, Ingredients 35%, Instructions 35%, Metadata 20%")
}
