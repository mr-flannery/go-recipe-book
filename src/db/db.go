package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/lib/pq"
	"github.com/mr-flannery/go-recipe-book/src/config"
	"github.com/mr-flannery/go-recipe-book/src/utils"
)

var pool *sql.DB

func InitPool() (*sql.DB, error) {
	if pool != nil {
		return pool, nil
	}

	cfg := config.GetConfig()

	connectionString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Name,
		cfg.DB.SSLMode,
	)

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)

	pool = db
	return pool, nil
}

func GetPool() *sql.DB {
	return pool
}

func ClosePool() error {
	if pool != nil {
		return pool.Close()
	}
	return nil
}

func GetConnection() (*sql.DB, error) {
	cfg := config.GetConfig()

	connectionString := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Name,
		cfg.DB.SSLMode,
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

	migrationsPath := filepath.Join(utils.GetCallerDir(0), "migrations")
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
