package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/mr-flannery/go-recipe-book/src/auth"
	"github.com/mr-flannery/go-recipe-book/src/db"
	"github.com/mr-flannery/go-recipe-book/src/models"
)

type TagSearchResponse struct {
	Tags []string `json:"tags"`
}

type TagResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

func SearchTagsHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	tags, err := models.SearchTags(query)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Failed to search tags"})
		return
	}

	tagNames := make([]string, len(tags))
	for i, tag := range tags {
		tagNames[i] = tag.Name
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TagSearchResponse{Tags: tagNames})
}

func SearchUserTagsHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	database, err := db.GetConnection()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Database connection error"})
		return
	}
	defer database.Close()

	user, err := auth.GetUserBySession(database, r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Unauthorized"})
		return
	}

	tags, err := models.SearchUserTags(user.ID, query)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Failed to search user tags"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TagSearchResponse{Tags: tags})
}

func AddTagToRecipeHandler(w http.ResponseWriter, r *http.Request) {
	recipeID := r.PathValue("id")
	if recipeID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Recipe ID is required"})
		return
	}

	recipeIDInt, err := strconv.Atoi(recipeID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Invalid recipe ID"})
		return
	}

	database, err := db.GetConnection()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Database connection error"})
		return
	}
	defer database.Close()

	user, err := auth.GetUserBySession(database, r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Unauthorized"})
		return
	}

	recipe, err := models.GetRecipeByID(recipeID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Recipe not found"})
		return
	}

	if user.ID != recipe.AuthorID {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Only the recipe author can add tags"})
		return
	}

	if err := r.ParseMultipartForm(1024); err != nil {
		r.ParseForm()
	}
	tagName := strings.TrimSpace(r.FormValue("tag"))
	if tagName == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Tag name is required"})
		return
	}

	tag, err := models.GetOrCreateTag(tagName)
	if err != nil {
		log.Printf("ERROR: GetOrCreateTag failed for tag '%s': %v", tagName, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Failed to create tag"})
		return
	}

	err = models.AddTagToRecipe(recipeIDInt, tag.ID)
	if err != nil {
		log.Printf("ERROR: AddTagToRecipe failed for recipe %d, tag %d: %v", recipeIDInt, tag.ID, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Failed to add tag to recipe"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TagResponse{Success: true})
}

func RemoveTagFromRecipeHandler(w http.ResponseWriter, r *http.Request) {
	recipeID := r.PathValue("id")
	tagID := r.PathValue("tagId")
	if recipeID == "" || tagID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Recipe ID and Tag ID are required"})
		return
	}

	recipeIDInt, err := strconv.Atoi(recipeID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Invalid recipe ID"})
		return
	}

	tagIDInt, err := strconv.Atoi(tagID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Invalid tag ID"})
		return
	}

	database, err := db.GetConnection()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Database connection error"})
		return
	}
	defer database.Close()

	user, err := auth.GetUserBySession(database, r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Unauthorized"})
		return
	}

	recipe, err := models.GetRecipeByID(recipeID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Recipe not found"})
		return
	}

	if user.ID != recipe.AuthorID {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Only the recipe author can remove tags"})
		return
	}

	err = models.RemoveTagFromRecipe(recipeIDInt, tagIDInt)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Failed to remove tag from recipe"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TagResponse{Success: true})
}

func AddUserTagToRecipeHandler(w http.ResponseWriter, r *http.Request) {
	recipeID := r.PathValue("id")
	if recipeID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Recipe ID is required"})
		return
	}

	recipeIDInt, err := strconv.Atoi(recipeID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Invalid recipe ID"})
		return
	}

	database, err := db.GetConnection()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Database connection error"})
		return
	}
	defer database.Close()

	user, err := auth.GetUserBySession(database, r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Unauthorized"})
		return
	}

	_, err = models.GetRecipeByID(recipeID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Recipe not found"})
		return
	}

	if err := r.ParseMultipartForm(1024); err != nil {
		r.ParseForm()
	}
	tagName := strings.TrimSpace(r.FormValue("tag"))
	if tagName == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Tag name is required"})
		return
	}

	_, err = models.GetOrCreateUserTag(user.ID, recipeIDInt, tagName)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Failed to add user tag"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TagResponse{Success: true})
}

func RemoveUserTagHandler(w http.ResponseWriter, r *http.Request) {
	tagID := r.PathValue("tagId")
	if tagID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Tag ID is required"})
		return
	}

	tagIDInt, err := strconv.Atoi(tagID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Invalid tag ID"})
		return
	}

	database, err := db.GetConnection()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Database connection error"})
		return
	}
	defer database.Close()

	user, err := auth.GetUserBySession(database, r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Unauthorized"})
		return
	}

	err = models.RemoveUserTag(user.ID, tagIDInt)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Failed to remove user tag"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TagResponse{Success: true})
}
