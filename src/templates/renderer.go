package templates

import (
	"html/template"
	"io"
	"net/http"

	"github.com/mr-flannery/go-recipe-book/src/auth"
)

type Renderer interface {
	Render(w io.Writer, name string, data any) error
	RenderPage(w http.ResponseWriter, name string, data any)
	RenderFragment(w http.ResponseWriter, name string, data any)
	RenderError(w http.ResponseWriter, r *http.Request, statusCode int, message string)
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

type ErrorPageData struct {
	StatusCode int
	Title      string
	Message    string
	UserInfo   *auth.UserInfo
}

func statusTitle(code int) string {
	switch code {
	case http.StatusBadRequest:
		return "Bad Request"
	case http.StatusUnauthorized:
		return "Unauthorized"
	case http.StatusForbidden:
		return "Access Denied"
	case http.StatusNotFound:
		return "Not Found"
	case http.StatusInternalServerError:
		return "Something Went Wrong"
	default:
		return "Error"
	}
}

func (r *renderer) RenderError(w http.ResponseWriter, req *http.Request, statusCode int, message string) {
	data := ErrorPageData{
		StatusCode: statusCode,
		Title:      statusTitle(statusCode),
		Message:    message,
		UserInfo:   auth.GetUserInfoFromContext(req.Context()),
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(statusCode)
	if err := r.Render(w, "error.gohtml", data); err != nil {
		http.Error(w, message, statusCode)
	}
}
