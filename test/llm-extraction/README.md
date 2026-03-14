# LLM Recipe Extraction Testing

Test prompts and compare LLM models for extracting recipes from images, websites, and video transcripts.

## Directory Structure

```
test/llm-extraction/
├── samples/           # Test input files (images, HTML, transcripts)
├── expected/          # Ground truth JSON for each sample
├── prompts/           # Prompt templates to test
├── results/           # Model outputs organized by model
│   ├── gpt-4o-mini/
│   ├── claude-3.5-sonnet/
│   ├── gemini-2.5-flash/
│   └── gemini-2.5-flash-lite/
└── scripts/           # Automation scripts
```

## Setup

### Prerequisites

```bash
# For YouTube transcript extraction
pip install yt-dlp

# For running the test automation (Go)
go mod tidy
```

### API Keys

Set environment variables for the models you want to test:

```bash
export OPENAI_API_KEY="sk-..."           # For GPT-4o-mini
export ANTHROPIC_API_KEY="sk-ant-..."    # For Claude 3.5 Sonnet
export GOOGLE_API_KEY="..."              # For Gemini 1.5 Flash
export MISTRAL_API_KEY="..."             # For Pixtral 12B, Mistral Small
```

Alternatively, use OpenRouter for all models:

```bash
export OPENROUTER_API_KEY="sk-or-..."
```

## Test Samples

### Sample Naming Convention

`{type}-{number}-{description}-{language}.{ext}`

Examples:
- `image-01-handwritten-en.jpg`
- `website-01-schema-en.html`
- `video-01-transcript-en.txt`

### Required Samples

| ID | Type | Description | Language | Status |
|----|------|-------------|----------|--------|
| image-01 | Image | Handwritten recipe card | EN | Placeholder |
| image-02 | Image | Cookbook page or printed recipe | DE | Placeholder |
| image-03 | Image | Screenshot from recipe website | EN | Placeholder |
| website-01 | Website | Recipe with schema.org markup | EN | Placeholder |
| website-02 | Website | Blog post with embedded recipe | DE | Placeholder |
| video-01 | Transcript | YouTube cooking video | EN | Placeholder |
| video-02 | Transcript | German cooking video | DE | Placeholder |

## Usage

### 1. Add Test Samples

Replace placeholder files in `samples/` with real test data:
- Images: JPG/PNG photos of recipe cards, cookbook pages, or screenshots
- Websites: Save HTML or use the extraction script
- Videos: Use the transcript extraction script

### 2. Extract YouTube Transcripts

```bash
./scripts/extract-transcript.sh "https://www.youtube.com/watch?v=VIDEO_ID" > samples/video-01-transcript-en.txt
```

### 3. Download YouTube Audio (for videos without captions)

For videos that don't have captions, you can download the audio and use Gemini's audio processing:

```bash
# Install yt-dlp if not already installed
pip install yt-dlp

# Download audio (MP3 format, 128kbps)
python scripts/download-video.py "https://www.youtube.com/watch?v=VIDEO_ID"

# Download with custom name
python scripts/download-video.py "https://www.youtube.com/watch?v=VIDEO_ID" video-03-cookies-en
```

Audio files are saved to `samples/` as MP3 files (typically 2-5MB vs 50-100MB for video). Gemini 2.5 Flash can process audio directly for transcription and recipe extraction.

### 4. Create Ground Truth

For each sample, create the expected JSON output in `expected/`:

```bash
# Copy the schema template and fill in manually
cp expected/_template.json expected/image-01-handwritten-en.json
# Edit with correct values
```

### 5. Run Extraction Tests

```bash
cd test/llm-extraction

# Test all samples against all models
go run ./scripts/run-extraction

# Test specific sample against specific model
go run ./scripts/run-extraction -sample image-01-handwritten-en -model gpt-4o-mini
```

### 6. Compare Results

```bash
cd test/llm-extraction

# Generate comparison table
go run ./scripts/compare-results
```

## Prompt Template

See `prompts/extract-recipe-v1.txt` for the current prompt template.

Key features:
- Outputs structured JSON matching the recipe model
- Includes confidence score (0.0-1.0) for flagging uncertain extractions
- Supports English and German
- Uses markdown for ingredients/instructions (matching existing data model)

## Expected Output Format

```json
{
  "title": "Recipe Title",
  "description": "1-2 sentence description",
  "ingredients_md": "- 250g flour\n- 2 eggs\n- 100ml milk",
  "instructions_md": "1. Mix dry ingredients\n2. Add wet ingredients\n3. Bake at 180C for 30 minutes",
  "prep_time_minutes": 15,
  "cook_time_minutes": 30,
  "calories_per_serving": 350,
  "suggested_tags": ["baking", "dessert"],
  "confidence": 0.9,
  "confidence_notes": "Baking time was slightly unclear in original"
}
```

## Evaluation Criteria

| Criterion | Weight | Description |
|-----------|--------|-------------|
| Title | 10% | Exact or semantic match |
| Ingredients completeness | 30% | % of ingredients captured |
| Ingredients accuracy | 20% | Quantities and units correct |
| Instructions completeness | 25% | All steps present, correct order |
| Metadata | 10% | Times, calories present and reasonable |
| Formatting | 5% | Valid markdown, consistent style |

## Models

| Model | Provider | Vision | Video | Approx Cost/Recipe |
|-------|----------|--------|-------|-------------------|
| GPT-4o-mini | OpenAI | Yes | No | ~$0.01-0.03 |
| Claude 3.5 Sonnet | Anthropic | Yes | No | ~$0.02-0.05 |
| Gemini 2.5 Flash | Google | Yes | Yes | ~$0.02-0.05 |
| Gemini 2.5 Flash Lite | Google | Yes | Yes | ~$0.01-0.02 |
| Pixtral 12B | Mistral | Yes | No | ~$0.01-0.02 |
| Mistral Small | Mistral | No | No | ~$0.01 |
| Mistral Large | Mistral | No | No | ~$0.02-0.04 |

## Notes

- For video extraction, we prefer transcripts when available (faster, cheaper)
- For videos without captions, Gemini 2.5 Flash can process video files directly
- Low volume expected (~10 recipes/month), so cost is not a major concern
- Focus on accuracy and trustworthiness of extractions
