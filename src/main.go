package main

import (
	"log/slog"
	"net/http"
	"os"
	"strings"

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

	// Home page
	http.HandleFunc("/", handlers.HomeHandler)

	// Auth routes
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/logout", handlers.LogoutHandler)

	// Recipe routes
	http.HandleFunc("/recipes", handlers.ListRecipesHandler)
	http.HandleFunc("/recipes/create", handlers.CreateRecipeHandler)
	http.HandleFunc("/recipes/update", handlers.UpdateRecipeHandler)
	http.HandleFunc("/recipes/delete", handlers.DeleteRecipeHandler)

	// Recipe view and comment routes - handle /recipes/{id}, /recipes/{id}/comments, and /recipes/{id}/comments/htmx
	http.HandleFunc("/recipes/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/comments/htmx") {
			handlers.CommentHTMXHandler(w, r)
		} else {
			handlers.ViewRecipeHandler(w, r)
		}
	})

	slog.Info("Ready to serve!")

	slog.Error("Server failed to start", "error", http.ListenAndServe(addr, nil))
}
