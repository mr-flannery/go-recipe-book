package extraction

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

var (
	ErrInvalidURL      = errors.New("invalid URL")
	ErrFetchFailed     = errors.New("failed to fetch website")
	ErrContentTooLarge = errors.New("content too large")
	ErrNoArchive       = errors.New("no archive found")
)

const maxContentSize = 5 * 1024 * 1024 // 5MB

func newHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return errors.New("too many redirects")
			}
			return nil
		},
	}
}

// FetchWebsiteContent fetches and extracts text from the given URL.
// If the direct fetch fails for any reason, it falls back to the most
// recent Wayback Machine snapshot.
// The second return value is the URL that was actually used (may differ from
// the input when the Wayback Machine fallback is used).
func FetchWebsiteContent(websiteURL string) (content string, usedURL string, err error) {
	content, directErr := fetchURL(websiteURL)
	if directErr == nil {
		return content, websiteURL, nil
	}

	// Try Wayback Machine as fallback.
	archiveURL, archiveErr := lookupWaybackURL(websiteURL)
	if archiveErr != nil {
		return "", "", fmt.Errorf("%w (wayback lookup also failed: %v)", directErr, archiveErr)
	}

	content, archiveFetchErr := fetchURL(archiveURL)
	if archiveFetchErr != nil {
		return "", "", fmt.Errorf("%w (wayback fetch also failed: %v)", directErr, archiveFetchErr)
	}

	return content, archiveURL, nil
}

// fetchURL performs the actual HTTP fetch of a single URL and returns
// the extracted text content. All errors include the URL for context.
func fetchURL(websiteURL string) (string, error) {
	parsedURL, err := url.Parse(websiteURL)
	if err != nil {
		return "", ErrInvalidURL
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return "", ErrInvalidURL
	}

	if parsedURL.Host == "" {
		return "", ErrInvalidURL
	}

	client := newHTTPClient()

	req, err := http.NewRequest("GET", websiteURL, nil)
	if err != nil {
		return "", fmt.Errorf("%w %s: %v", ErrFetchFailed, websiteURL, err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,de;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return "", technicalErrorf("%w %s: %v", ErrFetchFailed, websiteURL, err)
	}
	defer resp.Body.Close()

	finalURL := resp.Request.URL.String()
	urlContext := websiteURL
	if finalURL != websiteURL {
		urlContext = fmt.Sprintf("%s (redirected to %s)", websiteURL, finalURL)
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return "", fmt.Errorf("%w %s: status %d", ErrFetchFailed, urlContext, resp.StatusCode)
		}
		return "", technicalErrorf("%w %s: status %d", ErrFetchFailed, urlContext, resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") && !strings.Contains(contentType, "application/xhtml") {
		return "", fmt.Errorf("%w %s: not an HTML page (content-type: %s)", ErrFetchFailed, urlContext, contentType)
	}

	limitedReader := io.LimitReader(resp.Body, maxContentSize+1)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", technicalErrorf("%w %s: %v", ErrFetchFailed, urlContext, err)
	}

	if len(body) > maxContentSize {
		return "", ErrContentTooLarge
	}

	return ExtractTextContent(string(body)), nil
}

