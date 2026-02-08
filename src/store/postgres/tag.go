package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/mr-flannery/go-recipe-book/src/models"
)

type TagStore struct {
	db *sql.DB
}

func NewTagStore(db *sql.DB) *TagStore {
	return &TagStore{db: db}
}

func (s *TagStore) GetOrCreate(name string) (models.Tag, error) {
	normalizedName := strings.ToLower(strings.TrimSpace(name))
	if normalizedName == "" {
		return models.Tag{}, fmt.Errorf("tag name cannot be empty")
	}

	var tag models.Tag
	err := s.db.QueryRow("SELECT id, name FROM tags WHERE name = $1", normalizedName).Scan(&tag.ID, &tag.Name)
	if err == nil {
		return tag, nil
	}
	if err != sql.ErrNoRows {
		return models.Tag{}, fmt.Errorf("failed to query tag: %v", err)
	}

	err = s.db.QueryRow("INSERT INTO tags (name) VALUES ($1) RETURNING id, name", normalizedName).Scan(&tag.ID, &tag.Name)
	if err != nil {
		return models.Tag{}, fmt.Errorf("failed to create tag: %v", err)
	}

	return tag, nil
}

func (s *TagStore) Search(query string) ([]models.Tag, error) {
	searchPattern := "%" + strings.ToLower(strings.TrimSpace(query)) + "%"
	rows, err := s.db.Query("SELECT id, name FROM tags WHERE name LIKE $1 ORDER BY name LIMIT 20", searchPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search tags: %v", err)
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		if err := rows.Scan(&tag.ID, &tag.Name); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %v", err)
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

func (s *TagStore) GetByRecipeID(recipeID int) ([]models.Tag, error) {
	rows, err := s.db.Query(`
		SELECT t.id, t.name 
		FROM tags t 
		INNER JOIN recipe_tags rt ON t.id = rt.tag_id 
		WHERE rt.recipe_id = $1 
		ORDER BY t.name`, recipeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %v", err)
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		if err := rows.Scan(&tag.ID, &tag.Name); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %v", err)
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

func (s *TagStore) GetForRecipes(recipeIDs []int) (map[int][]models.Tag, error) {
	result := make(map[int][]models.Tag)
	if len(recipeIDs) == 0 {
		return result, nil
	}

	placeholders := make([]string, len(recipeIDs))
	args := make([]interface{}, len(recipeIDs))
	for i, id := range recipeIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT rt.recipe_id, t.id, t.name 
		FROM tags t 
		INNER JOIN recipe_tags rt ON t.id = rt.tag_id 
		WHERE rt.recipe_id IN (%s) 
		ORDER BY rt.recipe_id, t.name`, strings.Join(placeholders, ","))

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var recipeID int
		var tag models.Tag
		if err := rows.Scan(&recipeID, &tag.ID, &tag.Name); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %v", err)
		}
		result[recipeID] = append(result[recipeID], tag)
	}

	return result, nil
}

func (s *TagStore) AddToRecipe(recipeID int, tagID int) error {
	_, err := s.db.Exec("INSERT INTO recipe_tags (recipe_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", recipeID, tagID)
	if err != nil {
		return fmt.Errorf("failed to add tag to recipe: %v", err)
	}

	return nil
}

func (s *TagStore) RemoveFromRecipe(recipeID int, tagID int) error {
	_, err := s.db.Exec("DELETE FROM recipe_tags WHERE recipe_id = $1 AND tag_id = $2", recipeID, tagID)
	if err != nil {
		return fmt.Errorf("failed to remove tag from recipe: %v", err)
	}

	return nil
}

func (s *TagStore) SetRecipeTags(recipeID int, tagNames []string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}

	_, err = tx.Exec("DELETE FROM recipe_tags WHERE recipe_id = $1", recipeID)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to clear existing tags: %v", err)
	}

	for _, name := range tagNames {
		normalizedName := strings.ToLower(strings.TrimSpace(name))
		if normalizedName == "" {
			continue
		}

		var tagID int
		err = tx.QueryRow("SELECT id FROM tags WHERE name = $1", normalizedName).Scan(&tagID)
		if err != nil {
			err = tx.QueryRow("INSERT INTO tags (name) VALUES ($1) RETURNING id", normalizedName).Scan(&tagID)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("failed to create tag: %v", err)
			}
		}

		_, err = tx.Exec("INSERT INTO recipe_tags (recipe_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", recipeID, tagID)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to add tag to recipe: %v", err)
		}
	}

	return tx.Commit()
}
