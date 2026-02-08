package models

import (
	"database/sql"
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
	Tags           []Tag
}

// ImageBase64 returns the base64 encoded image for display in templates
func (r Recipe) ImageBase64() string {
	if len(r.Image) == 0 {
		return ""
	}
	return base64.StdEncoding.EncodeToString(r.Image)
}

// Tag represents a global author tag
type Tag struct {
	ID   int
	Name string
}

// UserTag represents a personal tag that a user has added to a recipe
type UserTag struct {
	ID       int
	UserID   int
	RecipeID int
	Name     string
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

// SaveRecipe saves a recipe to the database and returns the created recipe ID
func SaveRecipe(recipe Recipe) (int, error) {
	query := `INSERT INTO recipes (title, ingredients_md, instructions_md, prep_time, cook_time, calories, author_id, image, parent_id, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11) RETURNING id`

	dbConnection, err := db.GetConnection()
	if err != nil {
		return 0, fmt.Errorf("failed to connect to database: %v", err)
	}
	defer dbConnection.Close()

	var id int
	err = dbConnection.QueryRow(query, recipe.Title, recipe.IngredientsMD, recipe.InstructionsMD, recipe.PrepTime, recipe.CookTime, recipe.Calories, recipe.AuthorID, recipe.Image, recipe.ParentID, time.Now(), time.Now()).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
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
	Tags          []string // Author tags to filter by
	UserID        int      // User ID for user tag filtering
	UserTags      []string // User tags to filter by (requires UserID)
}

// GetFilteredRecipes fetches recipes based on filter parameters
func GetFilteredRecipes(params FilterParams) ([]Recipe, error) {
	dbConnection, err := db.GetConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}
	defer dbConnection.Close()

	query := "SELECT DISTINCT r.id, r.title, r.ingredients_md, r.instructions_md, r.prep_time, r.cook_time, r.calories, r.author_id, r.image, r.parent_id, r.created_at, r.updated_at FROM recipes r"
	args := []interface{}{}
	argIndex := 1

	// Join with tags if filtering by author tags or searching
	if len(params.Tags) > 0 || params.Search != "" {
		query += " LEFT JOIN recipe_tags rt ON r.id = rt.recipe_id LEFT JOIN tags t ON rt.tag_id = t.id"
	}

	// Join with user_tags if filtering by user tags or searching with user context
	if len(params.UserTags) > 0 || (params.Search != "" && params.UserID > 0) {
		query += " LEFT JOIN user_tags ut ON r.id = ut.recipe_id"
		if params.UserID > 0 {
			query += fmt.Sprintf(" AND ut.user_id = $%d", argIndex)
			args = append(args, params.UserID)
			argIndex++
		}
	}

	query += " WHERE 1=1"

	// Add fuzzy search filter (includes tags)
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

	// Add author tags filter (recipe must have ALL specified tags)
	for _, tagName := range params.Tags {
		normalizedTag := strings.ToLower(strings.TrimSpace(tagName))
		if normalizedTag == "" {
			continue
		}
		query += fmt.Sprintf(" AND r.id IN (SELECT rt2.recipe_id FROM recipe_tags rt2 INNER JOIN tags t2 ON rt2.tag_id = t2.id WHERE t2.name = $%d)", argIndex)
		args = append(args, normalizedTag)
		argIndex++
	}

	// Add user tags filter (recipe must have ALL specified user tags)
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

	// Add calories filter
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

	// Add prep time filter
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

	// Add cook time filter
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

// GetRandomRecipeID retrieves a random recipe ID from the database
func GetRandomRecipeID() (int, error) {
	var id int

	dbConnection, err := db.GetConnection()
	if err != nil {
		return 0, fmt.Errorf("failed to connect to database: %v", err)
	}
	defer dbConnection.Close()

	err = dbConnection.QueryRow("SELECT id FROM recipes ORDER BY RANDOM() LIMIT 1").Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to get random recipe: %v", err)
	}

	return id, nil
}

// GetOrCreateTag retrieves a tag by name or creates it if it doesn't exist
// Tag names are normalized to lowercase for case-insensitivity
func GetOrCreateTag(name string) (Tag, error) {
	normalizedName := strings.ToLower(strings.TrimSpace(name))
	if normalizedName == "" {
		return Tag{}, fmt.Errorf("tag name cannot be empty")
	}

	dbConnection, err := db.GetConnection()
	if err != nil {
		return Tag{}, fmt.Errorf("failed to connect to database: %v", err)
	}
	defer dbConnection.Close()

	var tag Tag
	err = dbConnection.QueryRow("SELECT id, name FROM tags WHERE name = $1", normalizedName).Scan(&tag.ID, &tag.Name)
	if err == nil {
		return tag, nil
	}
	if err != sql.ErrNoRows {
		return Tag{}, fmt.Errorf("failed to query tag: %v", err)
	}

	err = dbConnection.QueryRow("INSERT INTO tags (name) VALUES ($1) RETURNING id, name", normalizedName).Scan(&tag.ID, &tag.Name)
	if err != nil {
		return Tag{}, fmt.Errorf("failed to create tag: %v", err)
	}

	return tag, nil
}

