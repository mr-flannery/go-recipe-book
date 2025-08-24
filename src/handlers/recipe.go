package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/mr-flannery/go-recipe-book/src/auth"
	"github.com/mr-flannery/go-recipe-book/src/db"
	"github.com/mr-flannery/go-recipe-book/src/models"
)

var recipeTemplates = template.Must(template.ParseGlob("templates/recipes/*.gohtml"))

// GetCreateRecipeHandler handles displaying the create recipe form
func GetCreateRecipeHandler(w http.ResponseWriter, r *http.Request) {
	err := recipeTemplates.ExecuteTemplate(w, "create.gohtml", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// PostCreateRecipeHandler handles the creation of a new recipe
func PostCreateRecipeHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	// Get database connection
	database, err := db.GetConnection()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer database.Close()

	// Get the logged-in user
	user, err := auth.GetUserBySession(database, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	recipe := models.Recipe{
		Title:          r.FormValue("title"),
		IngredientsMD:  r.FormValue("ingredients"),
		InstructionsMD: r.FormValue("instructions"),
		AuthorID:       user.ID, // Set the correct AuthorID
	}
	if err := models.SaveRecipe(recipe); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save recipe: %v", err), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/recipes", http.StatusSeeOther)
}

// ListRecipesHandler lists all recipes
func ListRecipesHandler(w http.ResponseWriter, r *http.Request) {
	recipes, err := models.GetAllRecipes()
	if err != nil {
		http.Error(w, "Failed to fetch recipes", http.StatusInternalServerError)
		return
	}

	// Get database connection to check user authentication
	database, err := db.GetConnection()
	if err != nil {
		// If DB fails, assume not logged in
		data := struct {
			Recipes     []models.Recipe
			IsLoggedIn  bool
			CurrentUser *auth.User
		}{
			Recipes:     recipes,
			IsLoggedIn:  false,
			CurrentUser: nil,
		}
		err = recipeTemplates.ExecuteTemplate(w, "list.gohtml", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	defer database.Close()

	currentUser, err := auth.GetUserBySession(database, r)
	isLoggedIn := err == nil

	data := struct {
		Recipes     []models.Recipe
		IsLoggedIn  bool
		CurrentUser *auth.User
	}{
		Recipes:     recipes,
		IsLoggedIn:  isLoggedIn,
		CurrentUser: currentUser,
	}

	err = recipeTemplates.ExecuteTemplate(w, "list.gohtml", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// GetUpdateRecipeHandler handles displaying the update recipe form
func GetUpdateRecipeHandler(w http.ResponseWriter, r *http.Request) {
	recipeID := r.URL.Query().Get("id")
	recipe, err := models.GetRecipeByID(recipeID)
	if err != nil {
		http.Error(w, "Recipe not found", http.StatusNotFound)
		return
	}
	err = recipeTemplates.ExecuteTemplate(w, "update.gohtml", recipe)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// PostUpdateRecipeHandler handles updating an existing recipe
func PostUpdateRecipeHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	recipeID, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		// ... handle error
		http.Error(
			w,
			fmt.Sprintf("Failed to update recipe: failed to convert ID to int. %s", err.Error()),
			http.StatusInternalServerError,
		)
	}

	updatedRecipe := models.Recipe{
		ID:             recipeID,
		Title:          r.FormValue("title"),
		IngredientsMD:  r.FormValue("ingredients"),
		InstructionsMD: r.FormValue("instructions"),
	}
	if err := models.UpdateRecipe(updatedRecipe); err != nil {
		http.Error(w, "Failed to update recipe", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/recipes", http.StatusSeeOther)
}

// ViewRecipeHandler handles viewing a single recipe with comments
func ViewRecipeHandler(w http.ResponseWriter, r *http.Request) {
	recipeID := r.PathValue("id")
	if recipeID == "" {
		// TODO this should show a dedicated not found page with a link back to the overview page
		http.Error(w, "Recipe ID is required", http.StatusBadRequest)
		return
	}

	// Handle GET request - display recipe and comments
	recipe, err := models.GetRecipeByID(recipeID)
	if err != nil {
		http.Error(w, "Recipe not found", http.StatusNotFound)
		return
	}

	comments, err := models.GetCommentsByRecipeID(recipeID)
	if err != nil {
		http.Error(w, "Failed to fetch comments", http.StatusInternalServerError)
		return
	}

	// Add usernames to comments
	type CommentWithUsername struct {
		models.Comment
		Username string
	}

	var commentsWithUsernames []CommentWithUsername
	for _, comment := range comments {
		username, err := models.GetUsernameByID(comment.AuthorID)
		if err != nil {
			username = "Unknown User"
		}
		commentsWithUsernames = append(commentsWithUsernames, CommentWithUsername{
			Comment:  comment,
			Username: username,
		})
	}

	// Check if user is logged in and get user info
	database, err := db.GetConnection()
	if err != nil {
		// If DB fails, assume not logged in
		data := struct {
			Recipe     models.Recipe
			Comments   []CommentWithUsername
			IsLoggedIn bool
			CurrentUser *auth.User
			IsAuthor   bool
		}{
			Recipe:      recipe,
			Comments:    commentsWithUsernames,
			IsLoggedIn:  false,
			CurrentUser: nil,
			IsAuthor:    false,
		}
		err = recipeTemplates.ExecuteTemplate(w, "view.gohtml", data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	defer database.Close()

	currentUser, err := auth.GetUserBySession(database, r)
	isLoggedIn := err == nil
	isAuthor := isLoggedIn && currentUser.ID == recipe.AuthorID

	data := struct {
		Recipe     models.Recipe
		Comments   []CommentWithUsername
		IsLoggedIn bool
		CurrentUser *auth.User
		IsAuthor   bool
	}{
		Recipe:      recipe,
		Comments:    commentsWithUsernames,
		IsLoggedIn:  isLoggedIn,
		CurrentUser: currentUser,
		IsAuthor:    isAuthor,
	}

	err = recipeTemplates.ExecuteTemplate(w, "view.gohtml", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// CommentHTMXHandler handles adding comments via HTMX and returns HTML fragment
func CommentHTMXHandler(w http.ResponseWriter, r *http.Request) {
	recipeID := r.PathValue("id")
	if recipeID == "" {
		http.Error(w, "Recipe ID is required", http.StatusBadRequest)
		return
	}

	// Get database connection
	database, err := db.GetConnection()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer database.Close()

	// Check authentication
	user, err := auth.GetUserBySession(database, r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse form data
	r.ParseForm()
	commentContent := r.FormValue("comment")
	if commentContent == "" {
		http.Error(w, "Comment content is required", http.StatusBadRequest)
		return
	}

	// Convert recipe ID to int
	recipeIDInt, err := strconv.Atoi(recipeID)
	if err != nil {
		http.Error(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}

	// Create and save comment
	comment := models.Comment{
		RecipeID:  recipeIDInt,
		AuthorID:  user.ID,
		ContentMD: commentContent,
	}

	if err := models.SaveComment(comment); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save comment: %v", err), http.StatusInternalServerError)
		return
	}

	// Get the saved comment with timestamp
	savedComment, err := models.GetLatestCommentByUserAndRecipe(user.ID, recipeIDInt)
	if err != nil {
		http.Error(w, "Failed to retrieve saved comment", http.StatusInternalServerError)
		return
	}

	// Create comment data with username for template
	type CommentWithUsername struct {
		models.Comment
		Username string
	}

	commentData := CommentWithUsername{
		Comment:  savedComment,
		Username: user.Username,
	}

	// Return HTML fragment
	w.Header().Set("Content-Type", "text/html")
	err = recipeTemplates.ExecuteTemplate(w, "comment.gohtml", commentData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// DeleteRecipeHandler handles deleting a recipe
func DeleteRecipeHandler(w http.ResponseWriter, r *http.Request) {
	recipeID := r.PathValue("id")
	if recipeID == "" {
		http.Error(w, "Recipe ID is required", http.StatusBadRequest)
		return
	}

	// Get database connection
	database, err := db.GetConnection()
	if err != nil {
		http.Error(w, "Database connection error", http.StatusInternalServerError)
		return
	}
	defer database.Close()

	// Get the current user
	currentUser, err := auth.GetUserBySession(database, r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Get the recipe to check ownership
	recipe, err := models.GetRecipeByID(recipeID)
	if err != nil {
		http.Error(w, "Recipe not found", http.StatusNotFound)
		return
	}

	// Check if the current user is the author of the recipe
	if currentUser.ID != recipe.AuthorID {
		http.Error(w, "Forbidden: You can only delete your own recipes", http.StatusForbidden)
		return
	}

	// Delete the recipe
	if err := models.DeleteRecipe(recipeID); err != nil {
		http.Error(w, "Failed to delete recipe", http.StatusInternalServerError)
		return
	}

	// For DELETE requests, return a success response instead of redirect
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Recipe deleted successfully"))
}
