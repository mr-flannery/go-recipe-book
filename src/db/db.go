package db

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/lib/pq"
)

func Connect() (*sql.DB, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, dbname, sslmode)
	return sql.Open("postgres", dsn)
}

func RunMigrations(dataSourceName string) {
	db, err := sql.Open("postgres", dataSourceName)
	if err != nil {
		slog.Error("Failed to connect to database", "error", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		slog.Error("Failed to create migration driver", "error", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://db/migrations",
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
}