// SearchTags searches for tags matching a query (fuzzy search)
func SearchTags(query string) ([]Tag, error) {
	dbConnection, err := db.GetConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}
	defer dbConnection.Close()

	searchPattern := "%" + strings.ToLower(strings.TrimSpace(query)) + "%"
	rows, err := dbConnection.Query("SELECT id, name FROM tags WHERE name LIKE $1 ORDER BY name LIMIT 20", searchPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search tags: %v", err)
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var tag Tag
		if err := rows.Scan(&tag.ID, &tag.Name); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %v", err)
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// GetTagsByRecipeID retrieves all author tags for a recipe
func GetTagsByRecipeID(recipeID int) ([]Tag, error) {
	dbConnection, err := db.GetConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}
	defer dbConnection.Close()

	rows, err := dbConnection.Query(`
		SELECT t.id, t.name 
		FROM tags t 
		INNER JOIN recipe_tags rt ON t.id = rt.tag_id 
		WHERE rt.recipe_id = $1 
		ORDER BY t.name`, recipeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %v", err)
	}
	defer rows.Close()

	var tags []Tag
	for rows.Next() {
		var tag Tag
		if err := rows.Scan(&tag.ID, &tag.Name); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %v", err)
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// GetTagsForRecipes retrieves tags for multiple recipes in a single query
// Returns a map of recipe ID to slice of tags
func GetTagsForRecipes(recipeIDs []int) (map[int][]Tag, error) {
	result := make(map[int][]Tag)
	if len(recipeIDs) == 0 {
		return result, nil
	}

	dbConnection, err := db.GetConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}
	defer dbConnection.Close()

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

	rows, err := dbConnection.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var recipeID int
		var tag Tag
		if err := rows.Scan(&recipeID, &tag.ID, &tag.Name); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %v", err)
		}
		result[recipeID] = append(result[recipeID], tag)
	}

	return result, nil
}

// AddTagToRecipe adds an author tag to a recipe
func AddTagToRecipe(recipeID int, tagID int) error {
	dbConnection, err := db.GetConnection()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer dbConnection.Close()

	_, err = dbConnection.Exec("INSERT INTO recipe_tags (recipe_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING", recipeID, tagID)
	if err != nil {
		return fmt.Errorf("failed to add tag to recipe: %v", err)
	}

	return nil
}

// RemoveTagFromRecipe removes an author tag from a recipe
func RemoveTagFromRecipe(recipeID int, tagID int) error {
	dbConnection, err := db.GetConnection()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer dbConnection.Close()

	_, err = dbConnection.Exec("DELETE FROM recipe_tags WHERE recipe_id = $1 AND tag_id = $2", recipeID, tagID)
	if err != nil {
		return fmt.Errorf("failed to remove tag from recipe: %v", err)
	}

	return nil
}

// SetRecipeTags replaces all author tags on a recipe with the given tag names
func SetRecipeTags(recipeID int, tagNames []string) error {
	dbConnection, err := db.GetConnection()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer dbConnection.Close()

	tx, err := dbConnection.Begin()
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

// GetOrCreateUserTag retrieves or creates a user tag for a specific recipe
func GetOrCreateUserTag(userID int, recipeID int, name string) (UserTag, error) {
	normalizedName := strings.ToLower(strings.TrimSpace(name))
	if normalizedName == "" {
		return UserTag{}, fmt.Errorf("tag name cannot be empty")
	}

	dbConnection, err := db.GetConnection()
	if err != nil {
		return UserTag{}, fmt.Errorf("failed to connect to database: %v", err)
	}
	defer dbConnection.Close()

	var userTag UserTag
	err = dbConnection.QueryRow(
		"SELECT id, user_id, recipe_id, name FROM user_tags WHERE user_id = $1 AND recipe_id = $2 AND name = $3",
		userID, recipeID, normalizedName,
	).Scan(&userTag.ID, &userTag.UserID, &userTag.RecipeID, &userTag.Name)
	if err == nil {
		return userTag, nil
	}

	err = dbConnection.QueryRow(
		"INSERT INTO user_tags (user_id, recipe_id, name) VALUES ($1, $2, $3) RETURNING id, user_id, recipe_id, name",
		userID, recipeID, normalizedName,
	).Scan(&userTag.ID, &userTag.UserID, &userTag.RecipeID, &userTag.Name)
	if err != nil {
		return UserTag{}, fmt.Errorf("failed to create user tag: %v", err)
	}

	return userTag, nil
}

// SearchUserTags searches for user tags matching a query (for a specific user)
func SearchUserTags(userID int, query string) ([]string, error) {
	dbConnection, err := db.GetConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}
	defer dbConnection.Close()

	searchPattern := "%" + strings.ToLower(strings.TrimSpace(query)) + "%"
	rows, err := dbConnection.Query(
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

// GetUserTagsByRecipeID retrieves all user tags for a specific user on a recipe
func GetUserTagsByRecipeID(userID int, recipeID int) ([]UserTag, error) {
	dbConnection, err := db.GetConnection()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %v", err)
	}
	defer dbConnection.Close()

	rows, err := dbConnection.Query(
		"SELECT id, user_id, recipe_id, name FROM user_tags WHERE user_id = $1 AND recipe_id = $2 ORDER BY name",
		userID, recipeID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user tags: %v", err)
	}
	defer rows.Close()

	var tags []UserTag
	for rows.Next() {
		var tag UserTag
		if err := rows.Scan(&tag.ID, &tag.UserID, &tag.RecipeID, &tag.Name); err != nil {
			return nil, fmt.Errorf("failed to scan user tag: %v", err)
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// RemoveUserTag removes a user tag by ID
func RemoveUserTag(userID int, tagID int) error {
	dbConnection, err := db.GetConnection()
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer dbConnection.Close()

	_, err = dbConnection.Exec("DELETE FROM user_tags WHERE id = $1 AND user_id = $2", tagID, userID)
	if err != nil {
		return fmt.Errorf("failed to remove user tag: %v", err)
	}

	return nil
}
