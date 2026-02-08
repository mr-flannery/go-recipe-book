package main

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/mr-flannery/go-recipe-book/src/auth"
	"github.com/mr-flannery/go-recipe-book/src/config"
	"github.com/mr-flannery/go-recipe-book/src/db"
	"github.com/mr-flannery/go-recipe-book/src/handlers"
	"github.com/mr-flannery/go-recipe-book/src/utils"

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
	userContext := auth.UserContextMiddleware()
	requireAuth := auth.RequireAuth()
	requireAPIKey := auth.RequireAPIKey()

	// Create a new ServeMux
	mux := http.NewServeMux()

	// Static files (CSS, JS, images, etc.)
	staticPath := filepath.Join(utils.GetCallerDir(0), "static")
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticPath))))

	// Serve robots.txt from root
	mux.HandleFunc("GET /robots.txt", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(staticPath, "robots.txt"))
	})

	// Home page
	mux.Handle("/", userContext(http.HandlerFunc(handlers.HomeHandler)))

	// Imprint page
	mux.Handle("/imprint", userContext(http.HandlerFunc(handlers.ImprintHandler)))

	// Auth routes
	mux.HandleFunc("GET /login", handlers.GetLoginHandler)
	mux.HandleFunc("POST /login", handlers.PostLoginHandler)
	mux.HandleFunc("GET /logout", handlers.LogoutHandler)
	mux.HandleFunc("GET /register", handlers.GetRegisterHandler)
	mux.HandleFunc("POST /register", handlers.PostRegisterHandler)

	// Admin routes - require authentication and admin privileges
	requireAdminAuth := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get current user from context
			if !auth.IsUserAdmin(r.Context()) {
				http.Error(w, "Access denied - admin privileges required", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}

	mux.Handle("GET /admin/registrations",
		userContext(
			requireAuth(
				requireAdminAuth(
					http.HandlerFunc(handlers.GetPendingRegistrationsHandler)))))
	mux.Handle("POST /admin/registrations/{id}/approve",
		userContext(
			requireAuth(
				requireAdminAuth(
					http.HandlerFunc(handlers.ApproveRegistrationHandler)))))
	mux.Handle("POST /admin/registrations/{id}/deny",
		userContext(
			requireAuth(
				requireAdminAuth(
					http.HandlerFunc(handlers.DenyRegistrationHandler)))))

	// Recipe routes with parameters
	mux.Handle("GET /recipes/create",
		userContext(
			requireAuth(
				http.HandlerFunc(handlers.GetCreateRecipeHandler))))
	mux.Handle("POST /recipes/create",
		userContext(
			requireAuth(
				http.HandlerFunc(handlers.PostCreateRecipeHandler))))
	mux.Handle("GET /recipes/update",
		userContext(
			requireAuth(
				http.HandlerFunc(handlers.GetUpdateRecipeHandler))))
	mux.Handle("POST /recipes/update",
		userContext(
			requireAuth(
				http.HandlerFunc(handlers.PostUpdateRecipeHandler))))
	mux.Handle("DELETE /recipes/{id}/delete",
		userContext(
			requireAuth(
				http.HandlerFunc(handlers.DeleteRecipeHandler))))
	mux.Handle("GET /recipes",
		userContext(
			http.HandlerFunc(handlers.ListRecipesHandler)))
	mux.Handle("GET /recipes/random",
		userContext(
			http.HandlerFunc(handlers.RandomRecipeHandler)))
	mux.Handle("POST /recipes/filter",
		userContext(
			http.HandlerFunc(handlers.FilterRecipesHTMXHandler)))

	// Recipe view route with ID parameter - /recipes/{id}
	mux.Handle("GET /recipes/{id}",
		userContext(
			http.HandlerFunc(handlers.ViewRecipeHandler)))

	// Recipe comments route with ID parameter - /recipes/{id}/comments/htmx
	mux.Handle("POST /recipes/{id}/comments/htmx",
		userContext(
			requireAuth(
				http.HandlerFunc(handlers.CommentHTMXHandler))))

	// Tag routes
	mux.HandleFunc("GET /api/tags/search", handlers.SearchTagsHandler)
	mux.Handle("GET /api/tags/user/search",
		userContext(
			requireAuth(
				http.HandlerFunc(handlers.SearchUserTagsHandler))))
	mux.Handle("POST /recipes/{id}/tags",
		userContext(
			requireAuth(
				http.HandlerFunc(handlers.AddTagToRecipeHandler))))
	mux.Handle("DELETE /recipes/{id}/tags/{tagId}",
		userContext(
			requireAuth(
				http.HandlerFunc(handlers.RemoveTagFromRecipeHandler))))
	mux.Handle("POST /recipes/{id}/user-tags",
		userContext(
			requireAuth(
				http.HandlerFunc(handlers.AddUserTagToRecipeHandler))))
	mux.Handle("DELETE /user-tags/{tagId}",
		userContext(
			requireAuth(
				http.HandlerFunc(handlers.RemoveUserTagHandler))))

	// API routes - protected by API key authentication
	mux.HandleFunc("GET /api/health", handlers.APIHealthHandler)
	mux.Handle("POST /api/recipe/upload",
		requireAPIKey(
			http.HandlerFunc(handlers.APICreateRecipeHandler)))

	slog.Info("Ready to serve!")

	slog.Error("Server failed to start", "error", http.ListenAndServe(addr, mux))
}
