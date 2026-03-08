package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/mr-flannery/go-recipe-book/src/auth"
	"github.com/mr-flannery/go-recipe-book/src/logging"
)

type TagSearchResponse struct {
	Tags []string `json:"tags"`
}

type TagResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

func (h *Handler) SearchTagsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query().Get("q")

	tags, err := h.TagStore.Search(ctx, query)
	if err != nil {
		logging.AddError(ctx, err, "Failed to search tags")
		logging.Add(ctx, "action", "tag.search")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Failed to search tags"})
		return
	}

	tagNames := make([]string, len(tags))
	for i, tag := range tags {
		tagNames[i] = tag.Name
	}

	logging.AddMany(ctx, map[string]any{
		"action":       "tag.search",
		"result.count": len(tagNames),
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TagSearchResponse{Tags: tagNames})
}

func (h *Handler) SearchUserTagsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	query := r.URL.Query().Get("q")

	user, err := auth.GetUserBySession(ctx, h.AuthStore, r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Unauthorized"})
		return
	}

	tags, err := h.UserTagStore.Search(ctx, user.ID, query)
	if err != nil {
		logging.AddError(ctx, err, "Failed to search user tags")
		logging.Add(ctx, "action", "user_tag.search")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Failed to search user tags"})
		return
	}

	logging.AddMany(ctx, map[string]any{
		"action":       "user_tag.search",
		"result.count": len(tags),
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TagSearchResponse{Tags: tags})
}

func (h *Handler) AddTagToRecipeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
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

	user, err := auth.GetUserBySession(ctx, h.AuthStore, r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Unauthorized"})
		return
	}

	recipe, err := h.RecipeStore.GetByID(ctx, recipeID)
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

	tag, err := h.TagStore.GetOrCreate(ctx, tagName)
	if err != nil {
		logging.AddError(ctx, err, "Failed to get or create tag")
		logging.AddMany(ctx, map[string]any{
			"action":   "tag.add",
			"tag.name": tagName,
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Failed to create tag"})
		return
	}

	err = h.TagStore.AddToRecipe(ctx, recipeIDInt, tag.ID)
	if err != nil {
		logging.AddError(ctx, err, "Failed to add tag to recipe")
		logging.AddMany(ctx, map[string]any{
			"action":    "tag.add",
			"recipe.id": recipeIDInt,
			"tag.id":    tag.ID,
			"tag.name":  tagName,
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Failed to add tag to recipe"})
		return
	}

	logging.AddMany(ctx, map[string]any{
		"action":    "tag.add",
		"recipe.id": recipeIDInt,
		"tag.id":    tag.ID,
		"tag.name":  tagName,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TagResponse{Success: true})
}

func (h *Handler) RemoveTagFromRecipeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
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

	user, err := auth.GetUserBySession(ctx, h.AuthStore, r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Unauthorized"})
		return
	}

	recipe, err := h.RecipeStore.GetByID(ctx, recipeID)
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

	err = h.TagStore.RemoveFromRecipe(ctx, recipeIDInt, tagIDInt)
	if err != nil {
		logging.AddError(ctx, err, "Failed to remove tag from recipe")
		logging.AddMany(ctx, map[string]any{
			"action":    "tag.remove",
			"recipe.id": recipeIDInt,
			"tag.id":    tagIDInt,
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Failed to remove tag from recipe"})
		return
	}

	logging.AddMany(ctx, map[string]any{
		"action":    "tag.remove",
		"recipe.id": recipeIDInt,
		"tag.id":    tagIDInt,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TagResponse{Success: true})
}

func (h *Handler) AddUserTagToRecipeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
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

	user, err := auth.GetUserBySession(ctx, h.AuthStore, r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Unauthorized"})
		return
	}

	_, err = h.RecipeStore.GetByID(ctx, recipeID)
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

	_, err = h.UserTagStore.GetOrCreate(ctx, user.ID, recipeIDInt, tagName)
	if err != nil {
		logging.AddError(ctx, err, "Failed to add user tag")
		logging.AddMany(ctx, map[string]any{
			"action":        "user_tag.add",
			"recipe.id":     recipeIDInt,
			"user_tag.name": tagName,
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Failed to add user tag"})
		return
	}

	logging.AddMany(ctx, map[string]any{
		"action":        "user_tag.add",
		"recipe.id":     recipeIDInt,
		"user_tag.name": tagName,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TagResponse{Success: true})
}

func (h *Handler) RemoveUserTagHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
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

	user, err := auth.GetUserBySession(ctx, h.AuthStore, r)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Unauthorized"})
		return
	}

	err = h.UserTagStore.Remove(ctx, user.ID, tagIDInt)
	if err != nil {
		logging.AddError(ctx, err, "Failed to remove user tag")
		logging.AddMany(ctx, map[string]any{
			"action":      "user_tag.remove",
			"user_tag.id": tagIDInt,
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(TagResponse{Success: false, Error: "Failed to remove user tag"})
		return
	}

	logging.AddMany(ctx, map[string]any{
		"action":      "user_tag.remove",
		"user_tag.id": tagIDInt,
	})

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(TagResponse{Success: true})
}
