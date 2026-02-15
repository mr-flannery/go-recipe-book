package mocks

import (
	"io"
	"net/http"
)

type MockRenderer struct {
	RenderFunc         func(w io.Writer, name string, data any) error
	RenderPageFunc     func(w http.ResponseWriter, name string, data any)
	RenderFragmentFunc func(w http.ResponseWriter, name string, data any)
	RenderErrorFunc    func(w http.ResponseWriter, r *http.Request, statusCode int, message string)
}

func (m *MockRenderer) Render(w io.Writer, name string, data any) error {
	if m.RenderFunc != nil {
		return m.RenderFunc(w, name, data)
	}
	return nil
}

func (m *MockRenderer) RenderPage(w http.ResponseWriter, name string, data any) {
	if m.RenderPageFunc != nil {
		m.RenderPageFunc(w, name, data)
	}
}

func (m *MockRenderer) RenderFragment(w http.ResponseWriter, name string, data any) {
	if m.RenderFragmentFunc != nil {
		m.RenderFragmentFunc(w, name, data)
	}
}

func (m *MockRenderer) RenderError(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
	if m.RenderErrorFunc != nil {
		m.RenderErrorFunc(w, r, statusCode, message)
	} else {
		w.WriteHeader(statusCode)
	}
}
