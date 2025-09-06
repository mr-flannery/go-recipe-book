package main

import (
	"log/slog"
	"net/http"
	"os"
	"time"

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
	config := config.GetConfig()

	slog.Info("Running migrations...")
	// Run database migrations
	err := db.RunMigrations()
	if err != nil {
		slog.Error("Failed to run migrations", "error", err)
		panic(err)
	}

	// Get database connection
	database, err := db.GetConnection()
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		panic(err)
	}
	defer database.Close()

	// Create seed admin account
	slog.Info("Creating seed admin account...")
	err = auth.CreateSeedAdmin(database, config.DB.Admin.Username, config.DB.Admin.Email, config.DB.Admin.Password)
	if err != nil {
		slog.Error("Failed to create seed admin", "error", err)
		panic(err)
	}

	// Start session cleanup routine
	go func() {
		for {
			time.Sleep(1 * time.Hour) // Run cleanup every hour
			if err := auth.CleanupExpiredSessions(database); err != nil {
				slog.Error("Failed to cleanup expired sessions", "error", err)
			}
		}
	}()

	// Create authentication middleware
	requireAuth := auth.RequireAuth(database)
	requireAPIKey := auth.RequireAPIKey()

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Home page
	mux.HandleFunc("/", handlers.HomeHandler)

	// Auth routes
	mux.HandleFunc("GET /login", handlers.GetLoginHandler)
	mux.HandleFunc("POST /login", handlers.PostLoginHandler)
	mux.HandleFunc("GET /logout", handlers.LogoutHandler)
	mux.HandleFunc("GET /register", handlers.GetRegisterHandler)
	mux.HandleFunc("POST /register", handlers.PostRegisterHandler)

	// Admin routes - require authentication and admin privileges
	requireAdminAuth := func(next http.Handler) http.Handler {
		return requireAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get database connection
			database, err := db.GetConnection()
			if err != nil {
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			defer database.Close()

			// Get current user
			user, err := auth.GetUserBySession(database, r)
			if err != nil || !user.IsAdmin {
				http.Error(w, "Access denied - admin privileges required", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		}))
	}

	mux.Handle("GET /admin/registrations", requireAdminAuth(http.HandlerFunc(handlers.GetPendingRegistrationsHandler)))
	mux.Handle("POST /admin/registrations/{id}/approve", requireAdminAuth(http.HandlerFunc(handlers.ApproveRegistrationHandler)))
	mux.Handle("POST /admin/registrations/{id}/deny", requireAdminAuth(http.HandlerFunc(handlers.DenyRegistrationHandler)))

	// Recipe routes with parameters
	mux.Handle("GET /recipes/create", requireAuth(http.HandlerFunc(handlers.GetCreateRecipeHandler)))
	mux.Handle("POST /recipes/create", requireAuth(http.HandlerFunc(handlers.PostCreateRecipeHandler)))
	mux.Handle("GET /recipes/update", requireAuth(http.HandlerFunc(handlers.GetUpdateRecipeHandler)))
	mux.Handle("POST /recipes/update", requireAuth(http.HandlerFunc(handlers.PostUpdateRecipeHandler)))
	mux.Handle("DELETE /recipes/{id}/delete", requireAuth(http.HandlerFunc(handlers.DeleteRecipeHandler)))
	mux.HandleFunc("GET /recipes", handlers.ListRecipesHandler)
	mux.HandleFunc("POST /recipes/filter", handlers.FilterRecipesHTMXHandler)

	// Recipe view route with ID parameter - /recipes/{id}
	mux.HandleFunc("GET /recipes/{id}", handlers.ViewRecipeHandler)

	// Recipe comments route with ID parameter - /recipes/{id}/comments/htmx
	mux.Handle("POST /recipes/{id}/comments/htmx", requireAuth(http.HandlerFunc(handlers.CommentHTMXHandler)))

	// API routes - protected by API key authentication
	mux.HandleFunc("GET /api/health", handlers.APIHealthHandler)
	mux.Handle("POST /api/recipe/upload", requireAPIKey(http.HandlerFunc(handlers.APICreateRecipeHandler)))

	slog.Info("Ready to serve!")

	slog.Error("Server failed to start", "error", http.ListenAndServe(addr, mux))
}
