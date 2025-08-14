package main

import (
	"log/slog"
	"net/http"
	"os"
	"strings"

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

	// Home page
	http.HandleFunc("/", handlers.HomeHandler)

	// Auth routes
	http.HandleFunc("/login", handlers.LoginHandler)
	http.HandleFunc("/logout", handlers.LogoutHandler)

	// Recipe routes - order matters! More specific routes first
	http.Handle("/recipes/create", auth.RequireAuth(http.HandlerFunc(handlers.CreateRecipeHandler)))
	http.Handle("/recipes/update", auth.RequireAuth(http.HandlerFunc(handlers.UpdateRecipeHandler)))
	http.Handle("/recipes/delete", auth.RequireAuth(http.HandlerFunc(handlers.DeleteRecipeHandler)))
	http.HandleFunc("/recipes", handlers.ListRecipesHandler)

	// Recipe view and comment routes - handle /recipes/{id}, /recipes/{id}/comments, and /recipes/{id}/comments/htmx
	http.HandleFunc("/recipes/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/comments/htmx") {
			// Comments require authentication
			auth.RequireAuth(http.HandlerFunc(handlers.CommentHTMXHandler)).ServeHTTP(w, r)
		} else {
			// Recipe viewing doesn't require authentication
			handlers.ViewRecipeHandler(w, r)
		}
	})

	slog.Info("Ready to serve!")

	slog.Error("Server failed to start", "error", http.ListenAndServe(addr, nil))
}
