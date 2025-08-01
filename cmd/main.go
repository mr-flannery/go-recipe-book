package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/yourusername/agent-coding-recipe-book/auth"
	"github.com/yourusername/agent-coding-recipe-book/internal/handlers"
	"github.com/yourusername/agent-coding-recipe-book/internal/models"
)

// importHandlersLogin calls the Login handler from internal/handlers
func importHandlersLogin(w http.ResponseWriter, r *http.Request) {
	templates := template.Must(template.ParseGlob("templates/*.gohtml"))
	if r.Method == http.MethodGet {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		err := templates.ExecuteTemplate(w, "login.gohtml", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	if r.Method == http.MethodPost {
		r.ParseForm()
		username := r.FormValue("username")
		password := r.FormValue("password")

		ok := auth.Authenticate(username, password)

		if !ok {
			w.Write([]byte("<p>Invalid credentials</p>"))
			return
		}

		auth.SetSession(w, username)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
}

// importHandlersLogout logs the user out
func importHandlersLogout(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{Name: "session", Value: "", Path: "/", MaxAge: -1, Expires: time.Date(1970, time.January, 1, 0, 0, 0, 0, time.UTC)}
	http.SetCookie(w, cookie)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func runMigrations(dataSourceName string) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatalf("Failed to create migration driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://db/migrations",
		"postgres",
		driver,
	)
	if err != nil {
		log.Fatalf("Failed to initialize migrations: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Failed to apply migrations: %v", err)
	}

 	log.Println("Migrations applied successfully")
}

func main() {
	addr := ":8080"
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}
	log.Printf("Starting server on %s", addr)

	log.Print("Initializing database...")

	dataSourceName := "host=localhost port=5432 user=local-recipe-user password=local-recipe-password dbname=recipe-book sslmode=disable"

	err := models.InitializeDB(dataSourceName)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
		panic(err)
	}

	log.Print("Running migrations...")
	// Run database migrations
	runMigrations(dataSourceName)

	// Home page
	http.HandleFunc("/", importHandlersHome)

	// Auth routes
	http.HandleFunc("/login", importHandlersLogin)
	http.HandleFunc("/logout", importHandlersLogout)

	// Recipe routes
	http.HandleFunc("/recipes", handlers.ListRecipesHandler)
	http.HandleFunc("/recipes/create", handlers.CreateRecipeHandler)
	http.HandleFunc("/recipes/update", handlers.UpdateRecipeHandler)
	http.HandleFunc("/recipes/delete", handlers.DeleteRecipeHandler)

	log.Print("Ready to serve!")

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
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	username, isLoggedIn := auth.GetUser(r) // Check if the user is logged in
	data := struct {
		IsLoggedIn bool
		Username   string
	}{
		IsLoggedIn: isLoggedIn,
		Username:   username,
	}

	templates := template.Must(template.ParseGlob("templates/*.gohtml"))
	err := templates.ExecuteTemplate(w, "home.gohtml", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
