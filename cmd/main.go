package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
)

func main() {
	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}
	log.Printf("Starting server on %s", addr)

	// Home page
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Use the home handler from internal/handlers
		// Lazy import to avoid circular import for now
		importHandlersHome(w, r)
	})

	log.Fatal(http.ListenAndServe(addr, nil))
}

// importHandlersHome calls the Home handler from internal/handlers
func importHandlersHome(w http.ResponseWriter, r *http.Request) {
	// Import the Home handler from the handlers package
	// This import path assumes your module name is github.com/yourusername/agent-coding-recipe-book
	// Adjust the import path if your module name is different
	//
	// import "github.com/yourusername/agent-coding-recipe-book/internal/handlers"
	// handlers.Home(w, r)

	// To avoid import issues in this patch, inline the handler logic for now
	// Replace this with a direct call to handlers.Home when possible
	templates := template.Must(template.ParseGlob("templates/*.gohtml"))
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	err := templates.ExecuteTemplate(w, "home.gohtml", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
