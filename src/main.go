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
	"github.com/mr-flannery/go-recipe-book/src/store/postgres"
	"github.com/mr-flannery/go-recipe-book/src/templates"
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
	err := db.RunMigrations()
	if err != nil {
		slog.Error("Failed to run migrations", "error", err)
		panic(err)
	}

	slog.Info("Initializing database connection pool...")
	database, err := db.InitPool()
	if err != nil {
		slog.Error("Failed to initialize database pool", "error", err)
		panic(err)
	}
	defer db.ClosePool()

	authStore := postgres.NewAuthStore(database)

	slog.Info("Creating seed admin account...")
	err = auth.CreateSeedAdmin(authStore, config.DB.Admin.Username, config.DB.Admin.Email, config.DB.Admin.Password)
	if err != nil {
		slog.Error("Failed to create seed admin", "error", err)
		panic(err)
	}

	go func() {
		for {
			time.Sleep(1 * time.Hour)
			if err := auth.CleanupExpiredSessions(authStore); err != nil {
				slog.Error("Failed to cleanup expired sessions", "error", err)
			}
		}
	}()

	recipeStore := postgres.NewRecipeStore(database)
	tagStore := postgres.NewTagStore(database)
	userTagStore := postgres.NewUserTagStore(database)
	commentStore := postgres.NewCommentStore(database)
	userStore := postgres.NewUserStore(database)
	renderer := templates.NewRenderer(templates.Templates)

	h := handlers.NewHandler(database, recipeStore, tagStore, userTagStore, commentStore, userStore, authStore, renderer)

	userContext := auth.UserContextMiddleware(authStore)
	requireAuth := auth.RequireAuth()
	requireAPIKey := auth.RequireAPIKey()

	mux := http.NewServeMux()

	staticPath := filepath.Join(utils.GetCallerDir(0), "static")
	staticHandler := http.StripPrefix("/static/", http.FileServer(http.Dir(staticPath)))
	if config.Environment.Mode == "development" {
		mux.Handle("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			staticHandler.ServeHTTP(w, r)
		}))
	} else {
		mux.Handle("/static/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Cache-Control", "public, max-age=86400")
			staticHandler.ServeHTTP(w, r)
		}))
	}

	mux.HandleFunc("GET /robots.txt", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join(staticPath, "robots.txt"))
	})

	mux.Handle("/", userContext(http.HandlerFunc(h.HomeHandler)))

	mux.Handle("/imprint", userContext(http.HandlerFunc(h.ImprintHandler)))

	mux.Handle("GET /login", userContext(http.HandlerFunc(h.GetLoginHandler)))
	mux.Handle("POST /login", userContext(http.HandlerFunc(h.PostLoginHandler)))
	mux.HandleFunc("GET /logout", h.LogoutHandler)
	mux.Handle("GET /register", userContext(http.HandlerFunc(h.GetRegisterHandler)))
	mux.Handle("POST /register", userContext(http.HandlerFunc(h.PostRegisterHandler)))

	requireAdminAuth := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
					http.HandlerFunc(h.GetPendingRegistrationsHandler)))))
	mux.Handle("POST /admin/registrations/{id}/approve",
		userContext(
			requireAuth(
				requireAdminAuth(
					http.HandlerFunc(h.ApproveRegistrationHandler)))))
	mux.Handle("POST /admin/registrations/{id}/deny",
		userContext(
			requireAuth(
				requireAdminAuth(
					http.HandlerFunc(h.DenyRegistrationHandler)))))
	mux.Handle("GET /admin/users",
		userContext(
			requireAuth(
				requireAdminAuth(
					http.HandlerFunc(h.GetUsersHandler)))))
	mux.Handle("DELETE /admin/users/{id}",
		userContext(
			requireAuth(
				requireAdminAuth(
					http.HandlerFunc(h.DeleteUserHandler)))))

	mux.Handle("GET /recipes/create",
		userContext(
			requireAuth(
				http.HandlerFunc(h.GetCreateRecipeHandler))))
	mux.Handle("POST /recipes/create",
		userContext(
			requireAuth(
				http.HandlerFunc(h.PostCreateRecipeHandler))))
	mux.Handle("GET /recipes/update",
		userContext(
			requireAuth(
				http.HandlerFunc(h.GetUpdateRecipeHandler))))
	mux.Handle("POST /recipes/update",
		userContext(
			requireAuth(
				http.HandlerFunc(h.PostUpdateRecipeHandler))))
	mux.Handle("DELETE /recipes/{id}/delete",
		userContext(
			requireAuth(
				http.HandlerFunc(h.DeleteRecipeHandler))))
	mux.Handle("GET /recipes",
		userContext(
			http.HandlerFunc(h.ListRecipesHandler)))
	mux.Handle("GET /recipes/random",
		userContext(
			http.HandlerFunc(h.RandomRecipeHandler)))
	mux.Handle("POST /recipes/filter",
		userContext(
			http.HandlerFunc(h.FilterRecipesHTMXHandler)))

	mux.Handle("GET /recipes/{id}",
		userContext(
			http.HandlerFunc(h.ViewRecipeHandler)))

	mux.Handle("POST /recipes/{id}/comments/htmx",
		userContext(
			requireAuth(
				http.HandlerFunc(h.CommentHTMXHandler))))

	mux.Handle("PUT /comments/{id}",
		userContext(
			requireAuth(
				http.HandlerFunc(h.UpdateCommentHandler))))

	mux.Handle("DELETE /comments/{id}",
		userContext(
			requireAuth(
				http.HandlerFunc(h.DeleteCommentHandler))))

	mux.HandleFunc("GET /api/tags/search", h.SearchTagsHandler)
	mux.Handle("GET /api/tags/user/search",
		userContext(
			requireAuth(
				http.HandlerFunc(h.SearchUserTagsHandler))))
	mux.Handle("POST /recipes/{id}/tags",
		userContext(
			requireAuth(
				http.HandlerFunc(h.AddTagToRecipeHandler))))
	mux.Handle("DELETE /recipes/{id}/tags/{tagId}",
		userContext(
			requireAuth(
				http.HandlerFunc(h.RemoveTagFromRecipeHandler))))
	mux.Handle("POST /recipes/{id}/user-tags",
		userContext(
			requireAuth(
				http.HandlerFunc(h.AddUserTagToRecipeHandler))))
	mux.Handle("DELETE /user-tags/{tagId}",
		userContext(
			requireAuth(
				http.HandlerFunc(h.RemoveUserTagHandler))))

	mux.HandleFunc("GET /api/health", handlers.APIHealthHandler)
	mux.Handle("POST /api/recipe/upload",
		requireAPIKey(
			http.HandlerFunc(h.APICreateRecipeHandler)))

	slog.Info("Ready to serve!")

	slog.Error("Server failed to start", "error", http.ListenAndServe(addr, mux))
}
