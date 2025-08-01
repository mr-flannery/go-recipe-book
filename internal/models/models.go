package models

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// User represents a user in the system
type User struct {
	ID           int
	Username     string
	PasswordHash string
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

// TODO: this is probably garbage...
var db *sql.DB

// InitializeDB initializes the database connection
func InitializeDB(dataSourceName string) error {
	var err error

	db, err = sql.Open("postgres", dataSourceName)
	if err != nil {
		return err
	}
	return db.Ping()
}

// SaveRecipe saves a recipe to the database
func SaveRecipe(recipe Recipe) error {
	query := `INSERT INTO recipes (title, ingredients_md, instructions_md, prep_time, cook_time, calories, author_id, image, parent_id, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`
	_, err := db.Exec(query, recipe.Title, recipe.IngredientsMD, recipe.InstructionsMD, recipe.PrepTime, recipe.CookTime, recipe.Calories, recipe.AuthorID, recipe.Image, recipe.ParentID, time.Now(), time.Now())
	return err
}

// GetRecipeByID retrieves a recipe by its ID
func GetRecipeByID(id string) (Recipe, error) {
	var recipe Recipe
	err := db.QueryRow("SELECT id, title, ingredients_md, instructions_md FROM recipes WHERE id = $1", id).
		Scan(&recipe.ID, &recipe.Title, &recipe.IngredientsMD, &recipe.InstructionsMD)
	if err != nil {
		return Recipe{}, err
	}
	return recipe, nil
}

// UpdateRecipe updates an existing recipe in the database
func UpdateRecipe(recipe Recipe) error {
	_, err := db.Exec("UPDATE recipes SET title = $1, ingredients_md = $2, instructions_md = $3 WHERE id = $4",
		recipe.Title, recipe.IngredientsMD, recipe.InstructionsMD, recipe.ID)
	return err
}

// DeleteRecipe deletes a recipe from the database
func DeleteRecipe(id string) error {
	_, err := db.Exec("DELETE FROM recipes WHERE id = $1", id)
	return err
}

// GetAllRecipes fetches all recipes from the database
func GetAllRecipes() ([]Recipe, error) {
	db, err := sql.Open("postgres", "host=localhost port=5432 user=local-recipe-user password=local-recipe-password dbname=recipe-book sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT id, title, ingredients_md, instructions_md, prep_time, cook_time, calories, author_id, created_at, updated_at FROM recipes")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch recipes: %v", err)
	}
	defer rows.Close()

	var recipes []Recipe
	for rows.Next() {
		var recipe Recipe
		if err := rows.Scan(&recipe.ID, &recipe.Title, &recipe.IngredientsMD, &recipe.InstructionsMD, &recipe.PrepTime, &recipe.CookTime, &recipe.Calories, &recipe.AuthorID, &recipe.CreatedAt, &recipe.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan recipe: %v", err)
		}
		recipes = append(recipes, recipe)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over recipes: %v", err)
	}

	return recipes, nil
}
