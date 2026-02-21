package markdown

import (
	"strings"
	"testing"
)

func TestRender_ConvertsMarkdownToHTML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:     "simple paragraph",
			input:    "Hello world",
			contains: []string{"<p>Hello world</p>"},
		},
		{
			name:     "headers",
			input:    "# Heading 1\n\n## Heading 2",
			contains: []string{"<h1", "Heading 1", "<h2", "Heading 2"},
		},
		{
			name:     "bold text",
			input:    "This is **bold** text",
			contains: []string{"<strong>bold</strong>"},
		},
		{
			name:     "unordered list",
			input:    "- item 1\n- item 2",
			contains: []string{"<ul>", "<li>item 1</li>", "<li>item 2</li>", "</ul>"},
		},
		{
			name:     "ordered list",
			input:    "1. first\n2. second",
			contains: []string{"<ol>", "<li>first</li>", "<li>second</li>", "</ol>"},
		},
		{
			name:     "links",
			input:    "[example](https://example.com)",
			contains: []string{`<a href="https://example.com">example</a>`},
		},
		{
			name:     "GFM strikethrough",
			input:    "~~deleted~~",
			contains: []string{"<del>deleted</del>"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.input)
			if err != nil {
				t.Fatalf("Render failed: %v", err)
			}

			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("expected result to contain %q, got %q", expected, result)
				}
			}
		})
	}
}

func TestRender_ProcessesIngredientSyntax(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantContains []string
	}{
		{
			name:  "single ingredient",
			input: "@ingredient{flour|2 cups}",
			wantContains: []string{
				`<span class="ingredient" data-name="flour">2 cups flour</span>`,
			},
		},
		{
			name:  "ingredient with spaces in name",
			input: "@ingredient{olive oil|3 tbsp}",
			wantContains: []string{
				`<span class="ingredient" data-name="olive oil">3 tbsp olive oil</span>`,
			},
		},
		{
			name:  "multiple ingredients",
			input: "@ingredient{sugar|1 cup} and @ingredient{butter|2 tbsp}",
			wantContains: []string{
				`data-name="sugar">1 cup sugar</span>`,
				`data-name="butter">2 tbsp butter</span>`,
			},
		},
		{
			name:  "ingredient in list",
			input: "- @ingredient{eggs|3}",
			wantContains: []string{
				`<span class="ingredient" data-name="eggs">3 eggs</span>`,
			},
		},
		{
			name:  "trims whitespace in name and quantity",
			input: "@ingredient{ salt | 1 tsp }",
			wantContains: []string{
				`data-name="salt">1 tsp salt</span>`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.input)
			if err != nil {
				t.Fatalf("Render failed: %v", err)
			}

			for _, expected := range tt.wantContains {
				if !strings.Contains(result, expected) {
					t.Errorf("expected result to contain %q, got %q", expected, result)
				}
			}
		})
	}
}

func TestRender_ProcessesWikilinks(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		wantContains []string
	}{
		{
			name:  "simple wikilink",
			input: "See [[Chocolate Cake]]",
			wantContains: []string{
				`href="/recipes/chocolate-cake"`,
				`>Chocolate Cake</a>`,
			},
		},
		{
			name:  "wikilink with special chars",
			input: "Try [[Mom's Apple Pie]]",
			wantContains: []string{
				`href="/recipes/mom-s-apple-pie"`,
			},
		},
		{
			name:  "multiple wikilinks",
			input: "Pair [[Pasta]] with [[Garlic Bread]]",
			wantContains: []string{
				`href="/recipes/pasta"`,
				`href="/recipes/garlic-bread"`,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Render(tt.input)
			if err != nil {
				t.Fatalf("Render failed: %v", err)
			}

			for _, expected := range tt.wantContains {
				if !strings.Contains(result, expected) {
					t.Errorf("expected result to contain %q, got %q", expected, result)
				}
			}
		})
	}
}

func TestRender_CombinesIngredientAndWikilink(t *testing.T) {
	input := "Use @ingredient{butter|2 tbsp} from [[Homemade Butter]]"

	result, err := Render(input)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}

	if !strings.Contains(result, `data-name="butter">2 tbsp butter</span>`) {
		t.Errorf("expected ingredient span, got %q", result)
	}
	if !strings.Contains(result, `href="/recipes/homemade-butter"`) {
		t.Errorf("expected wikilink href, got %q", result)
	}
}

func TestSlugify_ConvertsStringsToURLSlugs(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Simple", "simple"},
		{"Two Words", "two-words"},
		{"Chocolate Cake", "chocolate-cake"},
		{"Mom's Apple Pie", "mom-s-apple-pie"},
		{"ALL CAPS", "all-caps"},
		{"  extra  spaces  ", "extra-spaces"},
		{"Special!@#Characters", "special-characters"},
		{"Numbers123", "numbers123"},
		{"---dashes---", "dashes"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := slugify(tt.input)
			if got != tt.want {
				t.Errorf("slugify(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestProcessIngredients_HandlesEdgeCases(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "no ingredients",
			input: "plain text",
			want:  "plain text",
		},
		{
			name:  "incomplete syntax - no closing brace",
			input: "@ingredient{flour|2 cups",
			want:  "@ingredient{flour|2 cups",
		},
		{
			name:  "incomplete syntax - no pipe",
			input: "@ingredient{flour}",
			want:  "@ingredient{flour}",
		},
		{
			name:  "valid syntax",
			input: "@ingredient{flour|2 cups}",
			want:  `<span class="ingredient" data-name="flour">2 cups flour</span>`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := processIngredients(tt.input)
			if got != tt.want {
				t.Errorf("processIngredients(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
