package main

import (
	"log/slog"
	"net/http"
	"os"

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

	slog.Info("Ready to serve!")

	slog.Error("Server failed to start", "error", http.ListenAndServe(addr, nil))
}
