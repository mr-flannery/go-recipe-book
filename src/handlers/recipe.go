package handlers

import (
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/mr-flannery/go-recipe-book/src/auth"
	"github.com/mr-flannery/go-recipe-book/src/models"
)

func (h *Handler) GetCreateRecipeHandler(w http.ResponseWriter, r *http.Request) {
	data := struct {
		UserInfo *auth.UserInfo
	}{
		UserInfo: auth.GetUserInfoFromContext(r.Context()),
	}
	h.Renderer.RenderPage(w, "create.gohtml", data)
}

func (h *Handler) PostCreateRecipeHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		slog.Error("Failed to parse multipart form", "error", err)
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	user, err := auth.GetUserBySession(h.AuthStore, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var prepTime, cookTime, calories int
	if prepTimeStr := r.FormValue("preptime"); prepTimeStr != "" {
		prepTime, err = strconv.Atoi(prepTimeStr)
		if err != nil {
			slog.Error("Invalid prep time", "value", prepTimeStr, "error", err)
			http.Error(w, "Invalid prep time", http.StatusBadRequest)
			return
		}
	}

	if cookTimeStr := r.FormValue("cooktime"); cookTimeStr != "" {
		cookTime, err = strconv.Atoi(cookTimeStr)
		if err != nil {
			slog.Error("Invalid cook time", "value", cookTimeStr, "error", err)
			http.Error(w, "Invalid cook time", http.StatusBadRequest)
			return
		}
	}

	if caloriesStr := r.FormValue("calories"); caloriesStr != "" {
		calories, err = strconv.Atoi(caloriesStr)
		if err != nil {
			slog.Error("Invalid calories", "value", caloriesStr, "error", err)
			http.Error(w, "Invalid calories", http.StatusBadRequest)
			return
		}
	}

	var imageData []byte

	croppedImageData := r.FormValue("cropped_image_data")
	if croppedImageData != "" {
		if strings.HasPrefix(croppedImageData, "data:image/") {
			commaIndex := strings.Index(croppedImageData, ",")
			if commaIndex != -1 {
				croppedImageData = croppedImageData[commaIndex+1:]
			}
		}

		decodedData, err := base64.StdEncoding.DecodeString(croppedImageData)
		if err != nil {
			slog.Error("Failed to decode cropped image data", "error", err)
			http.Error(w, "Failed to process cropped image", http.StatusBadRequest)
			return
		}
		imageData = decodedData
		slog.Info("Cropped image processed", "size", len(imageData))
	} else {
		file, _, err := r.FormFile("image")
		if err == nil {
			defer file.Close()
			imageData, err = io.ReadAll(file)
			if err != nil {
				slog.Error("Failed to read image file", "error", err)
				http.Error(w, "Failed to read image file", http.StatusInternalServerError)
				return
			}
			slog.Info("Original image uploaded", "size", len(imageData))
		} else if err != http.ErrMissingFile {
			slog.Error("Error processing image file", "error", err)
			http.Error(w, "Error processing image file", http.StatusBadRequest)
			return
		}
	}

	recipe := models.Recipe{
		Title:          r.FormValue("title"),
		IngredientsMD:  r.FormValue("ingredients"),
		InstructionsMD: r.FormValue("instructions"),
		PrepTime:       prepTime,
		CookTime:       cookTime,
		Calories:       calories,
		Image:          imageData,
		AuthorID:       user.ID,
	}

	recipeID, err := h.RecipeStore.Save(recipe)
	if err != nil {
		slog.Error("Failed to save recipe", "error", err)
		http.Error(w, fmt.Sprintf("Failed to save recipe: %v", err), http.StatusInternalServerError)
		return
	}

	tagsStr := r.FormValue("tags")
	if tagsStr != "" {
		tagNames := strings.Split(tagsStr, ",")
		if err := h.TagStore.SetRecipeTags(recipeID, tagNames); err != nil {
			slog.Error("Failed to set recipe tags", "error", err)
		}
	}

	slog.Info("Recipe created successfully", "id", recipeID, "title", recipe.Title, "author", user.Username)
	http.Redirect(w, r, fmt.Sprintf("/recipes/%d", recipeID), http.StatusSeeOther)
}

