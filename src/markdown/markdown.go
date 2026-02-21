package markdown

import (
	"bytes"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"go.abhg.dev/goldmark/wikilink"
)

type recipeResolver struct{}

func (r *recipeResolver) ResolveWikilink(n *wikilink.Node) ([]byte, error) {
	slug := slugify(string(n.Target))
	return []byte("/recipes/" + slug), nil
}

var md goldmark.Markdown

func init() {
	md = goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			&wikilink.Extender{
				Resolver: &recipeResolver{},
			},
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
			html.WithUnsafe(),
		),
	)
}

func Render(source string) (string, error) {
	processed := processIngredients(source)

	var buf bytes.Buffer
	if err := md.Convert([]byte(processed), &buf); err != nil {
		return "", err
	}

	return buf.String(), nil
}

var ingredientRegex = regexp.MustCompile(`@ingredient\{([^|]+)\|([^}]+)\}`)

func processIngredients(source string) string {
	return ingredientRegex.ReplaceAllStringFunc(source, func(match string) string {
		parts := ingredientRegex.FindStringSubmatch(match)
		if len(parts) != 3 {
			return match
		}
		name := strings.TrimSpace(parts[1])
		quantity := strings.TrimSpace(parts[2])
		return `<span class="ingredient" data-name="` + name + `">` + quantity + " " + name + `</span>`
	})
}

var nonAlphanumericRegex = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(s string) string {
	s = strings.ToLower(s)
	s = nonAlphanumericRegex.ReplaceAllString(s, "-")
	s = strings.Trim(s, "-")
	return s
}