// lookupWaybackURL queries the Wayback Machine CDX API for the most recent
// successful snapshot of the given URL and returns the playback URL.
func lookupWaybackURL(originalURL string) (string, error) {
	cdxURL := fmt.Sprintf(
		"https://web.archive.org/cdx/search/cdx?url=%s&output=json&limit=1&fl=timestamp,statuscode&filter=statuscode:200&from=20200101&to=&collapse=digest&fastLatest=true",
		url.QueryEscape(originalURL),
	)

	client := newHTTPClient()
	resp, err := client.Get(cdxURL)
	if err != nil {
		return "", fmt.Errorf("CDX API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("CDX API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	if err != nil {
		return "", fmt.Errorf("failed to read CDX API response: %w", err)
	}

	// CDX JSON format: [["timestamp","statuscode"], ["20240101120000","200"], ...]
	// First element is the header row.
	var rows [][]string
	if err := json.Unmarshal(body, &rows); err != nil {
		return "", fmt.Errorf("failed to parse CDX API response: %w", err)
	}

	// rows[0] is the header; rows[1] is the first (most recent) result.
	if len(rows) < 2 {
		return "", ErrNoArchive
	}

	timestamp := rows[1][0]
	archiveURL := fmt.Sprintf("https://web.archive.org/web/%sid_/%s", timestamp, originalURL)
	return archiveURL, nil
}

func ExtractTextContent(htmlContent string) string {
	content := htmlContent

	content = removeHTMLComments(content)
	content = removeTag(content, "script")
	content = removeTag(content, "style")
	content = removeTag(content, "nav")
	content = removeTag(content, "header")
	content = removeTag(content, "footer")
	content = removeTag(content, "aside")
	content = removeTag(content, "noscript")

	mainContent := extractMainContent(content)
	if mainContent != "" {
		content = mainContent
	}

	content = stripHTMLTags(content)
	content = decodeHTMLEntities(content)
	content = normalizeWhitespace(content)

	return content
}

func removeHTMLComments(content string) string {
	re := regexp.MustCompile(`<!--[\s\S]*?-->`)
	return re.ReplaceAllString(content, "")
}

func removeTag(content string, tagName string) string {
	re := regexp.MustCompile(fmt.Sprintf(`(?i)<%s[^>]*>[\s\S]*?</%s>`, tagName, tagName))
	return re.ReplaceAllString(content, "")
}

func extractMainContent(content string) string {
	patterns := []string{
		`(?i)<article[^>]*>([\s\S]*?)</article>`,
		`(?i)<main[^>]*>([\s\S]*?)</main>`,
		`(?i)<div[^>]*class="[^"]*recipe[^"]*"[^>]*>([\s\S]*?)</div>`,
		`(?i)<div[^>]*class="[^"]*content[^"]*"[^>]*>([\s\S]*?)</div>`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		if matches := re.FindStringSubmatch(content); len(matches) > 1 {
			return matches[1]
		}
	}

	return ""
}

func stripHTMLTags(content string) string {
	blockTags := regexp.MustCompile(`(?i)</(p|div|br|h[1-6]|li|tr)>`)
	content = blockTags.ReplaceAllString(content, "\n")

	brTags := regexp.MustCompile(`(?i)<br\s*/?>`)
	content = brTags.ReplaceAllString(content, "\n")

	allTags := regexp.MustCompile(`<[^>]+>`)
	return allTags.ReplaceAllString(content, "")
}

func decodeHTMLEntities(content string) string {
	replacements := map[string]string{
		"&nbsp;":   " ",
		"&amp;":    "&",
		"&lt;":     "<",
		"&gt;":     ">",
		"&quot;":   "\"",
		"&#39;":    "'",
		"&apos;":   "'",
		"&ndash;":  "-",
		"&mdash;":  "-",
		"&deg;":    "°",
		"&frac12;": "½",
		"&frac14;": "¼",
		"&frac34;": "¾",
	}

	for entity, replacement := range replacements {
		content = strings.ReplaceAll(content, entity, replacement)
	}

	numericEntity := regexp.MustCompile(`&#(\d+);`)
	content = numericEntity.ReplaceAllStringFunc(content, func(match string) string {
		var num int
		fmt.Sscanf(match, "&#%d;", &num)
		if num > 0 && num < 65536 {
			return string(rune(num))
		}
		return match
	})

	return content
}

func normalizeWhitespace(content string) string {
	multipleSpaces := regexp.MustCompile(`[ \t]+`)
	content = multipleSpaces.ReplaceAllString(content, " ")

	multipleNewlines := regexp.MustCompile(`\n{3,}`)
	content = multipleNewlines.ReplaceAllString(content, "\n\n")

	lines := strings.Split(content, "\n")
	var trimmedLines []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" {
			trimmedLines = append(trimmedLines, trimmed)
		}
	}

	return strings.Join(trimmedLines, "\n")
}
