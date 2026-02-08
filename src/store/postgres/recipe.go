package postgres

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/mr-flannery/go-recipe-book/src/models"
)

type RecipeStore struct {
	db *sql.DB
}

func NewRecipeStore(db *sql.DB) *RecipeStore {
	return &RecipeStore{db: db}
}

func (s *RecipeStore) Save(recipe models.Recipe) (int, error) {
	query := `INSERT INTO recipes (title, ingredients_md, instructions_md, prep_time, cook_time, calories, author_id, image, parent_id, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id`

	var id int
	err := s.db.QueryRow(query, recipe.Title, recipe.IngredientsMD, recipe.InstructionsMD, recipe.PrepTime, recipe.CookTime, recipe.Calories, recipe.AuthorID, recipe.Image, recipe.ParentID, time.Now(), time.Now()).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (s *RecipeStore) GetByID(id string) (models.Recipe, error) {
	var recipe models.Recipe

	err := s.db.
		QueryRow("SELECT id, title, ingredients_md, instructions_md, prep_time, cook_time, calories, author_id, image, parent_id, created_at, updated_at FROM recipes WHERE id = $1", id).
		Scan(&recipe.ID, &recipe.Title, &recipe.IngredientsMD, &recipe.InstructionsMD, &recipe.PrepTime, &recipe.CookTime, &recipe.Calories, &recipe.AuthorID, &recipe.Image, &recipe.ParentID, &recipe.CreatedAt, &recipe.UpdatedAt)

	if err != nil {
		return models.Recipe{}, err
	}

	return recipe, nil
}

func (s *RecipeStore) Update(recipe models.Recipe) error {
	_, err := s.db.Exec("UPDATE recipes SET title = $1, ingredients_md = $2, instructions_md = $3, prep_time = $4, cook_time = $5, calories = $6, image = $7, updated_at = $8 WHERE id = $9",
		recipe.Title, recipe.IngredientsMD, recipe.InstructionsMD, recipe.PrepTime, recipe.CookTime, recipe.Calories, recipe.Image, time.Now(), recipe.ID)

	return err
}

func (s *RecipeStore) Delete(id string) error {
	_, err := s.db.Exec("DELETE FROM recipes WHERE id = $1", id)
	return err
}

func (s *RecipeStore) GetAll() ([]models.Recipe, error) {
	rows, err := s.db.Query("SELECT id, title, ingredients_md, instructions_md, prep_time, cook_time, calories, author_id, image, parent_id, created_at, updated_at FROM recipes")

	if err != nil {
		return nil, fmt.Errorf("failed to fetch recipes: %v", err)
	}
	defer rows.Close()

	var recipes []models.Recipe
	for rows.Next() {
		var recipe models.Recipe
		if err := rows.Scan(&recipe.ID, &recipe.Title, &recipe.IngredientsMD, &recipe.InstructionsMD, &recipe.PrepTime, &recipe.CookTime, &recipe.Calories, &recipe.AuthorID, &recipe.Image, &recipe.ParentID, &recipe.CreatedAt, &recipe.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan recipe: %v", err)
		}
		recipes = append(recipes, recipe)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over recipes: %v", err)
	}

	return recipes, nil
}

func (s *RecipeStore) GetFiltered(params models.FilterParams) ([]models.Recipe, error) {
	query := "SELECT DISTINCT r.id, r.title, r.ingredients_md, r.instructions_md, r.prep_time, r.cook_time, r.calories, r.author_id, r.image, r.parent_id, r.created_at, r.updated_at FROM recipes r"
	args := []interface{}{}
	argIndex := 1

	if len(params.Tags) > 0 || params.Search != "" {
		query += " LEFT JOIN recipe_tags rt ON r.id = rt.recipe_id LEFT JOIN tags t ON rt.tag_id = t.id"
	}

	if len(params.UserTags) > 0 || (params.Search != "" && params.UserID > 0) {
		query += " LEFT JOIN user_tags ut ON r.id = ut.recipe_id"
		if params.UserID > 0 {
			query += fmt.Sprintf(" AND ut.user_id = $%d", argIndex)
			args = append(args, params.UserID)
			argIndex++
		}
	}

	query += " WHERE 1=1"

	if params.Search != "" {
		searchPattern := "%" + strings.ToLower(params.Search) + "%"
		if params.UserID > 0 {
			query += fmt.Sprintf(" AND (LOWER(r.title) LIKE $%d OR LOWER(r.ingredients_md) LIKE $%d OR LOWER(r.instructions_md) LIKE $%d OR LOWER(t.name) LIKE $%d OR LOWER(ut.name) LIKE $%d)", argIndex, argIndex, argIndex, argIndex, argIndex)
		} else {
			query += fmt.Sprintf(" AND (LOWER(r.title) LIKE $%d OR LOWER(r.ingredients_md) LIKE $%d OR LOWER(r.instructions_md) LIKE $%d OR LOWER(t.name) LIKE $%d)", argIndex, argIndex, argIndex, argIndex)
		}
		args = append(args, searchPattern)
		argIndex++
	}

	for _, tagName := range params.Tags {
		normalizedTag := strings.ToLower(strings.TrimSpace(tagName))
		if normalizedTag == "" {
			continue
		}
		query += fmt.Sprintf(" AND r.id IN (SELECT rt2.recipe_id FROM recipe_tags rt2 INNER JOIN tags t2 ON rt2.tag_id = t2.id WHERE t2.name = $%d)", argIndex)
		args = append(args, normalizedTag)
		argIndex++
	}

	if params.UserID > 0 && len(params.UserTags) > 0 {
		for _, tagName := range params.UserTags {
			normalizedTag := strings.ToLower(strings.TrimSpace(tagName))
			if normalizedTag == "" {
				continue
			}
			query += fmt.Sprintf(" AND r.id IN (SELECT ut2.recipe_id FROM user_tags ut2 WHERE ut2.user_id = $%d AND ut2.name = $%d)", argIndex, argIndex+1)
			args = append(args, params.UserID, normalizedTag)
			argIndex += 2
		}
	}

	if params.CaloriesValue > 0 && params.CaloriesOp != "" {
		switch params.CaloriesOp {
		case "eq":
			query += fmt.Sprintf(" AND r.calories = $%d", argIndex)
		case "gt":
			query += fmt.Sprintf(" AND r.calories > $%d", argIndex)
		case "gte":
			query += fmt.Sprintf(" AND r.calories >= $%d", argIndex)
		case "lt":
			query += fmt.Sprintf(" AND r.calories < $%d", argIndex)
		case "lte":
			query += fmt.Sprintf(" AND r.calories <= $%d", argIndex)
		}
		args = append(args, params.CaloriesValue)
		argIndex++
	}

	if params.PrepTimeValue > 0 && params.PrepTimeOp != "" {
		switch params.PrepTimeOp {
		case "eq":
			query += fmt.Sprintf(" AND r.prep_time = $%d", argIndex)
		case "gt":
			query += fmt.Sprintf(" AND r.prep_time > $%d", argIndex)
		case "gte":
			query += fmt.Sprintf(" AND r.prep_time >= $%d", argIndex)
		case "lt":
			query += fmt.Sprintf(" AND r.prep_time < $%d", argIndex)
		case "lte":
			query += fmt.Sprintf(" AND r.prep_time <= $%d", argIndex)
		}
		args = append(args, params.PrepTimeValue)
		argIndex++
	}

	if params.CookTimeValue > 0 && params.CookTimeOp != "" {
		switch params.CookTimeOp {
		case "eq":
			query += fmt.Sprintf(" AND r.cook_time = $%d", argIndex)
		case "gt":
			query += fmt.Sprintf(" AND r.cook_time > $%d", argIndex)
		case "gte":
			query += fmt.Sprintf(" AND r.cook_time >= $%d", argIndex)
		case "lt":
			query += fmt.Sprintf(" AND r.cook_time < $%d", argIndex)
		case "lte":
			query += fmt.Sprintf(" AND r.cook_time <= $%d", argIndex)
		}
		args = append(args, params.CookTimeValue)
		argIndex++
	}

	query += " ORDER BY r.created_at DESC"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch filtered recipes: %v", err)
	}
	defer rows.Close()

	var recipes []models.Recipe
	for rows.Next() {
		var recipe models.Recipe
		if err := rows.Scan(&recipe.ID, &recipe.Title, &recipe.IngredientsMD, &recipe.InstructionsMD, &recipe.PrepTime, &recipe.CookTime, &recipe.Calories, &recipe.AuthorID, &recipe.Image, &recipe.ParentID, &recipe.CreatedAt, &recipe.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan recipe: %v", err)
		}
		recipes = append(recipes, recipe)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over filtered recipes: %v", err)
	}

	return recipes, nil
}

func (s *RecipeStore) GetRandomID() (int, error) {
	var id int

	err := s.db.QueryRow("SELECT id FROM recipes ORDER BY RANDOM() LIMIT 1").Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to get random recipe: %v", err)
	}

	return id, nil
}
