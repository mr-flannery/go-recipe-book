package postgres

import (
	"database/sql"
	"fmt"
	"strings"
)

type IngredientStore struct {
	db *sql.DB
}

func NewIngredientStore(db *sql.DB) *IngredientStore {
	return &IngredientStore{db: db}
}

func (s *IngredientStore) Search(query string, limit int) ([]string, error) {
	searchPattern := "%" + strings.ToLower(query) + "%"
	rows, err := s.db.Query(
		"SELECT name FROM ingredients WHERE LOWER(name) LIKE $1 ORDER BY name LIMIT $2",
		searchPattern, limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search ingredients: %v", err)
	}
	defer rows.Close()

	var results []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("failed to scan ingredient: %v", err)
		}
		results = append(results, name)
	}

	return results, rows.Err()
}

func (s *IngredientStore) GetOrCreate(name string) (int, error) {
	normalizedName := strings.ToLower(strings.TrimSpace(name))

	var id int
	err := s.db.QueryRow(
		"INSERT INTO ingredients (name) VALUES ($1) ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name RETURNING id",
		normalizedName,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to get or create ingredient: %v", err)
	}

	return id, nil
}
