package extraction

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var (
	ErrNoCaptions        = errors.New("this video does not have captions/subtitles available")
	ErrVideoUnavailable  = errors.New("video is unavailable or does not exist")
	ErrInvalidYouTubeURL = errors.New("invalid YouTube URL")
)

type VideoMetadata struct {
	Title       string
	Description string
	RecipeLinks []string
}

type TranscriptSegment struct {
	Text     string
	Start    float64
	Duration float64
}

var videoIDRegexps = []*regexp.Regexp{
	regexp.MustCompile(`(?:youtube\.com/watch\?v=|youtu\.be/|youtube\.com/embed/|youtube\.com/v/)([a-zA-Z0-9_-]{11})`),
	regexp.MustCompile(`^([a-zA-Z0-9_-]{11})$`),
}

func ExtractVideoID(input string) (string, error) {
	input = strings.TrimSpace(input)
	for _, re := range videoIDRegexps {
		if matches := re.FindStringSubmatch(input); len(matches) > 1 {
			return matches[1], nil
		}
	}
	return "", ErrInvalidYouTubeURL
}

func FetchYouTubeTranscript(videoURL string, preferredLangs []string) (string, error) {
	videoID, err := ExtractVideoID(videoURL)
	if err != nil {
		return "", err
	}

	if len(preferredLangs) == 0 {
		preferredLangs = []string{"en", "de"}
	}

	captionURL, err := getCaptionTrackURL(videoID, preferredLangs)
	if err != nil {
		return "", err
	}

	segments, err := fetchAndParseTranscript(captionURL)
	if err != nil {
		return "", err
	}

	return formatTranscriptAsText(segments), nil
}

