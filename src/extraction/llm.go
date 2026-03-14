package extraction

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	openRouterURL = "https://openrouter.ai/api/v1/chat/completions"
	defaultModel  = "google/gemini-2.5-flash-lite"
)

type LLMClient struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

func NewLLMClient(apiKey string) *LLMClient {
	return &LLMClient{
		apiKey: apiKey,
		model:  defaultModel,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role    string        `json:"role"`
	Content []contentPart `json:"content"`
}

type contentPart struct {
	Type     string    `json:"type"`
	Text     string    `json:"text,omitempty"`
	ImageURL *imageURL `json:"image_url,omitempty"`
}

type imageURL struct {
	URL string `json:"url"`
}

type chatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Code    string `json:"code"`
	} `json:"error"`
}

func (c *LLMClient) ExtractRecipeFromText(ctx context.Context, sourceType, content string) (string, *ExtractedRecipe, error) {
	prompt := buildPrompt(sourceType, content)

	request := chatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{
				Role: "user",
				Content: []contentPart{
					{Type: "text", Text: prompt},
				},
			},
		},
	}

	responseText, err := c.sendRequest(ctx, request)
	if err != nil {
		return prompt, nil, err
	}

	recipe, err := parseRecipeResponse(responseText)
	return prompt, recipe, err
}

func (c *LLMClient) ExtractRecipeFromImage(ctx context.Context, imageData []byte, mimeType string) (string, *ExtractedRecipe, error) {
	if mimeType == "" {
		mimeType = detectMimeType(imageData)
	}

	base64Image := base64.StdEncoding.EncodeToString(imageData)
	dataURL := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Image)

	prompt := buildPrompt("image", "")

	request := chatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{
				Role: "user",
				Content: []contentPart{
					{Type: "text", Text: prompt},
					{Type: "image_url", ImageURL: &imageURL{URL: dataURL}},
				},
			},
		},
	}

	responseText, err := c.sendRequest(ctx, request)
	if err != nil {
		return prompt, nil, err
	}

	recipe, err := parseRecipeResponse(responseText)
	return prompt, recipe, err
}

func (c *LLMClient) sendRequest(ctx context.Context, request chatRequest) (string, error) {
	jsonBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", openRouterURL, bytes.NewReader(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("HTTP-Referer", "https://recipe-book.app")
	req.Header.Set("X-Title", "Recipe Book")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("API error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned")
	}

	return chatResp.Choices[0].Message.Content, nil
}

func buildPrompt(sourceType, content string) string {
	sourceDescription := map[string]string{
		"website":    "website HTML content",
		"video":      "video transcript",
		"image":      "image",
		"transcript": "video transcript",
	}

	desc := sourceDescription[sourceType]
	if desc == "" {
		desc = sourceType
	}

	prompt := fmt.Sprintf(`Extract a recipe from the following %s. Return the result as valid JSON only, with no additional text.

## Output Format

{
  "title": "Recipe title",
  "description": "1-2 sentence description of the dish",
  "ingredients_md": "Markdown bullet list of ingredients with quantities",
  "instructions_md": "Markdown numbered list of steps",
  "prep_time_minutes": <integer or null if unknown>,
  "cook_time_minutes": <integer or null if unknown>,
  "calories_per_serving": <integer or null if unknown>,
  "suggested_tags": ["tag1", "tag2"],
  "confidence": <0.0 to 1.0>,
  "confidence_notes": "Any issues, uncertainties, or assumptions made"
}

## Rules

### Ingredients (ingredients_md)
- Use markdown bullet list format: "- 250g flour"
- Include quantity, unit, and ingredient name
- Prefer metric units (g, ml, °C) but preserve original if clearly imperial
- One ingredient per line
- Include preparation notes in parentheses: "- 2 onions (finely diced)"

### Instructions (instructions_md)
- Use markdown numbered list: "1. Preheat oven to 180°C"
- Each step should be a single, clear action
- Preserve the original order
- Include temperatures, times, and visual cues where mentioned

### Metadata
- prep_time_minutes: Time for preparation before cooking starts
- cook_time_minutes: Active cooking/baking time
- calories_per_serving: Per single serving, if mentioned or calculable
- Use null if information is not available or cannot be reasonably inferred

### Tags
- Suggest 2-5 relevant tags based on the recipe
- Use lowercase, single words or hyphenated phrases
- Examples: "vegetarian", "quick-meal", "german", "dessert", "one-pot"

### Confidence
- 1.0: Perfect extraction, all information clear
- 0.8-0.9: Minor uncertainties (e.g., portion size unclear)
- 0.6-0.7: Some information missing or ambiguous
- Below 0.6: Significant issues, recommend manual review
- Always explain any uncertainties in confidence_notes

### Language
- Output in the same language as the source
- If source is German, output German text
- If source is English, output English text`, desc)

	if content != "" {
		prompt += fmt.Sprintf("\n\n## Source Content\n\n%s", content)
	}

	return prompt
}

func parseRecipeResponse(responseText string) (*ExtractedRecipe, error) {
	responseText = strings.TrimSpace(responseText)

	if strings.HasPrefix(responseText, "```json") {
		responseText = strings.TrimPrefix(responseText, "```json")
		responseText = strings.TrimSuffix(responseText, "```")
		responseText = strings.TrimSpace(responseText)
	} else if strings.HasPrefix(responseText, "```") {
		responseText = strings.TrimPrefix(responseText, "```")
		responseText = strings.TrimSuffix(responseText, "```")
		responseText = strings.TrimSpace(responseText)
	}

	var recipe ExtractedRecipe
	if err := json.Unmarshal([]byte(responseText), &recipe); err != nil {
		return nil, fmt.Errorf("failed to parse recipe JSON: %w (response: %s)", err, truncate(responseText, 200))
	}

	if recipe.Title == "" {
		return nil, fmt.Errorf("extracted recipe has no title")
	}

	return &recipe, nil
}

func detectMimeType(data []byte) string {
	if len(data) < 8 {
		return "application/octet-stream"
	}

	if data[0] == 0xFF && data[1] == 0xD8 && data[2] == 0xFF {
		return "image/jpeg"
	}

	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return "image/png"
	}

	if data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 {
		return "image/gif"
	}

	if (data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46) &&
		(data[8] == 0x57 && data[9] == 0x45 && data[10] == 0x42 && data[11] == 0x50) {
		return "image/webp"
	}

	return "image/jpeg"
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
