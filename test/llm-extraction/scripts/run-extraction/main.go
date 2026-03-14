// run-extraction.go
//
// Run recipe extraction prompts against LLM APIs and save results.
//
// Usage:
//   go run ./scripts/run-extraction                           # Run all samples against all models
//   go run ./scripts/run-extraction -sample image-01          # Run specific sample against all models
//   go run ./scripts/run-extraction -model gpt-4o-mini        # Run all samples against specific model
//   go run ./scripts/run-extraction -sample image-01 -model gpt-4o-mini
//
// Environment variables:
//   OPENAI_API_KEY      - For GPT-4o-mini
//   ANTHROPIC_API_KEY   - For Claude 3.5 Sonnet
//   GOOGLE_API_KEY      - For Gemini 1.5 Flash
//   OPENROUTER_API_KEY  - Alternative: use OpenRouter for all models

package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var (
	baseDir    = "test/llm-extraction"
	promptFile = "prompts/extract-recipe-v1.txt"
)

type Model struct {
	Name     string
	Provider string
	Endpoint string
	APIKey   string
	Model    string
}

func getModels() []Model {
	models := []Model{}
	
	if key := os.Getenv("OPENROUTER_API_KEY"); key != "" && len(models) == 0 {
		models = append(models,
			Model{
				Name:     "gpt-4o-mini",
				Provider: "openrouter",
				Endpoint: "https://openrouter.ai/api/v1/chat/completions",
				APIKey:   key,
				Model:    "openai/gpt-4o-mini",
			},
			Model{
				Name:     "claude-3.5-sonnet",
				Provider: "openrouter",
				Endpoint: "https://openrouter.ai/api/v1/chat/completions",
				APIKey:   key,
				Model:    "anthropic/claude-3.5-sonnet",
			},
			Model{
				Name:     "gemini-2.5-flash-lite",
				Provider: "openrouter",
				Endpoint: "https://openrouter.ai/api/v1/chat/completions",
				APIKey:   key,
				Model:    "google/gemini-2.5-flash-lite",
			},
			Model{
				Name:     "pixtral-12b",
				Provider: "openrouter",
				Endpoint: "https://openrouter.ai/api/v1/chat/completions",
				APIKey:   key,
				Model:    "mistralai/pixtral-12b",
			},
			Model{
				Name:     "mistral-small",
				Provider: "openrouter",
				Endpoint: "https://openrouter.ai/api/v1/chat/completions",
				APIKey:   key,
				Model:    "mistralai/mistral-small-24b-instruct-2501",
			},
			Model{
				Name:     "mistral-large",
				Provider: "openrouter",
				Endpoint: "https://openrouter.ai/api/v1/chat/completions",
				APIKey:   key,
				Model:    "mistralai/mistral-large-2512",
			},
		)
	}

	return models
}

type Sample struct {
	ID         string
	Type       string // image, website, video
	Path       string
	SourceType string // for prompt template
}

func getSamples() ([]Sample, error) {
	samplesDir := filepath.Join(baseDir, "samples")
	entries, err := os.ReadDir(samplesDir)
	if err != nil {
		return nil, err
	}

	var samples []Sample
	for _, e := range entries {
		name := e.Name()
	
		var sampleType, sourceType string
		switch {
		case strings.HasPrefix(name, "image-"):
			sampleType = "image"
			sourceType = "image"
		case strings.HasPrefix(name, "website-"):
			sampleType = "website"
			sourceType = "website HTML"
		case strings.HasPrefix(name, "video-"):
			sampleType = "video"
			sourceType = "video transcript"
		default:
			continue
		}

		// Extract ID (filename without extension)
		id := strings.TrimSuffix(name, filepath.Ext(name))

		samples = append(samples, Sample{
			ID:         id,
			Type:       sampleType,
			Path:       filepath.Join(samplesDir, name),
			SourceType: sourceType,
		})
	}

	return samples, nil
}

