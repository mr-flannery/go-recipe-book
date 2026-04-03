package extraction

import (
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
)

const maxContentSize = 5 * 1024 * 1024 // 5MB

func FetchWebsiteContent(websiteURL string) (string, error) {
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

	client := &http.Client{
		Timeout: 30 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return errors.New("too many redirects")
			}
			return nil
		},
	}

	req, err := http.NewRequest("GET", websiteURL, nil)
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrFetchFailed, err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,de;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		return "", technicalErrorf("%w: %v", ErrFetchFailed, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode >= 400 && resp.StatusCode < 500 {
			return "", fmt.Errorf("%w: status %d", ErrFetchFailed, resp.StatusCode)
		} else {
			return "", technicalErrorf("%w: status %d", ErrFetchFailed, resp.StatusCode)
		}
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/html") && !strings.Contains(contentType, "application/xhtml") {
		return "", fmt.Errorf("%w: not an HTML page", ErrFetchFailed)
	}

	limitedReader := io.LimitReader(resp.Body, maxContentSize+1)
	body, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", technicalErrorf("%w: %v", ErrFetchFailed, err)
	}

	if len(body) > maxContentSize {
		return "", ErrContentTooLarge
	}

	return ExtractTextContent(string(body)), nil
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
