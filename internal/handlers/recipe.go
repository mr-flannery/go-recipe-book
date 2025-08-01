package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/yourusername/agent-coding-recipe-book/auth"
	"github.com/yourusername/agent-coding-recipe-book/internal/models"
)

var recipeTemplates = template.Must(template.ParseGlob("templates/recipes/*.gohtml"))

// CreateRecipeHandler handles the creation of a new recipe
func CreateRecipeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		err := recipeTemplates.ExecuteTemplate(w, "create.gohtml", nil)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	if r.Method == http.MethodPost {
		r.ParseForm()

		// Get the logged-in user
		username, isLoggedIn := auth.GetUser(r)
		if !isLoggedIn {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Fetch the user ID from the database
		userID, err := auth.GetUserIDByUsername(username)
		if err != nil {
			http.Error(w, "Failed to fetch user ID", http.StatusInternalServerError)
			return
		}

		recipe := models.Recipe{
			Title:          r.FormValue("title"),
			IngredientsMD:  r.FormValue("ingredients"),
			InstructionsMD: r.FormValue("instructions"),
			AuthorID:       userID, // Set the correct AuthorID
		}
		if err := models.SaveRecipe(recipe); err != nil {
			http.Error(w, fmt.Sprintf("Failed to save recipe: %v", err), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/recipes", http.StatusSeeOther)
	}
}

// ListRecipesHandler lists all recipes
func ListRecipesHandler(w http.ResponseWriter, r *http.Request) {
	recipes, err := models.GetAllRecipes()
	if err != nil {
		http.Error(w, "Failed to fetch recipes", http.StatusInternalServerError)
		return
	}

	err = recipeTemplates.ExecuteTemplate(w, "list.gohtml", recipes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// UpdateRecipeHandler handles updating an existing recipe
func UpdateRecipeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
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
		return
	}
	if r.Method == http.MethodPost {
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
}

// DeleteRecipeHandler handles deleting a recipe
func DeleteRecipeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		recipeID := r.FormValue("id")
		if err := models.DeleteRecipe(recipeID); err != nil {
			http.Error(w, "Failed to delete recipe", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/recipes", http.StatusSeeOther)
	}
}
