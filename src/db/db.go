package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"path/filepath"
	"runtime"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/lib/pq"
	"github.com/mr-flannery/go-recipe-book/src/config"
)

// getPackageDir returns the directory containing this source file
func getPackageDir() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("failed to get current file path")
	}
	return filepath.Dir(filename)
}

// implement connection pool at some point in time?
func GetConnection() (*sql.DB, error) {
	config := config.GetConfig()

	connectionString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.DB.Host,
		config.DB.Port,
		config.DB.User,
		config.DB.Password,
		config.DB.Name,
		config.DB.SSLMode,
	)
	return sql.Open("postgres", connectionString)
}

func RunMigrations() error {
	db, err := GetConnection()
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
		return err
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		slog.Error("Failed to create migration driver", "error", err)
	}

	migrationsPath := filepath.Join(getPackageDir(), "migrations")
	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"postgres",
		driver,
	)
	if err != nil {
		slog.Error("Failed to initialize migrations", "error", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		slog.Error("Failed to apply migrations", "error", err)
	}

	slog.Info("Migrations applied successfully")
	return nil
}
