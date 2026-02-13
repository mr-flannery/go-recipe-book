package templates

import (
	"html/template"
	"io"
	"net/http"
)

type Renderer interface {
	Render(w io.Writer, name string, data any) error
	RenderPage(w http.ResponseWriter, name string, data any)
	RenderFragment(w http.ResponseWriter, name string, data any)
}

type renderer struct {
	templates *template.Template
}

func NewRenderer(templates *template.Template) Renderer {
	return &renderer{
		templates: templates,
	}
}

func (r *renderer) Render(w io.Writer, name string, data any) error {
	return r.templates.ExecuteTemplate(w, name, data)
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