func loadPromptTemplate() (string, error) {
	data, err := os.ReadFile(filepath.Join(baseDir, promptFile))
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func buildPrompt(template string, sample Sample) (string, []byte, error) {
	content, err := os.ReadFile(sample.Path)
	if err != nil {
		return "", nil, err
	}

	// For images, we'll pass the content separately
	var imageData []byte
	var textContent string

	if sample.Type == "image" {
		imageData = content
		textContent = "[Image attached]"
	} else {
		textContent = string(content)
	}

	prompt := strings.ReplaceAll(template, "{source_type}", sample.SourceType)
	prompt = strings.ReplaceAll(prompt, "{content}", textContent)

	return prompt, imageData, nil
}

func callOpenAI(model Model, prompt string, imageData []byte) (string, error) {
	var messages []map[string]interface{}

	if len(imageData) > 0 {
		// Vision request with image
		base64Image := base64.StdEncoding.EncodeToString(imageData)
		mimeType := "image/jpeg"
		if bytes.HasPrefix(imageData, []byte{0x89, 0x50, 0x4E, 0x47}) {
			mimeType = "image/png"
		}

		messages = []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"type": "text",
						"text": prompt,
					},
					{
						"type": "image_url",
						"image_url": map[string]string{
							"url": fmt.Sprintf("data:%s;base64,%s", mimeType, base64Image),
						},
					},
				},
			},
		}
	} else {
		messages = []map[string]interface{}{
			{"role": "user", "content": prompt},
		}
	}

	reqBody := map[string]interface{}{
		"model":       model.Model,
		"messages":    messages,
		"max_tokens":  4096,
		"temperature": 0.1,
	}

	return makeRequest(model.Endpoint, model.APIKey, "Bearer", reqBody)
}

func callAnthropic(model Model, prompt string, imageData []byte) (string, error) {
	var content []map[string]interface{}

	if len(imageData) > 0 {
		base64Image := base64.StdEncoding.EncodeToString(imageData)
		mimeType := "image/jpeg"
		if bytes.HasPrefix(imageData, []byte{0x89, 0x50, 0x4E, 0x47}) {
			mimeType = "image/png"
		}

		content = []map[string]interface{}{
			{
				"type": "image",
				"source": map[string]string{
					"type":       "base64",
					"media_type": mimeType,
					"data":       base64Image,
				},
			},
			{
				"type": "text",
				"text": prompt,
			},
		}
	} else {
		content = []map[string]interface{}{
			{"type": "text", "text": prompt},
		}
	}

	reqBody := map[string]interface{}{
		"model":      model.Model,
		"max_tokens": 4096,
		"messages": []map[string]interface{}{
			{"role": "user", "content": content},
		},
	}

	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", model.Endpoint, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", model.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	// Extract text from Anthropic response
	if content, ok := result["content"].([]interface{}); ok && len(content) > 0 {
		if block, ok := content[0].(map[string]interface{}); ok {
			if text, ok := block["text"].(string); ok {
				return text, nil
			}
		}
	}

	return "", fmt.Errorf("unexpected response format: %s", string(body))
}

func callGoogle(model Model, prompt string, imageData []byte) (string, error) {
	var parts []map[string]interface{}

	parts = append(parts, map[string]interface{}{
		"text": prompt,
	})

	if len(imageData) > 0 {
		base64Image := base64.StdEncoding.EncodeToString(imageData)
		mimeType := "image/jpeg"
		if bytes.HasPrefix(imageData, []byte{0x89, 0x50, 0x4E, 0x47}) {
			mimeType = "image/png"
		}

		parts = append(parts, map[string]interface{}{
			"inline_data": map[string]string{
				"mime_type": mimeType,
				"data":      base64Image,
			},
		})
	}

	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{"parts": parts},
		},
		"generationConfig": map[string]interface{}{
			"temperature":     0.1,
			"maxOutputTokens": 4096,
		},
	}

	jsonBody, _ := json.Marshal(reqBody)
	url := model.Endpoint + "?key=" + model.APIKey
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	// Extract text from Google response
	if candidates, ok := result["candidates"].([]interface{}); ok && len(candidates) > 0 {
		if candidate, ok := candidates[0].(map[string]interface{}); ok {
			if content, ok := candidate["content"].(map[string]interface{}); ok {
				if parts, ok := content["parts"].([]interface{}); ok && len(parts) > 0 {
					if part, ok := parts[0].(map[string]interface{}); ok {
						if text, ok := part["text"].(string); ok {
							return text, nil
						}
					}
				}
			}
		}
	}

	return "", fmt.Errorf("unexpected response format: %s", string(body))
}

func callOpenRouter(model Model, prompt string, imageData []byte) (string, error) {
	// OpenRouter uses OpenAI-compatible format
	return callOpenAI(model, prompt, imageData)
}

func callMistral(model Model, prompt string, imageData []byte) (string, error) {
	// Mistral uses OpenAI-compatible format
	return callOpenAI(model, prompt, imageData)
}

