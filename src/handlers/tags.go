package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/mr-flannery/go-recipe-book/src/auth"
)

type TagSearchResponse struct {
	Tags []string `json:"tags"`
}

type TagResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

func (h *Handler) SearchTagsHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	tags, err := h.TagStore.Search(query)
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

func (h *Handler) SearchUserTagsHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")

	user, err := auth.GetUserBySession(h.AuthStore, r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Unauthorized"})
		return
	}

	tags, err := h.UserTagStore.Search(user.ID, query)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Failed to search user tags"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TagSearchResponse{Tags: tags})
}

func (h *Handler) AddTagToRecipeHandler(w http.ResponseWriter, r *http.Request) {
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

	user, err := auth.GetUserBySession(h.AuthStore, r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Unauthorized"})
		return
	}

	recipe, err := h.RecipeStore.GetByID(recipeID)
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

	tag, err := h.TagStore.GetOrCreate(tagName)
	if err != nil {
		log.Printf("ERROR: GetOrCreateTag failed for tag '%s': %v", tagName, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Failed to create tag"})
		return
	}

	err = h.TagStore.AddToRecipe(recipeIDInt, tag.ID)
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

func (h *Handler) RemoveTagFromRecipeHandler(w http.ResponseWriter, r *http.Request) {
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

	user, err := auth.GetUserBySession(h.AuthStore, r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Unauthorized"})
		return
	}

	recipe, err := h.RecipeStore.GetByID(recipeID)
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

	err = h.TagStore.RemoveFromRecipe(recipeIDInt, tagIDInt)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Failed to remove tag from recipe"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TagResponse{Success: true})
}

func (h *Handler) AddUserTagToRecipeHandler(w http.ResponseWriter, r *http.Request) {
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

	user, err := auth.GetUserBySession(h.AuthStore, r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Unauthorized"})
		return
	}

	_, err = h.RecipeStore.GetByID(recipeID)
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

	_, err = h.UserTagStore.GetOrCreate(user.ID, recipeIDInt, tagName)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Failed to add user tag"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TagResponse{Success: true})
}

func (h *Handler) RemoveUserTagHandler(w http.ResponseWriter, r *http.Request) {
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

	user, err := auth.GetUserBySession(h.AuthStore, r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Unauthorized"})
		return
	}

	err = h.UserTagStore.Remove(user.ID, tagIDInt)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Failed to remove user tag"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TagResponse{Success: true})
}
