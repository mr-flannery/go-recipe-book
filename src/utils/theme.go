package utils

import "net/http"

type Theme string

const (
	ThemeDefault   Theme = "editorial"
	ThemeEditorial Theme = "editorial"
)

func GetThemeFromRequest(r *http.Request) Theme {
	theme := r.URL.Query().Get("theme")
	switch theme {
	case "editorial":
		return ThemeEditorial
	default:
		return ThemeDefault
	}
}

func GetThemedTemplateName(baseName string, theme Theme) string {
	switch theme {
	case ThemeEditorial:
		return baseName[:len(baseName)-7] + "-editorial.gohtml"
	default:
		return baseName[:len(baseName)-7] + "-editorial.gohtml"
	}
}

func BuildURLWithTheme(path string, theme Theme) string {
	if len(path) > 0 && path[0] != '/' {
		path = "/" + path
	}
	return path
}
