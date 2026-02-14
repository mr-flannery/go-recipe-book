package postgres

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/mr-flannery/go-recipe-book/src/models"
)

type UserTagStore struct {
	db *sql.DB
}

func NewUserTagStore(db *sql.DB) *UserTagStore {
	return &UserTagStore{db: db}
}

func (s *UserTagStore) GetOrCreate(userID int, recipeID int, name string) (models.UserTag, error) {
	normalizedName := strings.ToLower(strings.TrimSpace(name))
	if normalizedName == "" {
		return models.UserTag{}, fmt.Errorf("tag name cannot be empty")
	}

	var userTag models.UserTag
	err := s.db.QueryRow(
		"SELECT id, user_id, recipe_id, name FROM user_tags WHERE user_id = $1 AND recipe_id = $2 AND name = $3",
		userID, recipeID, normalizedName,
	).Scan(&userTag.ID, &userTag.UserID, &userTag.RecipeID, &userTag.Name)
	if err == nil {
		return userTag, nil
	}

	err = s.db.QueryRow(
		"INSERT INTO user_tags (user_id, recipe_id, name) VALUES ($1, $2, $3) RETURNING id, user_id, recipe_id, name",
		userID, recipeID, normalizedName,
	).Scan(&userTag.ID, &userTag.UserID, &userTag.RecipeID, &userTag.Name)
	if err != nil {
		return models.UserTag{}, fmt.Errorf("failed to create user tag: %v", err)
	}

	return userTag, nil
}

func (s *UserTagStore) Search(userID int, query string) ([]string, error) {
	searchPattern := "%" + strings.ToLower(strings.TrimSpace(query)) + "%"
	rows, err := s.db.Query(
		"SELECT DISTINCT name FROM user_tags WHERE user_id = $1 AND name LIKE $2 ORDER BY name LIMIT 20",
		userID, searchPattern,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search user tags: %v", err)
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("failed to scan user tag: %v", err)
		}
		tags = append(tags, name)
	}

	return tags, nil
}

func (s *UserTagStore) GetByRecipeID(userID int, recipeID int) ([]models.UserTag, error) {
	rows, err := s.db.Query(
		"SELECT id, user_id, recipe_id, name FROM user_tags WHERE user_id = $1 AND recipe_id = $2 ORDER BY name",
		userID, recipeID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user tags: %v", err)
	}
	defer rows.Close()

	var tags []models.UserTag
	for rows.Next() {
		var tag models.UserTag
		if err := rows.Scan(&tag.ID, &tag.UserID, &tag.RecipeID, &tag.Name); err != nil {
			return nil, fmt.Errorf("failed to scan user tag: %v", err)
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

func (s *UserTagStore) GetForRecipes(userID int, recipeIDs []int) (map[int][]models.UserTag, error) {
	result := make(map[int][]models.UserTag)
	if len(recipeIDs) == 0 {
		return result, nil
	}

	placeholders := make([]string, len(recipeIDs))
	args := make([]interface{}, len(recipeIDs)+1)
	args[0] = userID
	for i, id := range recipeIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = id
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, recipe_id, name 
		FROM user_tags 
		WHERE user_id = $1 AND recipe_id IN (%s) 
		ORDER BY recipe_id, name`, strings.Join(placeholders, ","))

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get user tags: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var tag models.UserTag
		if err := rows.Scan(&tag.ID, &tag.UserID, &tag.RecipeID, &tag.Name); err != nil {
			return nil, fmt.Errorf("failed to scan user tag: %v", err)
		}
		result[tag.RecipeID] = append(result[tag.RecipeID], tag)
	}

	return result, nil
}

func (s *UserTagStore) Remove(userID int, tagID int) error {
	_, err := s.db.Exec("DELETE FROM user_tags WHERE id = $1 AND user_id = $2", tagID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove user tag: %v", err)
	}

	return nil
}