func makeRequest(endpoint, apiKey, authType string, reqBody map[string]interface{}) (string, error) {
	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", endpoint, bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", authType+" "+apiKey)

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API error %d: %s", resp.StatusCode, string(body))
	}

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	// Extract text from OpenAI-style response
	if choices, ok := result["choices"].([]interface{}); ok && len(choices) > 0 {
		if choice, ok := choices[0].(map[string]interface{}); ok {
			if message, ok := choice["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					return content, nil
				}
			}
		}
	}

	return "", fmt.Errorf("unexpected response format: %s", string(body))
}

func callModel(model Model, prompt string, imageData []byte) (string, error) {
	switch model.Provider {
	case "openai":
		return callOpenAI(model, prompt, imageData)
	case "anthropic":
		return callAnthropic(model, prompt, imageData)
	case "google":
		return callGoogle(model, prompt, imageData)
	case "mistral":
		return callMistral(model, prompt, imageData)
	case "openrouter":
		return callOpenRouter(model, prompt, imageData)
	default:
		return "", fmt.Errorf("unknown provider: %s", model.Provider)
	}
}

func extractJSON(response string) string {
	// Try to extract JSON from the response (may be wrapped in markdown code blocks)
	response = strings.TrimSpace(response)

	// Remove markdown code blocks if present
	if strings.HasPrefix(response, "```json") {
		response = strings.TrimPrefix(response, "```json")
		if idx := strings.LastIndex(response, "```"); idx != -1 {
			response = response[:idx]
		}
	} else if strings.HasPrefix(response, "```") {
		response = strings.TrimPrefix(response, "```")
		if idx := strings.LastIndex(response, "```"); idx != -1 {
			response = response[:idx]
		}
	}

	return strings.TrimSpace(response)
}

func runExtraction(model Model, sample Sample, promptTemplate string) error {
	fmt.Printf("  Running %s on %s...\n", model.Name, sample.ID)

	prompt, imageData, err := buildPrompt(promptTemplate, sample)
	
	if err != nil {
		return fmt.Errorf("failed to build prompt: %w", err)
	}

	response, err := callModel(model, prompt, imageData)
	if err != nil {
		return fmt.Errorf("API call failed: %w", err)
	}

	// Extract and validate JSON
	jsonStr := extractJSON(response)

	// Validate it's valid JSON
	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		// Save raw response for debugging
		outputPath := filepath.Join(baseDir, "results", model.Name, sample.ID+".raw.txt")
		os.WriteFile(outputPath, []byte(response), 0644)
		return fmt.Errorf("invalid JSON in response (raw saved to %s): %w", outputPath, err)
	}

	// Pretty print and save
	prettyJSON, _ := json.MarshalIndent(parsed, "", "  ")
	outputPath := filepath.Join(baseDir, "results", model.Name, sample.ID+".json")
	if err := os.WriteFile(outputPath, prettyJSON, 0644); err != nil {
		return fmt.Errorf("failed to write output: %w", err)
	}

	fmt.Printf("    Saved to %s\n", outputPath)
	return nil
}

func main() {
	sampleFilter := flag.String("sample", "", "Run only this sample (partial match)")
	modelFilter := flag.String("model", "", "Run only this model (partial match)")
	flag.Parse()

	models := getModels()
	if len(models) == 0 {
		fmt.Println("Error: No API keys configured")
		fmt.Println("Set one of: OPENAI_API_KEY, ANTHROPIC_API_KEY, GOOGLE_API_KEY, or OPENROUTER_API_KEY")
		os.Exit(1)
	}

	samples, err := getSamples()
	if err != nil {
		fmt.Printf("Error loading samples: %v\n", err)
		os.Exit(1)
	}

	if len(samples) == 0 {
		fmt.Println("No samples found. Add test files to test/llm-extraction/samples/")
		fmt.Println("(Placeholder .txt files are ignored)")
		os.Exit(1)
	}

	promptTemplate, err := loadPromptTemplate()
	if err != nil {
		fmt.Printf("Error loading prompt template: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Found %d models and %d samples\n\n", len(models), len(samples))

	for _, model := range models {
		if *modelFilter != "" && !strings.Contains(model.Name, *modelFilter) {
			continue
		}

		fmt.Printf("Model: %s\n", model.Name)

		for _, sample := range samples {
			if *sampleFilter != "" && !strings.Contains(sample.ID, *sampleFilter) {
				fmt.Printf("  Skipping sample %s\n", sample.ID)
				continue
			}

			if err := runExtraction(model, sample, promptTemplate); err != nil {
				fmt.Printf("    Error: %v\n", err)
			}
		}
		fmt.Println()
	}

	fmt.Println("Done! Run 'go run scripts/compare-results.go' to compare results.")
}
