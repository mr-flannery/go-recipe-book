package templates

import (
	"html/template"
	"io/fs"
	"log/slog"
	"net/url"
	"path/filepath"
	"strings"

	"github.com/mr-flannery/go-recipe-book/src/markdown"
	"github.com/mr-flannery/go-recipe-book/src/models"
	"github.com/mr-flannery/go-recipe-book/src/utils"
)

var themeStylesheets = map[string]string{
	models.ThemeEditorial:   "/static/css/editorial.css",
	models.ThemeClassic:     "/static/css/styles.css",
	models.ThemeDiner:       "/static/css/diner.css",
	models.ThemeTrattoria:   "/static/css/trattoria.css",
	models.ThemeKuche:       "/static/css/kuche.css",
	models.ThemeNightowl:    "/static/css/nightowl.css",
	models.ThemeMilkbar:     "/static/css/milkbar.css",
	models.ThemeBodega:      "/static/css/bodega.css",
	models.ThemeMarket:      "/static/css/market.css",
	models.ThemeBistro:      "/static/css/bistro.css",
	models.ThemeComfort:     "/static/css/comfort.css",
	models.ThemeSpeakeasy:   "/static/css/speakeasy.css",
	models.ThemeCuchifritos: "/static/css/cuchifritos.css",
	models.ThemePizzeria:    "/static/css/pizzeria.css",
}

func stylesheetForTheme(theme string) string {
	if path, ok := themeStylesheets[theme]; ok {
		return path
	}
	return themeStylesheets[models.DefaultTheme]
}

var funcMap = template.FuncMap{
	"stylesheet": stylesheetForTheme,
	"dict": func(values ...any) map[string]any {
		if len(values)%2 != 0 {
			panic("dict requires even number of arguments")
		}
		m := make(map[string]any, len(values)/2)
		for i := 0; i < len(values); i += 2 {
			key, ok := values[i].(string)
			if !ok {
				panic("dict keys must be strings")
			}
			m[key] = values[i+1]
		}
		return m
	},
	"joinTagNames": func(tags []models.Tag) string {
		names := make([]string, len(tags))
		for i, t := range tags {
			names[i] = t.Name
		}
		return strings.Join(names, ",")
	},
	"renderMarkdown": func(source string) template.HTML {
		html, err := markdown.Render(source)
		if err != nil {
			slog.Error("Failed to render markdown", "error", err)
			return template.HTML(template.HTMLEscapeString(source))
		}
		return template.HTML(html)
	},
	"add": func(a, b int) int {
		return a + b
	},
	"subtract": func(a, b int) int {
		return a - b
	},
	"hasPrefix": strings.HasPrefix,
	"renderSource": func(source string) template.HTML {
		if source == "" {
			return ""
		}
		source = strings.TrimSpace(source)
		parsed, err := url.Parse(source)
		if err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") {
			escaped := template.HTMLEscapeString(source)
			return template.HTML(`<a href="` + escaped + `" target="_blank" rel="noopener noreferrer">` + escaped + `</a>`)
		}
		return template.HTML(template.HTMLEscapeString(source))
	},
}

func loadTemplates(root string) (*template.Template, error) {
	var files []string

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !d.IsDir() && strings.HasSuffix(path, ".gohtml") {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return template.New("").Funcs(funcMap).ParseFiles(files...)
}

var Templates = template.Must(loadTemplates(filepath.Join(utils.GetBasePath(), "src", "templates")))
