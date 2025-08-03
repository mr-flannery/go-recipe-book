package main

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/mr-flannery/go-recipe-book/src/db"
	"github.com/mr-flannery/go-recipe-book/src/handlers"
	"github.com/mr-flannery/go-recipe-book/src/models"

	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}
	slog.Info("Starting server", "address", addr)

	slog.Info("Initializing database...")

	dataSourceName := "host=localhost port=5432 user=local-recipe-user password=local-recipe-password dbname=recipe-book sslmode=disable"

	err := models.InitializeDB(dataSourceName)
	if err != nil {
		slog.Error("Failed to initialize database", "error", err)
		panic(err)
	}

	slog.Info("Running migrations...")
	// Run database migrations
	db.RunMigrations(dataSourceName)

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