// might want to make this configurable at some point
const RecipesPerPage = 20

func (h *Handler) ListRecipesHandler(w http.ResponseWriter, r *http.Request) {
	filterParams := models.FilterParams{
		Limit: RecipesPerPage,
	}
	recipes, err := h.RecipeStore.GetFiltered(filterParams)
	if err != nil {
		http.Error(w, "Failed to fetch recipes", http.StatusInternalServerError)
		return
	}

	recipeIDs := make([]int, len(recipes))
	for i, r := range recipes {
		recipeIDs[i] = r.ID
	}
	tagsMap, _ := h.TagStore.GetForRecipes(recipeIDs)

	for i := range recipes {
		recipes[i].Tags = tagsMap[recipes[i].ID]
	}

	userInfo := auth.GetUserInfoFromContext(r.Context())
	var currentUser *auth.User
	if userInfo.IsLoggedIn {
		currentUser, _ = auth.GetUserBySession(h.AuthStore, r)
		// this is slightly suboptimal for performance. we might want to write a dedicate query at some point that grabs everything at once. but let's do some performance testing first.
		userTagsMap, _ := h.UserTagStore.GetForRecipes(currentUser.ID, recipeIDs)
		for i := range recipes {
			recipes[i].UserTags = userTagsMap[recipes[i].ID]
		}
	}

	hasMore := len(recipes) == RecipesPerPage

	data := struct {
		Recipes     []models.Recipe
		UserInfo    *auth.UserInfo
		IsLoggedIn  bool
		CurrentUser *auth.User
		NextOffset  int
		HasMore     bool
	}{
		Recipes:     recipes,
		UserInfo:    userInfo,
		IsLoggedIn:  userInfo.IsLoggedIn,
		CurrentUser: currentUser,
		NextOffset:  RecipesPerPage,
		HasMore:     hasMore,
	}

	h.Renderer.RenderPage(w, "list.gohtml", data)
}

func (h *Handler) GetUpdateRecipeHandler(w http.ResponseWriter, r *http.Request) {
	recipeID := r.URL.Query().Get("id")
	recipe, err := h.RecipeStore.GetByID(recipeID)
	if err != nil {
		http.Error(w, "Recipe not found", http.StatusNotFound)
		return
	}

	recipeIDInt, _ := strconv.Atoi(recipeID)
	recipe.Tags, _ = h.TagStore.GetByRecipeID(recipeIDInt)

	data := struct {
		Recipe   models.Recipe
		UserInfo *auth.UserInfo
	}{
		Recipe:   recipe,
		UserInfo: auth.GetUserInfoFromContext(r.Context()),
	}

	h.Renderer.RenderPage(w, "update.gohtml", data)
}

