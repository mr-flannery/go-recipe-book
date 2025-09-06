package models

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"github.com/mr-flannery/go-recipe-book/src/db"
)

// User represents a user in the system
type User struct {
	ID           int
	Username     string
	Email        string
	PasswordHash string
	IsAdmin      bool
	IsActive     bool
	LastLogin    *time.Time
	CreatedAt    time.Time
}

// Recipe represents a recipe
type Recipe struct {
	ID             int
	Title          string
	IngredientsMD  string
	InstructionsMD string
	PrepTime       int
	CookTime       int
	Calories       int
	AuthorID       int
	Image          []byte
	ParentID       *int
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// ImageBase64 returns the base64 encoded image for display in templates
func (r Recipe) ImageBase64() string {
	if len(r.Image) == 0 {
		return ""
	}
	return base64.StdEncoding.EncodeToString(r.Image)
}

// Label represents a label
type Label struct {
	ID   int
	Name string
}

// RecipeLabel represents the many-to-many relationship
type RecipeLabel struct {
	RecipeID int
	LabelID  int
}

// Comment represents a comment on a recipe
type Comment struct {
	ID        int
	RecipeID  int
	AuthorID  int
	ContentMD string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ProposedChange represents a proposed change to a recipe
type ProposedChange struct {
	ID             int
	RecipeID       int
	ProposerID     int
	Title          string
	IngredientsMD  string
	InstructionsMD string
	PrepTime       int
	CookTime       int
	Calories       int
	Image          []byte
	CreatedAt      time.Time
	Status         string // pending, accepted, rejected
}

// SaveRecipe saves a recipe to the database
func SaveRecipe(recipe Recipe) error {
	// this is probably vulnerable to SQL injection...
	query := `INSERT INTO recipes (title, ingredients_md, instructions_md, prep_time, cook_time, calories, author_id, image, parent_id, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	dbConnection, err := db.GetConnection()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer dbConnection.Close()

	_, err = dbConnection.Exec(query, recipe.Title, recipe.IngredientsMD, recipe.InstructionsMD, recipe.PrepTime, recipe.CookTime, recipe.Calories, recipe.AuthorID, recipe.Image, recipe.ParentID, time.Now(), time.Now())

	return err
}

// GetRecipeByID retrieves a recipe by its ID with all fields
func GetRecipeByID(id string) (Recipe, error) {
	var recipe Recipe

	dbConnection, err := db.GetConnection()
	if err != nil {
		return Recipe{}, fmt.Errorf("failed to connect to database: %v", err)
	}
	defer dbConnection.Close()

	err = dbConnection.
		QueryRow("SELECT id, title, ingredients_md, instructions_md, prep_time, cook_time, calories, author_id, image, parent_id, created_at, updated_at FROM recipes WHERE id = $1", id).
		Scan(&recipe.ID, &recipe.Title, &recipe.IngredientsMD, &recipe.InstructionsMD, &recipe.PrepTime, &recipe.CookTime, &recipe.Calories, &recipe.AuthorID, &recipe.Image, &recipe.ParentID, &recipe.CreatedAt, &recipe.UpdatedAt)

	if err != nil {
		return Recipe{}, err
	}

	return recipe, nil
}

// UpdateRecipe updates an existing recipe in the database
func UpdateRecipe(recipe Recipe) error {
	dbConnection, err := db.GetConnection()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer dbConnection.Close()

	_, err = dbConnection.Exec("UPDATE recipes SET title = $1, ingredients_md = $2, instructions_md = $3, prep_time = $4, cook_time = $5, calories = $6, image = $7, updated_at = $8 WHERE id = $9",
		recipe.Title, recipe.IngredientsMD, recipe.InstructionsMD, recipe.PrepTime, recipe.CookTime, recipe.Calories, recipe.Image, time.Now(), recipe.ID)

	return err
}

// DeleteRecipe deletes a recipe from the database
func DeleteRecipe(id string) error {
	dbConnection, err := db.GetConnection()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer dbConnection.Close()

	_, err = dbConnection.Exec("DELETE FROM recipes WHERE id = $1", id)
	return err
}

// GetAllRecipes fetches all recipes from the database
func GetAllRecipes() ([]Recipe, error) {
	dbConnection, err := db.GetConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}
	defer dbConnection.Close()

	rows, err := dbConnection.Query("SELECT id, title, ingredients_md, instructions_md, prep_time, cook_time, calories, author_id, image, parent_id, created_at, updated_at FROM recipes")

	if err != nil {
		return nil, fmt.Errorf("failed to fetch recipes: %v", err)
	}
	defer rows.Close()

	var recipes []Recipe
	for rows.Next() {
		var recipe Recipe
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

// FilterParams represents the filtering parameters for recipes
type FilterParams struct {
	Search        string
	CaloriesOp    string
	CaloriesValue int
	PrepTimeOp    string
	PrepTimeValue int
	CookTimeOp    string
	CookTimeValue int
}

// GetFilteredRecipes fetches recipes based on filter parameters
func GetFilteredRecipes(params FilterParams) ([]Recipe, error) {
	dbConnection, err := db.GetConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}
	defer dbConnection.Close()

	// Build the query dynamically based on filters
	query := "SELECT id, title, ingredients_md, instructions_md, prep_time, cook_time, calories, author_id, image, parent_id, created_at, updated_at FROM recipes WHERE 1=1"
	args := []interface{}{}
	argIndex := 1

	// Add fuzzy search filter
	if params.Search != "" {
		query += fmt.Sprintf(" AND (LOWER(title) LIKE $%d OR LOWER(ingredients_md) LIKE $%d OR LOWER(instructions_md) LIKE $%d)", argIndex, argIndex, argIndex)
		searchPattern := "%" + strings.ToLower(params.Search) + "%"
		args = append(args, searchPattern)
		argIndex++
	}

	// Add calories filter
	if params.CaloriesValue > 0 && params.CaloriesOp != "" {
		switch params.CaloriesOp {
		case "eq":
			query += fmt.Sprintf(" AND calories = $%d", argIndex)
		case "gt":
			query += fmt.Sprintf(" AND calories > $%d", argIndex)
		case "gte":
			query += fmt.Sprintf(" AND calories >= $%d", argIndex)
		case "lt":
			query += fmt.Sprintf(" AND calories < $%d", argIndex)
		case "lte":
			query += fmt.Sprintf(" AND calories <= $%d", argIndex)
		}
		args = append(args, params.CaloriesValue)
		argIndex++
	}

	// Add prep time filter
	if params.PrepTimeValue > 0 && params.PrepTimeOp != "" {
		switch params.PrepTimeOp {
		case "eq":
			query += fmt.Sprintf(" AND prep_time = $%d", argIndex)
		case "gt":
			query += fmt.Sprintf(" AND prep_time > $%d", argIndex)
		case "gte":
			query += fmt.Sprintf(" AND prep_time >= $%d", argIndex)
		case "lt":
			query += fmt.Sprintf(" AND prep_time < $%d", argIndex)
		case "lte":
			query += fmt.Sprintf(" AND prep_time <= $%d", argIndex)
		}
		args = append(args, params.PrepTimeValue)
		argIndex++
	}

	// Add cook time filter
	if params.CookTimeValue > 0 && params.CookTimeOp != "" {
		switch params.CookTimeOp {
		case "eq":
			query += fmt.Sprintf(" AND cook_time = $%d", argIndex)
		case "gt":
			query += fmt.Sprintf(" AND cook_time > $%d", argIndex)
		case "gte":
			query += fmt.Sprintf(" AND cook_time >= $%d", argIndex)
		case "lt":
			query += fmt.Sprintf(" AND cook_time < $%d", argIndex)
		case "lte":
			query += fmt.Sprintf(" AND cook_time <= $%d", argIndex)
		}
		args = append(args, params.CookTimeValue)
		argIndex++
	}

	query += " ORDER BY created_at DESC"

	rows, err := dbConnection.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch filtered recipes: %v", err)
	}
	defer rows.Close()

	var recipes []Recipe
	for rows.Next() {
		var recipe Recipe
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

// GetCommentsByRecipeID fetches all comments for a specific recipe
func GetCommentsByRecipeID(recipeID string) ([]Comment, error) {
	dbConnection, err := db.GetConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}
	defer dbConnection.Close()

	rows, err := dbConnection.Query("SELECT id, recipe_id, author_id, content_md, created_at, updated_at FROM comments WHERE recipe_id = $1 ORDER BY created_at DESC", recipeID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch comments: %v", err)
	}
	defer rows.Close()

	var comments []Comment
	for rows.Next() {
		var comment Comment
		if err := rows.Scan(&comment.ID, &comment.RecipeID, &comment.AuthorID, &comment.ContentMD, &comment.CreatedAt, &comment.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan comment: %v", err)
		}
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over comments: %v", err)
	}

	return comments, nil
}

// SaveComment saves a comment to the database
func SaveComment(comment Comment) error {
	query := `INSERT INTO comments (recipe_id, author_id, content_md, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5)`

	dbConnection, err := db.GetConnection()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer dbConnection.Close()

	_, err = dbConnection.Exec(query, comment.RecipeID, comment.AuthorID, comment.ContentMD, time.Now(), time.Now())
	return err
}

// GetUsernameByID retrieves a username by user ID
func GetUsernameByID(userID int) (string, error) {
	var username string

	dbConnection, err := db.GetConnection()
	if err != nil {
		return "", fmt.Errorf("failed to connect to database: %v", err)
	}
	defer dbConnection.Close()

	err = dbConnection.QueryRow("SELECT username FROM users WHERE id = $1", userID).Scan(&username)
	if err != nil {
		return "", err
	}

	return username, nil
}

// GetLatestCommentByUserAndRecipe retrieves the latest comment by a specific user for a specific recipe
func GetLatestCommentByUserAndRecipe(userID int, recipeID int) (Comment, error) {
	var comment Comment

	dbConnection, err := db.GetConnection()
	if err != nil {
		return Comment{}, fmt.Errorf("failed to connect to database: %v", err)
	}
	defer dbConnection.Close()

	err = dbConnection.QueryRow(
		"SELECT id, recipe_id, author_id, content_md, created_at, updated_at FROM comments WHERE author_id = $1 AND recipe_id = $2 ORDER BY created_at DESC LIMIT 1",
		userID, recipeID,
	).Scan(&comment.ID, &comment.RecipeID, &comment.AuthorID, &comment.ContentMD, &comment.CreatedAt, &comment.UpdatedAt)

	if err != nil {
		return Comment{}, err
	}

	return comment, nil
}
