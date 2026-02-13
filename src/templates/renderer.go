package templates

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strings"
)

type Renderer interface {
	Render(w io.Writer, name string, data any) error
	RenderPage(w http.ResponseWriter, name string, data any)
	RenderFragment(w http.ResponseWriter, name string, data any)
}

type renderer struct {
	templates *template.Template
	theme     string
}

func NewRenderer(templates *template.Template) Renderer {
	return &renderer{
		templates: templates,
		theme:     "editorial",
	}
}

func (r *renderer) getThemedName(name string) string {
	if !strings.HasSuffix(name, ".gohtml") {
		return name
	}
	baseName := name[:len(name)-7]
	return fmt.Sprintf("%s-%s.gohtml", baseName, r.theme)
}

func (r *renderer) Render(w io.Writer, name string, data any) error {
	templateName := r.getThemedName(name)
	return r.templates.ExecuteTemplate(w, templateName, data)
}

func (r *renderer) RenderPage(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := r.Render(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (r *renderer) RenderFragment(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html")
	if err := r.templates.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