func (h *Handler) PostUpdateRecipeHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		slog.Error("Failed to parse multipart form", "error", err)
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	recipeID, err := strconv.Atoi(r.FormValue("id"))
	if err != nil {
		slog.Error("Failed to convert ID to int", "id", r.FormValue("id"), "error", err)
		http.Error(w, fmt.Sprintf("Failed to update recipe: failed to convert ID to int. %s", err.Error()), http.StatusInternalServerError)
		return
	}

	user, err := auth.GetUserBySession(h.AuthStore, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	existingRecipe, err := h.RecipeStore.GetByID(strconv.Itoa(recipeID))
	if err != nil {
		slog.Error("Recipe not found", "id", recipeID, "error", err)
		http.Error(w, "Recipe not found", http.StatusNotFound)
		return
	}

	if user.ID != existingRecipe.AuthorID {
		http.Error(w, "Forbidden: You can only edit your own recipes", http.StatusForbidden)
		return
	}

	var prepTime, cookTime, calories int
	if prepTimeStr := r.FormValue("preptime"); prepTimeStr != "" {
		prepTime, err = strconv.Atoi(prepTimeStr)
		if err != nil {
			slog.Error("Invalid prep time", "value", prepTimeStr, "error", err)
			http.Error(w, "Invalid prep time", http.StatusBadRequest)
			return
		}
	}

	if cookTimeStr := r.FormValue("cooktime"); cookTimeStr != "" {
		cookTime, err = strconv.Atoi(cookTimeStr)
		if err != nil {
			slog.Error("Invalid cook time", "value", cookTimeStr, "error", err)
			http.Error(w, "Invalid cook time", http.StatusBadRequest)
			return
		}
	}

	if caloriesStr := r.FormValue("calories"); caloriesStr != "" {
		calories, err = strconv.Atoi(caloriesStr)
		if err != nil {
			slog.Error("Invalid calories", "value", caloriesStr, "error", err)
			http.Error(w, "Invalid calories", http.StatusBadRequest)
			return
		}
	}

	imageData := existingRecipe.Image
	file, _, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		imageData, err = io.ReadAll(file)
		if err != nil {
			slog.Error("Failed to read image file", "error", err)
			http.Error(w, "Failed to read image file", http.StatusInternalServerError)
			return
		}
		slog.Info("New image uploaded", "size", len(imageData))
	} else if err != http.ErrMissingFile {
		slog.Error("Error processing image file", "error", err)
		http.Error(w, "Error processing image file", http.StatusBadRequest)
		return
	}

	updatedRecipe := models.Recipe{
		ID:             recipeID,
		Title:          r.FormValue("title"),
		IngredientsMD:  r.FormValue("ingredients"),
		InstructionsMD: r.FormValue("instructions"),
		PrepTime:       prepTime,
		CookTime:       cookTime,
		Calories:       calories,
		Image:          imageData,
		AuthorID:       user.ID,
	}

	if err := h.RecipeStore.Update(updatedRecipe); err != nil {
		slog.Error("Failed to update recipe", "error", err)
		http.Error(w, "Failed to update recipe", http.StatusInternalServerError)
		return
	}

	tagsStr := r.FormValue("tags")
	if tagsStr != "" {
		tagNames := strings.Split(tagsStr, ",")
		if err := h.TagStore.SetRecipeTags(recipeID, tagNames); err != nil {
			slog.Error("Failed to set recipe tags", "error", err)
		}
	} else {
		if err := h.TagStore.SetRecipeTags(recipeID, []string{}); err != nil {
			slog.Error("Failed to clear recipe tags", "error", err)
		}
	}

	slog.Info("Recipe updated successfully", "id", recipeID, "title", updatedRecipe.Title, "author", user.Username)
	http.Redirect(w, r, fmt.Sprintf("/recipes/%d", recipeID), http.StatusSeeOther)
}

