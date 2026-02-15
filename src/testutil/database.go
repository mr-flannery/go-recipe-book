package testutil

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	sharedDB   *TestDatabase
	sharedOnce sync.Once
	sharedMu   sync.Mutex
)

type TestDatabase struct {
	Container testcontainers.Container
	DB        *sql.DB
	Host      string
	Port      string
}

func SetupSharedTestDatabase() *TestDatabase {
	sharedOnce.Do(func() {
		ctx := context.Background()

		log.Println("Starting shared PostgreSQL container...")
		container, err := tcpostgres.Run(ctx,
			"postgres:16-alpine",
			tcpostgres.WithDatabase("testdb"),
			tcpostgres.WithUsername("testuser"),
			tcpostgres.WithPassword("testpass"),
			testcontainers.WithWaitStrategy(
				wait.ForLog("database system is ready to accept connections").
					WithOccurrence(2).
					WithStartupTimeout(30*time.Second),
			),
		)
		if err != nil {
			log.Fatalf("failed to start postgres container: %v", err)
		}

		host, err := container.Host(ctx)
		if err != nil {
			log.Fatalf("failed to get container host: %v", err)
		}

		mappedPort, err := container.MappedPort(ctx, "5432")
		if err != nil {
			log.Fatalf("failed to get mapped port: %v", err)
		}

		connStr := fmt.Sprintf("host=%s port=%s user=testuser password=testpass dbname=testdb sslmode=disable",
			host, mappedPort.Port())

		db, err := sql.Open("postgres", connStr)
		if err != nil {
			log.Fatalf("failed to connect to test database: %v", err)
		}

		if err := db.Ping(); err != nil {
			log.Fatalf("failed to ping test database: %v", err)
		}

		sharedDB = &TestDatabase{
			Container: container,
			DB:        db,
			Host:      host,
			Port:      mappedPort.Port(),
		}

		if err := sharedDB.runMigrations(); err != nil {
			log.Fatalf("failed to run migrations: %v", err)
		}

		log.Println("Shared PostgreSQL container ready")
	})

	return sharedDB
}

func TeardownSharedTestDatabase() {
	sharedMu.Lock()
	defer sharedMu.Unlock()

	if sharedDB != nil {
		if sharedDB.DB != nil {
			sharedDB.DB.Close()
		}
		if sharedDB.Container != nil {
			if err := sharedDB.Container.Terminate(context.Background()); err != nil {
				log.Printf("warning: failed to terminate container: %v", err)
			}
		}
		sharedDB = nil
	}
}

func GetTestDatabase(t *testing.T) *TestDatabase {
	t.Helper()
	SkipIfShort(t)

	td := SetupSharedTestDatabase()
	td.ResetAllTables(t)
	return td
}

func (td *TestDatabase) runMigrations() error {
	driver, err := postgres.WithInstance(td.DB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	_, currentFile, _, _ := runtime.Caller(0)
	migrationsPath := filepath.Join(filepath.Dir(currentFile), "..", "db", "migrations")

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrator: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

func (td *TestDatabase) TruncateTables(t *testing.T, tables ...string) {
	t.Helper()

	for _, table := range tables {
		_, err := td.DB.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			t.Fatalf("failed to truncate table %s: %v", table, err)
		}
	}
}

func (td *TestDatabase) ResetAllTables(t *testing.T) {
	t.Helper()

	tables := []string{
		"user_tags",
		"recipe_tags",
		"comments",
		"sessions",
		"registration_requests",
		"recipes",
		"tags",
		"users",
	}

	for _, table := range tables {
		_, err := td.DB.Exec(fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table))
		if err != nil {
			t.Logf("warning: failed to truncate table %s: %v", table, err)
		}
	}
}

func (td *TestDatabase) SeedUser(t *testing.T, username, email, passwordHash string, isAdmin bool) int {
	t.Helper()

	var userID int
	err := td.DB.QueryRow(`
		INSERT INTO users (username, email, password_hash, is_admin, is_active, created_at)
		VALUES ($1, $2, $3, $4, true, NOW())
		RETURNING id
	`, username, email, passwordHash, isAdmin).Scan(&userID)
	if err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}
	return userID
}

func (td *TestDatabase) SeedRecipe(t *testing.T, title, ingredientsMD, instructionsMD string, authorID int) int {
	t.Helper()

	var recipeID int
	err := td.DB.QueryRow(`
		INSERT INTO recipes (title, ingredients_md, instructions_md, author_id, prep_time, cook_time, calories, created_at, updated_at)
		VALUES ($1, $2, $3, $4, 10, 20, 300, NOW(), NOW())
		RETURNING id
	`, title, ingredientsMD, instructionsMD, authorID).Scan(&recipeID)
	if err != nil {
		t.Fatalf("failed to seed recipe: %v", err)
	}
	return recipeID
}

func (td *TestDatabase) SeedTag(t *testing.T, name string) int {
	t.Helper()

	var tagID int
	err := td.DB.QueryRow(`
		INSERT INTO tags (name) VALUES ($1)
		ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`, name).Scan(&tagID)
	if err != nil {
		t.Fatalf("failed to seed tag: %v", err)
	}
	return tagID
}

func (td *TestDatabase) SeedRecipeTag(t *testing.T, recipeID, tagID int) {
	t.Helper()

	_, err := td.DB.Exec(`
		INSERT INTO recipe_tags (recipe_id, tag_id) VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`, recipeID, tagID)
	if err != nil {
		t.Fatalf("failed to seed recipe tag: %v", err)
	}
}

func (td *TestDatabase) SeedUserTag(t *testing.T, userID, recipeID int, name string) int {
	t.Helper()

	var tagID int
	err := td.DB.QueryRow(`
		INSERT INTO user_tags (user_id, recipe_id, name) VALUES ($1, $2, $3)
		ON CONFLICT (user_id, recipe_id, name) DO UPDATE SET name = EXCLUDED.name
		RETURNING id
	`, userID, recipeID, name).Scan(&tagID)
	if err != nil {
		t.Fatalf("failed to seed user tag: %v", err)
	}
	return tagID
}

func (td *TestDatabase) SeedComment(t *testing.T, recipeID, authorID int, content string) int {
	t.Helper()

	var commentID int
	err := td.DB.QueryRow(`
		INSERT INTO comments (recipe_id, author_id, content_md, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id
	`, recipeID, authorID, content).Scan(&commentID)
	if err != nil {
		t.Fatalf("failed to seed comment: %v", err)
	}
	return commentID
}

func (td *TestDatabase) SeedSession(t *testing.T, sessionID string, userID int, expiresAt time.Time) {
	t.Helper()

	_, err := td.DB.Exec(`
		INSERT INTO sessions (id, user_id, created_at, expires_at, ip_address, user_agent)
		VALUES ($1, $2, NOW(), $3, '127.0.0.1', 'test-agent')
	`, sessionID, userID, expiresAt)
	if err != nil {
		t.Fatalf("failed to seed session: %v", err)
	}
}

func SkipIfShort(t *testing.T) {
	t.Helper()
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}
}
