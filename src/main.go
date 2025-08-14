package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/mr-flannery/go-recipe-book/src/auth"
	"github.com/mr-flannery/go-recipe-book/src/config"
	"github.com/mr-flannery/go-recipe-book/src/db"
	"github.com/mr-flannery/go-recipe-book/src/handlers"

	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}
	slog.Info("Starting server", "address", addr)
	slog.Info("Loading configuration...")
	_, err := config.GetConfig()
	if err != nil {
		slog.Error("Failed to load configuration", "error", err)
		panic(err)
	}

	slog.Info("Running migrations...")
	// Run database migrations
	db.RunMigrations()

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Home page
	mux.HandleFunc("/", handlers.HomeHandler)

	// Auth routes
	mux.HandleFunc("GET /login", handlers.GetLoginHandler)
	mux.HandleFunc("POST /login", handlers.PostLoginHandler)
	mux.HandleFunc("POST /logout", handlers.LogoutHandler)

	// Recipe routes with parameters
	mux.Handle("GET /recipes/create", auth.RequireAuth(http.HandlerFunc(handlers.GetCreateRecipeHandler)))
	mux.Handle("POST /recipes/create", auth.RequireAuth(http.HandlerFunc(handlers.PostCreateRecipeHandler)))
	mux.Handle("GET /recipes/update", auth.RequireAuth(http.HandlerFunc(handlers.GetUpdateRecipeHandler)))
	mux.Handle("POST /recipes/update", auth.RequireAuth(http.HandlerFunc(handlers.PostUpdateRecipeHandler)))
	mux.Handle("POST /recipes/delete", auth.RequireAuth(http.HandlerFunc(handlers.DeleteRecipeHandler)))
	mux.HandleFunc("GET /recipes", handlers.ListRecipesHandler)

	// Recipe view route with ID parameter - /recipes/{id}
	mux.HandleFunc("GET /recipes/{id}", handlers.ViewRecipeHandler)

	// Recipe comments route with ID parameter - /recipes/{id}/comments/htmx
	mux.Handle("POST /recipes/{id}/comments/htmx", auth.RequireAuth(http.HandlerFunc(handlers.CommentHTMXHandler)))

	slog.Info("Ready to serve!")

	slog.Error("Server failed to start", "error", http.ListenAndServe(addr, mux))
}