func (h *Handler) ViewRecipeHandler(w http.ResponseWriter, r *http.Request) {
	recipeID := r.PathValue("id")
	if recipeID == "" {
		http.Error(w, "Recipe ID is required", http.StatusBadRequest)
		return
	}

	recipe, err := h.RecipeStore.GetByID(recipeID)
	if err != nil {
		http.Error(w, "Recipe not found", http.StatusNotFound)
		return
	}

	recipeIDInt, _ := strconv.Atoi(recipeID)

	comments, err := h.CommentStore.GetByRecipeID(recipeID)
	if err != nil {
		http.Error(w, "Failed to fetch comments", http.StatusInternalServerError)
		return
	}

	recipe.Tags, _ = h.TagStore.GetByRecipeID(recipeIDInt)

	type CommentWithUsername struct {
		models.Comment
		Username string
	}

	var commentsWithUsernames []CommentWithUsername
	for _, comment := range comments {
		username, err := h.UserStore.GetUsernameByID(comment.AuthorID)
		if err != nil {
			username = "Unknown User"
		}
		commentsWithUsernames = append(commentsWithUsernames, CommentWithUsername{
			Comment:  comment,
			Username: username,
		})
	}

	userInfo := auth.GetUserInfoFromContext(r.Context())

	currentUser, err := auth.GetUserBySession(h.AuthStore, r)
	isLoggedIn := err == nil
	isAuthor := isLoggedIn && currentUser.ID == recipe.AuthorID

	var userTags []models.UserTag
	if isLoggedIn {
		userTags, _ = h.UserTagStore.GetByRecipeID(currentUser.ID, recipeIDInt)
	}

	data := struct {
		Recipe      models.Recipe
		UserTags    []models.UserTag
		Comments    []CommentWithUsername
		IsLoggedIn  bool
		CurrentUser *auth.User
		IsAuthor    bool
		UserInfo    *auth.UserInfo
	}{
		Recipe:      recipe,
		UserTags:    userTags,
		Comments:    commentsWithUsernames,
		IsLoggedIn:  isLoggedIn,
		CurrentUser: currentUser,
		IsAuthor:    isAuthor,
		UserInfo:    userInfo,
	}

	h.Renderer.RenderPage(w, "view.gohtml", data)
}