func getCaptionTrackURL(videoID string, preferredLangs []string) (string, error) {
	client := &http.Client{Timeout: 30 * time.Second}

	watchURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
	req, err := http.NewRequest("GET", watchURL, nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := client.Do(req)
	if err != nil {
		return "", technicalErrorf("failed to fetch video page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", technicalErrorf("video page returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read video page: %w", err)
	}

	pageContent := string(body)

	if strings.Contains(pageContent, "This video isn't available") {
		return "", ErrVideoUnavailable
	}

	captionTracks, err := extractCaptionTracks(pageContent)
	if err != nil {
		return "", err
	}

	if len(captionTracks) == 0 {
		return "", ErrNoCaptions
	}

	selectedTrack := selectBestTrack(captionTracks, preferredLangs)
	if selectedTrack == nil {
		selectedTrack = &captionTracks[0]
	}

	return selectedTrack.BaseURL, nil
}

type captionTrack struct {
	BaseURL      string `json:"baseUrl"`
	LanguageCode string `json:"languageCode"`
	Kind         string `json:"kind"`
	IsTranslated bool   `json:"isTranslatable"`
}

func extractCaptionTracks(pageContent string) ([]captionTrack, error) {
	captionsRegexp := regexp.MustCompile(`"captions"\s*:\s*(\{[^}]*"playerCaptionsTracklistRenderer"[^}]*\})`)
	matches := captionsRegexp.FindStringSubmatch(pageContent)

	if len(matches) < 2 {
		tracklistRegexp := regexp.MustCompile(`"captionTracks"\s*:\s*(\[[^\]]*\])`)
		trackMatches := tracklistRegexp.FindStringSubmatch(pageContent)
		if len(trackMatches) < 2 {
			return nil, ErrNoCaptions
		}

		var tracks []captionTrack
		if err := json.Unmarshal([]byte(trackMatches[1]), &tracks); err != nil {
			return nil, fmt.Errorf("failed to parse caption tracks: %w", err)
		}
		return tracks, nil
	}

	tracksJSON := matches[1]
	tracklistRegexp := regexp.MustCompile(`"captionTracks"\s*:\s*(\[[^\]]*\])`)
	trackMatches := tracklistRegexp.FindStringSubmatch(tracksJSON)
	if len(trackMatches) < 2 {
		return nil, ErrNoCaptions
	}

	var tracks []captionTrack
	if err := json.Unmarshal([]byte(trackMatches[1]), &tracks); err != nil {
		return nil, fmt.Errorf("failed to parse caption tracks: %w", err)
	}

	return tracks, nil
}

func selectBestTrack(tracks []captionTrack, preferredLangs []string) *captionTrack {
	for _, lang := range preferredLangs {
		for i := range tracks {
			if tracks[i].LanguageCode == lang && tracks[i].Kind != "asr" {
				return &tracks[i]
			}
		}
	}

	for _, lang := range preferredLangs {
		for i := range tracks {
			if tracks[i].LanguageCode == lang {
				return &tracks[i]
			}
		}
	}

	for _, lang := range preferredLangs {
		for i := range tracks {
			if strings.HasPrefix(tracks[i].LanguageCode, lang) {
				return &tracks[i]
			}
		}
	}

	return nil
}

type xmlTranscript struct {
	XMLName xml.Name  `xml:"transcript"`
	Texts   []xmlText `xml:"text"`
}

type xmlText struct {
	Start    string `xml:"start,attr"`
	Duration string `xml:"dur,attr"`
	Text     string `xml:",chardata"`
}

func fetchAndParseTranscript(captionURL string) ([]TranscriptSegment, error) {
	parsedURL, err := url.Parse(captionURL)
	if err != nil {
		return nil, err
	}
	query := parsedURL.Query()
	query.Set("fmt", "srv3")
	parsedURL.RawQuery = query.Encode()

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(parsedURL.String())
	if err != nil {
		return nil, technicalErrorf("failed to fetch transcript: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, technicalErrorf("transcript fetch returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read transcript: %w", err)
	}

	var transcript xmlTranscript
	if err := xml.Unmarshal(body, &transcript); err != nil {
		return nil, fmt.Errorf("failed to parse transcript XML: %w", err)
	}

	var segments []TranscriptSegment
	for _, text := range transcript.Texts {
		var start, dur float64
		fmt.Sscanf(text.Start, "%f", &start)
		fmt.Sscanf(text.Duration, "%f", &dur)

		cleanedText := html.UnescapeString(text.Text)
		cleanedText = strings.TrimSpace(cleanedText)

		if cleanedText != "" {
			segments = append(segments, TranscriptSegment{
				Text:     cleanedText,
				Start:    start,
				Duration: dur,
			})
		}
	}

	return segments, nil
}

func formatTranscriptAsText(segments []TranscriptSegment) string {
	var builder strings.Builder
	var prevText string

	for _, seg := range segments {
		if seg.Text != prevText {
			if builder.Len() > 0 {
				builder.WriteString("\n")
			}
			builder.WriteString(seg.Text)
			prevText = seg.Text
		}
	}

	return builder.String()
}

func FetchVideoMetadata(videoURL string) (*VideoMetadata, error) {
	videoID, err := ExtractVideoID(videoURL)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 30 * time.Second}

	watchURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", videoID)
	req, err := http.NewRequest("GET", watchURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := client.Do(req)
	if err != nil {
		return nil, technicalErrorf("failed to fetch video page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, technicalErrorf("video page returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read video page: %w", err)
	}

	pageContent := string(body)

	if strings.Contains(pageContent, "Video unavailable") ||
		strings.Contains(pageContent, "This video isn't available") {
		return nil, ErrVideoUnavailable
	}

	return extractMetadataFromPage(pageContent)
}

func extractMetadataFromPage(pageContent string) (*VideoMetadata, error) {
	metadata := &VideoMetadata{}

	titleRegexp := regexp.MustCompile(`"title"\s*:\s*"([^"]*)"`)
	if matches := titleRegexp.FindStringSubmatch(pageContent); len(matches) > 1 {
		metadata.Title = unescapeJSONString(matches[1])
	}

	descRegexp := regexp.MustCompile(`"shortDescription"\s*:\s*"([^"]*)"`)
	if matches := descRegexp.FindStringSubmatch(pageContent); len(matches) > 1 {
		metadata.Description = unescapeJSONString(matches[1])
	}

	if metadata.Description != "" {
		metadata.RecipeLinks = extractRecipeLinks(metadata.Description)
	}

	return metadata, nil
}

func unescapeJSONString(s string) string {
	s = strings.ReplaceAll(s, "\\n", "\n")
	s = strings.ReplaceAll(s, "\\r", "\r")
	s = strings.ReplaceAll(s, "\\t", "\t")
	s = strings.ReplaceAll(s, "\\\"", "\"")
	s = strings.ReplaceAll(s, "\\/", "/")
	s = strings.ReplaceAll(s, "\\\\", "\\")
	return s
}

var recipeURLPatterns = []*regexp.Regexp{
	regexp.MustCompile(`https?://[^\s<>"]+(?:recipe|rezept)[^\s<>"]*`),
	regexp.MustCompile(`https?://[^\s<>"]*(?:allrecipes|food\.com|epicurious|seriouseats|bonappetit|delish|tasty|cookpad|chefkoch|kitchen)[^\s<>"]*`),
}

func extractRecipeLinks(description string) []string {
	seen := make(map[string]bool)
	var links []string

	for _, pattern := range recipeURLPatterns {
		matches := pattern.FindAllString(description, -1)
		for _, match := range matches {
			match = strings.TrimRight(match, ".,;:!?)")
			if !seen[match] {
				seen[match] = true
				links = append(links, match)
			}
		}
	}

	return links
}