func (h *Handler) CommentHTMXHandler(w http.ResponseWriter, r *http.Request) {
	recipeID := r.PathValue("id")
	if recipeID == "" {
		http.Error(w, "Recipe ID is required", http.StatusBadRequest)
		return
	}

	user, err := auth.GetUserBySession(h.AuthStore, r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	r.ParseForm()
	commentContent := r.FormValue("comment")
	if commentContent == "" {
		http.Error(w, "Comment content is required", http.StatusBadRequest)
		return
	}

	recipeIDInt, err := strconv.Atoi(recipeID)
	if err != nil {
		http.Error(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}

	comment := models.Comment{
		RecipeID:  recipeIDInt,
		AuthorID:  user.ID,
		ContentMD: commentContent,
	}

	if err := h.CommentStore.Save(comment); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save comment: %v", err), http.StatusInternalServerError)
		return
	}

	savedComment, err := h.CommentStore.GetLatestByUserAndRecipe(user.ID, recipeIDInt)
	if err != nil {
		http.Error(w, "Failed to retrieve saved comment", http.StatusInternalServerError)
		return
	}

	type CommentWithUsername struct {
		models.Comment
		Username string
	}

	commentData := CommentWithUsername{
		Comment:  savedComment,
		Username: user.Username,
	}

	h.Renderer.RenderFragment(w, "comment.gohtml", commentData)
}

func (h *Handler) DeleteRecipeHandler(w http.ResponseWriter, r *http.Request) {
	recipeID := r.PathValue("id")
	if recipeID == "" {
		http.Error(w, "Recipe ID is required", http.StatusBadRequest)
		return
	}

	currentUser, err := auth.GetUserBySession(h.AuthStore, r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	recipe, err := h.RecipeStore.GetByID(recipeID)
	if err != nil {
		http.Error(w, "Recipe not found", http.StatusNotFound)
		return
	}

	if currentUser.ID != recipe.AuthorID {
		http.Error(w, "Forbidden: You can only delete your own recipes", http.StatusForbidden)
		return
	}

	if err := h.RecipeStore.Delete(recipeID); err != nil {
		http.Error(w, "Failed to delete recipe", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Recipe deleted successfully"))
}

func (h *Handler) RandomRecipeHandler(w http.ResponseWriter, r *http.Request) {
	recipeID, err := h.RecipeStore.GetRandomID()
	if err != nil {
		slog.Error("Failed to get random recipe", "error", err)
		http.Redirect(w, r, "/recipes", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/recipes/%d", recipeID), http.StatusSeeOther)
}

func (h *Handler) FilterRecipesHTMXHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	filterParams := models.FilterParams{
		Search: strings.TrimSpace(r.FormValue("search")),
	}

	if caloriesStr := r.FormValue("calories_value"); caloriesStr != "" {
		if calories, err := strconv.Atoi(caloriesStr); err == nil && calories > 0 {
			filterParams.CaloriesValue = calories
			filterParams.CaloriesOp = r.FormValue("calories_op")
		}
	}

	if prepTimeStr := r.FormValue("prep_time_value"); prepTimeStr != "" {
		if prepTime, err := strconv.Atoi(prepTimeStr); err == nil && prepTime > 0 {
			filterParams.PrepTimeValue = prepTime
			filterParams.PrepTimeOp = r.FormValue("prep_time_op")
		}
	}

	if cookTimeStr := r.FormValue("cook_time_value"); cookTimeStr != "" {
		if cookTime, err := strconv.Atoi(cookTimeStr); err == nil && cookTime > 0 {
			filterParams.CookTimeValue = cookTime
			filterParams.CookTimeOp = r.FormValue("cook_time_op")
		}
	}

	if tagsStr := r.FormValue("tags"); tagsStr != "" {
		tags := strings.Split(tagsStr, ",")
		for _, tag := range tags {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				filterParams.Tags = append(filterParams.Tags, tag)
			}
		}
	}

	slog.Info("Filtering recipes", "params", filterParams)

	recipes, err := h.RecipeStore.GetFiltered(filterParams)
	if err != nil {
		slog.Error("Failed to fetch filtered recipes", "error", err)
		http.Error(w, "Failed to fetch filtered recipes", http.StatusInternalServerError)
		return
	}

	recipeIDs := make([]int, len(recipes))
	for i, r := range recipes {
		recipeIDs[i] = r.ID
	}
	tagsMap, _ := h.TagStore.GetForRecipes(recipeIDs)

	for i := range recipes {
		recipes[i].Tags = tagsMap[recipes[i].ID]
	}

	currentUser, err := auth.GetUserBySession(h.AuthStore, r)
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

	h.Renderer.RenderFragment(w, "recipe-cards", data)
}

func (h *Handler) LoadMoreRecipesHTMXHandler(w http.ResponseWriter, r *http.Request) {
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if offset < 0 {
		offset = 0
	}

	filterParams := models.FilterParams{
		Limit:  RecipesPerPage,
		Offset: offset,
	}
	recipes, err := h.RecipeStore.GetFiltered(filterParams)
	if err != nil {
		slog.Error("Failed to fetch recipes", "error", err)
		http.Error(w, "Failed to fetch recipes", http.StatusInternalServerError)
		return
	}

	recipeIDs := make([]int, len(recipes))
	for i, r := range recipes {
		recipeIDs[i] = r.ID
	}
	tagsMap, _ := h.TagStore.GetForRecipes(recipeIDs)

	for i := range recipes {
		recipes[i].Tags = tagsMap[recipes[i].ID]
	}

	currentUser, err := auth.GetUserBySession(h.AuthStore, r)
	isLoggedIn := err == nil

	if isLoggedIn {
		userTagsMap, _ := h.UserTagStore.GetForRecipes(currentUser.ID, recipeIDs)
		for i := range recipes {
			recipes[i].UserTags = userTagsMap[recipes[i].ID]
		}
	}

	hasMore := len(recipes) == RecipesPerPage
	nextOffset := offset + RecipesPerPage

	data := struct {
		Recipes     []models.Recipe
		IsLoggedIn  bool
		CurrentUser *auth.User
		NextOffset  int
		HasMore     bool
	}{
		Recipes:     recipes,
		IsLoggedIn:  isLoggedIn,
		CurrentUser: currentUser,
		NextOffset:  nextOffset,
		HasMore:     hasMore,
	}

	h.Renderer.RenderFragment(w, "recipe-cards-more", data)
}
