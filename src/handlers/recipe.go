package handlers

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/mr-flannery/go-recipe-book/src/auth"
	"github.com/mr-flannery/go-recipe-book/src/logging"
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
	ctx := r.Context()
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		logging.AddError(ctx, err, "Failed to parse multipart form")
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	user, err := auth.GetUserBySession(ctx, h.AuthStore, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	var prepTime, cookTime, calories int
	if prepTimeStr := r.FormValue("preptime"); prepTimeStr != "" {
		prepTime, err = strconv.Atoi(prepTimeStr)
		if err != nil {
			logging.AddError(ctx, err, "Invalid prep time")
			http.Error(w, "Invalid prep time", http.StatusBadRequest)
			return
		}
	}

	if cookTimeStr := r.FormValue("cooktime"); cookTimeStr != "" {
		cookTime, err = strconv.Atoi(cookTimeStr)
		if err != nil {
			logging.AddError(ctx, err, "Invalid cook time")
			http.Error(w, "Invalid cook time", http.StatusBadRequest)
			return
		}
	}

	if caloriesStr := r.FormValue("calories"); caloriesStr != "" {
		calories, err = strconv.Atoi(caloriesStr)
		if err != nil {
			logging.AddError(ctx, err, "Invalid calories")
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
			logging.AddError(ctx, err, "Failed to decode cropped image data")
			http.Error(w, "Failed to process cropped image", http.StatusBadRequest)
			return
		}
		imageData = decodedData
	} else {
		file, _, err := r.FormFile("image")
		if err == nil {
			defer file.Close()
			imageData, err = io.ReadAll(file)
			if err != nil {
				logging.AddError(ctx, err, "Failed to read image file")
				http.Error(w, "Failed to read image file", http.StatusInternalServerError)
				return
			}
		} else if err != http.ErrMissingFile {
			logging.AddError(ctx, err, "Error processing image file")
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

	recipeID, err := h.RecipeStore.Save(ctx, recipe)
	if err != nil {
		logging.AddError(ctx, err, "Failed to save recipe")
		http.Error(w, fmt.Sprintf("Failed to save recipe: %v", err), http.StatusInternalServerError)
		return
	}

	tagsStr := r.FormValue("tags")
	if tagsStr != "" {
		tagNames := strings.Split(tagsStr, ",")
		if err := h.TagStore.SetRecipeTags(ctx, recipeID, tagNames); err != nil {
			logging.AddError(ctx, err, "Failed to set recipe tags")
		}
	}

	logging.AddMany(ctx, map[string]any{
		"action":       "recipe.create",
		"recipe.id":    recipeID,
		"recipe.title": recipe.Title,
	})
	http.Redirect(w, r, fmt.Sprintf("/recipes/%d", recipeID), http.StatusSeeOther)
}

func (h *Handler) ListRecipesHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userInfo := auth.GetUserInfoFromContext(ctx)
	var currentUser *auth.User
	pageSize := models.DefaultPageSize
	viewMode := models.DefaultViewMode

	if userInfo.IsLoggedIn {
		currentUser, _ = auth.GetUserBySession(ctx, h.AuthStore, r)
		if prefs, err := h.UserPreferencesStore.Get(ctx, currentUser.ID); err == nil {
			pageSize = prefs.PageSize
			if prefs.ViewMode != "" {
				viewMode = prefs.ViewMode
			}
		}
	}

	query := r.URL.Query()

	if ps := query.Get("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	currentPage := 1
	if p := query.Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			currentPage = parsed
		}
	}

	filterState := FilterState{
		Page:         currentPage,
		PageSize:     pageSize,
		Search:       strings.TrimSpace(query.Get("search")),
		Tags:         query.Get("tags"),
		UserTags:     query.Get("user_tags"),
		AuthoredByMe: query.Get("authored_by_me") == "1",
	}

	if calOp := query.Get("calories_op"); calOp != "" {
		filterState.CaloriesOp = calOp
		if calVal, err := strconv.Atoi(query.Get("calories_value")); err == nil && calVal > 0 {
			filterState.CaloriesValue = calVal
		}
	}
	if prepOp := query.Get("prep_time_op"); prepOp != "" {
		filterState.PrepTimeOp = prepOp
		if prepVal, err := strconv.Atoi(query.Get("prep_time_value")); err == nil && prepVal > 0 {
			filterState.PrepTimeValue = prepVal
		}
	}
	if cookOp := query.Get("cook_time_op"); cookOp != "" {
		filterState.CookTimeOp = cookOp
		if cookVal, err := strconv.Atoi(query.Get("cook_time_value")); err == nil && cookVal > 0 {
			filterState.CookTimeValue = cookVal
		}
	}

	offset := (currentPage - 1) * pageSize
	limit := pageSize

	filterParams := models.FilterParams{
		Search:        filterState.Search,
		Limit:         limit,
		Offset:        offset,
		CaloriesOp:    filterState.CaloriesOp,
		CaloriesValue: filterState.CaloriesValue,
		PrepTimeOp:    filterState.PrepTimeOp,
		PrepTimeValue: filterState.PrepTimeValue,
		CookTimeOp:    filterState.CookTimeOp,
		CookTimeValue: filterState.CookTimeValue,
	}

	if filterState.Tags != "" {
		for _, tag := range strings.Split(filterState.Tags, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				filterParams.Tags = append(filterParams.Tags, tag)
			}
		}
	}

	if filterState.AuthoredByMe && userInfo.IsLoggedIn && currentUser != nil {
		filterParams.AuthorID = currentUser.ID
	}

	if filterState.UserTags != "" && userInfo.IsLoggedIn && currentUser != nil {
		filterParams.UserID = currentUser.ID
		for _, tag := range strings.Split(filterState.UserTags, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				filterParams.UserTags = append(filterParams.UserTags, tag)
			}
		}
	}

	recipes, err := h.RecipeStore.GetFiltered(ctx, filterParams)
	if err != nil {
		h.Renderer.RenderError(w, r, http.StatusInternalServerError, "Failed to fetch recipes. Please try again later.")
		return
	}

	countParams := filterParams
	countParams.Limit = 0
	countParams.Offset = 0
	totalCount, _ := h.RecipeStore.CountFiltered(ctx, countParams)

	recipeIDs := make([]int, len(recipes))
	for i, rec := range recipes {
		recipeIDs[i] = rec.ID
	}
	tagsMap, _ := h.TagStore.GetForRecipes(ctx, recipeIDs)

	for i := range recipes {
		recipes[i].Tags = tagsMap[recipes[i].ID]
	}

	if userInfo.IsLoggedIn && currentUser != nil {
		userTagsMap, _ := h.UserTagStore.GetForRecipes(ctx, currentUser.ID, recipeIDs)
		for i := range recipes {
			recipes[i].UserTags = userTagsMap[recipes[i].ID]
		}
	}

	pagination := CalculatePagination(totalCount, currentPage, pageSize)

	logging.AddMany(ctx, map[string]any{
		"action":       "recipe.list",
		"result.count": len(recipes),
		"result.total": totalCount,
		"filter.page":  currentPage,
	})

	data := struct {
		Recipes     []models.Recipe
		UserInfo    *auth.UserInfo
		IsLoggedIn  bool
		CurrentUser *auth.User
		ViewMode    string
		FilterState FilterState
		PaginationData
	}{
		Recipes:        recipes,
		UserInfo:       userInfo,
		IsLoggedIn:     userInfo.IsLoggedIn,
		CurrentUser:    currentUser,
		ViewMode:       viewMode,
		FilterState:    filterState,
		PaginationData: pagination,
	}

	h.Renderer.RenderPage(w, "list.gohtml", data)
}

func (h *Handler) GetUpdateRecipeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	recipeID := r.PathValue("id")
	recipe, err := h.RecipeStore.GetByID(ctx, recipeID)
	if err != nil {
		h.Renderer.RenderError(w, r, http.StatusNotFound, "The recipe you're looking for doesn't exist or has been removed.")
		return
	}

	currentUser, err := auth.GetUserBySession(ctx, h.AuthStore, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	if currentUser.ID != recipe.AuthorID {
		h.Renderer.RenderError(w, r, http.StatusForbidden, "You can only edit your own recipes.")
		return
	}

	recipeIDInt, _ := strconv.Atoi(recipeID)
	recipe.Tags, _ = h.TagStore.GetByRecipeID(ctx, recipeIDInt)

	data := struct {
		Recipe   models.Recipe
		UserInfo *auth.UserInfo
	}{
		Recipe:   recipe,
		UserInfo: auth.GetUserInfoFromContext(ctx),
	}

	h.Renderer.RenderPage(w, "update.gohtml", data)
}

func (h *Handler) PostUpdateRecipeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	recipeID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		logging.AddError(ctx, err, "Invalid recipe ID")
		http.Error(w, "Invalid recipe ID", http.StatusBadRequest)
		return
	}

	err = r.ParseMultipartForm(10 << 20)
	if err != nil {
		logging.AddError(ctx, err, "Failed to parse multipart form")
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	user, err := auth.GetUserBySession(ctx, h.AuthStore, r)
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	existingRecipe, err := h.RecipeStore.GetByID(ctx, strconv.Itoa(recipeID))
	if err != nil {
		logging.AddError(ctx, err, "Recipe not found")
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
			logging.AddError(ctx, err, "Invalid prep time")
			http.Error(w, "Invalid prep time", http.StatusBadRequest)
			return
		}
	}

	if cookTimeStr := r.FormValue("cooktime"); cookTimeStr != "" {
		cookTime, err = strconv.Atoi(cookTimeStr)
		if err != nil {
			logging.AddError(ctx, err, "Invalid cook time")
			http.Error(w, "Invalid cook time", http.StatusBadRequest)
			return
		}
	}

	if caloriesStr := r.FormValue("calories"); caloriesStr != "" {
		calories, err = strconv.Atoi(caloriesStr)
		if err != nil {
			logging.AddError(ctx, err, "Invalid calories")
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
			logging.AddError(ctx, err, "Failed to read image file")
			http.Error(w, "Failed to read image file", http.StatusInternalServerError)
			return
		}
	} else if err != http.ErrMissingFile {
		logging.AddError(ctx, err, "Error processing image file")
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

	if err := h.RecipeStore.Update(ctx, updatedRecipe); err != nil {
		logging.AddError(ctx, err, "Failed to update recipe")
		http.Error(w, "Failed to update recipe", http.StatusInternalServerError)
		return
	}

	tagsStr := r.FormValue("tags")
	if tagsStr != "" {
		tagNames := strings.Split(tagsStr, ",")
		if err := h.TagStore.SetRecipeTags(ctx, recipeID, tagNames); err != nil {
			logging.AddError(ctx, err, "Failed to set recipe tags")
		}
	} else {
		if err := h.TagStore.SetRecipeTags(ctx, recipeID, []string{}); err != nil {
			logging.AddError(ctx, err, "Failed to clear recipe tags")
		}
	}

	logging.AddMany(ctx, map[string]any{
		"action":       "recipe.update",
		"recipe.id":    recipeID,
		"recipe.title": updatedRecipe.Title,
	})
	http.Redirect(w, r, fmt.Sprintf("/recipes/%d", recipeID), http.StatusSeeOther)
}

func (h *Handler) ViewRecipeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	recipeID := r.PathValue("id")
	if recipeID == "" {
		h.Renderer.RenderError(w, r, http.StatusBadRequest, "No recipe specified.")
		return
	}

	userInfo := auth.GetUserInfoFromContext(ctx)
	currentUser, err := auth.GetUserBySession(ctx, h.AuthStore, r)
	isLoggedIn := err == nil

	var userIDPtr *int
	if isLoggedIn {
		userIDPtr = &currentUser.ID
	}

	recipeDetails, err := h.RecipeStore.GetRecipeWithDetails(ctx, recipeID, userIDPtr)
	if err != nil {
		h.Renderer.RenderError(w, r, http.StatusNotFound, "The recipe you're looking for doesn't exist or has been removed.")
		return
	}

	recipeIDInt, _ := strconv.Atoi(recipeID)
	logging.AddMany(ctx, map[string]any{
		"action":       "recipe.view",
		"recipe.id":    recipeIDInt,
		"recipe.title": recipeDetails.Recipe.Title,
	})

	isRecipeAuthor := isLoggedIn && currentUser.ID == recipeDetails.Recipe.AuthorID

	commentsWithUsernames := make([]CommentTemplateData, len(recipeDetails.Comments))
	for i, c := range recipeDetails.Comments {
		isCommentAuthor := isLoggedIn && currentUser.ID == c.AuthorID
		commentsWithUsernames[i] = CommentTemplateData{
			Comment:  c.Comment,
			Username: c.AuthorUsername,
			IsAuthor: isCommentAuthor,
		}
	}

	data := struct {
		Recipe      models.Recipe
		UserTags    []models.UserTag
		Comments    []CommentTemplateData
		IsLoggedIn  bool
		CurrentUser *auth.User
		IsAuthor    bool
		UserInfo    *auth.UserInfo
	}{
		Recipe:      recipeDetails.Recipe,
		UserTags:    recipeDetails.Recipe.UserTags,
		Comments:    commentsWithUsernames,
		IsLoggedIn:  isLoggedIn,
		CurrentUser: currentUser,
		IsAuthor:    isRecipeAuthor,
		UserInfo:    userInfo,
	}

	h.Renderer.RenderPage(w, "view.gohtml", data)
}

func (h *Handler) CommentHTMXHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	recipeID := r.PathValue("id")
	if recipeID == "" {
		http.Error(w, "Recipe ID is required", http.StatusBadRequest)
		return
	}

	user, err := auth.GetUserBySession(ctx, h.AuthStore, r)
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

	if err := h.CommentStore.Save(ctx, comment); err != nil {
		http.Error(w, fmt.Sprintf("Failed to save comment: %v", err), http.StatusInternalServerError)
		return
	}

	savedComment, err := h.CommentStore.GetLatestByUserAndRecipe(ctx, user.ID, recipeIDInt)
	if err != nil {
		http.Error(w, "Failed to retrieve saved comment", http.StatusInternalServerError)
		return
	}

	logging.AddMany(ctx, map[string]any{
		"action":     "comment.create",
		"recipe.id":  recipeIDInt,
		"comment.id": savedComment.ID,
	})

	commentData := CommentTemplateData{
		Comment:  savedComment,
		Username: user.Username,
		IsAuthor: true,
	}

	h.Renderer.RenderFragment(w, "comment.gohtml", commentData)
}

type CommentTemplateData struct {
	models.Comment
	Username string
	IsAuthor bool
}

func (h *Handler) UpdateCommentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commentIDStr := r.PathValue("id")
	if commentIDStr == "" {
		http.Error(w, "Comment ID is required", http.StatusBadRequest)
		return
	}

	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	user, err := auth.GetUserBySession(ctx, h.AuthStore, r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	comment, err := h.CommentStore.GetByID(ctx, commentID)
	if err != nil {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	if comment.AuthorID != user.ID {
		http.Error(w, "Forbidden: You can only edit your own comments", http.StatusForbidden)
		return
	}

	r.ParseForm()
	newContent := r.FormValue("comment")
	if newContent == "" {
		http.Error(w, "Comment content is required", http.StatusBadRequest)
		return
	}

	if err := h.CommentStore.Update(ctx, commentID, newContent); err != nil {
		http.Error(w, "Failed to update comment", http.StatusInternalServerError)
		return
	}

	updatedComment, err := h.CommentStore.GetByID(ctx, commentID)
	if err != nil {
		http.Error(w, "Failed to retrieve updated comment", http.StatusInternalServerError)
		return
	}

	logging.AddMany(ctx, map[string]any{
		"action":     "comment.update",
		"comment.id": commentID,
	})

	commentData := CommentTemplateData{
		Comment:  updatedComment,
		Username: user.Username,
		IsAuthor: true,
	}

	h.Renderer.RenderFragment(w, "comment.gohtml", commentData)
}

func (h *Handler) DeleteCommentHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	commentIDStr := r.PathValue("id")
	if commentIDStr == "" {
		http.Error(w, "Comment ID is required", http.StatusBadRequest)
		return
	}

	commentID, err := strconv.Atoi(commentIDStr)
	if err != nil {
		http.Error(w, "Invalid comment ID", http.StatusBadRequest)
		return
	}

	user, err := auth.GetUserBySession(ctx, h.AuthStore, r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	comment, err := h.CommentStore.GetByID(ctx, commentID)
	if err != nil {
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	if comment.AuthorID != user.ID {
		http.Error(w, "Forbidden: You can only delete your own comments", http.StatusForbidden)
		return
	}

	if err := h.CommentStore.Delete(ctx, commentID); err != nil {
		http.Error(w, "Failed to delete comment", http.StatusInternalServerError)
		return
	}

	logging.AddMany(ctx, map[string]any{
		"action":     "comment.delete",
		"comment.id": commentID,
	})

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) DeleteRecipeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	recipeID := r.PathValue("id")
	if recipeID == "" {
		http.Error(w, "Recipe ID is required", http.StatusBadRequest)
		return
	}

	currentUser, err := auth.GetUserBySession(ctx, h.AuthStore, r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	recipe, err := h.RecipeStore.GetByID(ctx, recipeID)
	if err != nil {
		http.Error(w, "Recipe not found", http.StatusNotFound)
		return
	}

	if currentUser.ID != recipe.AuthorID {
		http.Error(w, "Forbidden: You can only delete your own recipes", http.StatusForbidden)
		return
	}

	if err := h.RecipeStore.Delete(ctx, recipeID); err != nil {
		http.Error(w, "Failed to delete recipe", http.StatusInternalServerError)
		return
	}

	logging.AddMany(ctx, map[string]any{
		"action":       "recipe.delete",
		"recipe.id":    recipeID,
		"recipe.title": recipe.Title,
	})

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Recipe deleted successfully"))
}

func (h *Handler) RandomRecipeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	recipeID, err := h.RecipeStore.GetRandomID(ctx)
	if err != nil {
		logging.AddError(ctx, err, "Failed to get random recipe")
		http.Redirect(w, r, "/recipes", http.StatusSeeOther)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/recipes/%d", recipeID), http.StatusSeeOther)
}

func (h *Handler) FilterRecipesHTMXHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	r.ParseForm()

	currentUser, err := auth.GetUserBySession(ctx, h.AuthStore, r)
	isLoggedIn := err == nil

	pageSize := models.DefaultPageSize
	viewMode := models.DefaultViewMode
	if pageSizeStr := r.FormValue("page_size"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 && ps <= 100 {
			pageSize = ps
		}
	} else if isLoggedIn {
		if prefs, err := h.UserPreferencesStore.Get(ctx, currentUser.ID); err == nil {
			pageSize = prefs.PageSize
		}
	}

	if vm := r.FormValue("view_mode"); vm == models.ViewModeGrid || vm == models.ViewModeList {
		viewMode = vm
		if isLoggedIn {
			h.UserPreferencesStore.SetViewMode(ctx, currentUser.ID, vm)
		}
	} else if isLoggedIn {
		if prefs, err := h.UserPreferencesStore.Get(ctx, currentUser.ID); err == nil && prefs.ViewMode != "" {
			viewMode = prefs.ViewMode
		}
	}

	currentPage := 1
	if p := r.FormValue("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			currentPage = parsed
		}
	}

	offset := (currentPage - 1) * pageSize
	limit := pageSize

	filterState := FilterState{
		Page:         currentPage,
		PageSize:     pageSize,
		Search:       strings.TrimSpace(r.FormValue("search")),
		Tags:         r.FormValue("tags"),
		UserTags:     r.FormValue("user_tags"),
		AuthoredByMe: r.FormValue("authored_by_me") == "1",
	}

	filterParams := models.FilterParams{
		Search: filterState.Search,
		Limit:  limit,
		Offset: offset,
	}

	if caloriesStr := r.FormValue("calories_value"); caloriesStr != "" {
		if calories, err := strconv.Atoi(caloriesStr); err == nil && calories > 0 {
			filterParams.CaloriesValue = calories
			filterParams.CaloriesOp = r.FormValue("calories_op")
			filterState.CaloriesValue = calories
			filterState.CaloriesOp = filterParams.CaloriesOp
		}
	}

	if prepTimeStr := r.FormValue("prep_time_value"); prepTimeStr != "" {
		if prepTime, err := strconv.Atoi(prepTimeStr); err == nil && prepTime > 0 {
			filterParams.PrepTimeValue = prepTime
			filterParams.PrepTimeOp = r.FormValue("prep_time_op")
			filterState.PrepTimeValue = prepTime
			filterState.PrepTimeOp = filterParams.PrepTimeOp
		}
	}

	if cookTimeStr := r.FormValue("cook_time_value"); cookTimeStr != "" {
		if cookTime, err := strconv.Atoi(cookTimeStr); err == nil && cookTime > 0 {
			filterParams.CookTimeValue = cookTime
			filterParams.CookTimeOp = r.FormValue("cook_time_op")
			filterState.CookTimeValue = cookTime
			filterState.CookTimeOp = filterParams.CookTimeOp
		}
	}

	if filterState.Tags != "" {
		tags := strings.Split(filterState.Tags, ",")
		for _, tag := range tags {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				filterParams.Tags = append(filterParams.Tags, tag)
			}
		}
	}

	if filterState.AuthoredByMe && isLoggedIn {
		filterParams.AuthorID = currentUser.ID
	}

	if filterState.UserTags != "" && isLoggedIn {
		filterParams.UserID = currentUser.ID
		tags := strings.Split(filterState.UserTags, ",")
		for _, tag := range tags {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				filterParams.UserTags = append(filterParams.UserTags, tag)
			}
		}
	}

	recipes, err := h.RecipeStore.GetFiltered(ctx, filterParams)
	if err != nil {
		logging.AddError(ctx, err, "Failed to fetch filtered recipes")
		http.Error(w, "Failed to fetch filtered recipes", http.StatusInternalServerError)
		return
	}

	countParams := filterParams
	countParams.Limit = 0
	countParams.Offset = 0
	totalCount, _ := h.RecipeStore.CountFiltered(ctx, countParams)

	recipeIDs := make([]int, len(recipes))
	for i, rec := range recipes {
		recipeIDs[i] = rec.ID
	}
	tagsMap, _ := h.TagStore.GetForRecipes(ctx, recipeIDs)

	for i := range recipes {
		recipes[i].Tags = tagsMap[recipes[i].ID]
	}

	if isLoggedIn {
		userTagsMap, _ := h.UserTagStore.GetForRecipes(ctx, currentUser.ID, recipeIDs)
		for i := range recipes {
			recipes[i].UserTags = userTagsMap[recipes[i].ID]
		}
	}

	pagination := CalculatePagination(totalCount, currentPage, pageSize)

	w.Header().Set("HX-Replace-Url", filterState.ToURLQuery())

	data := struct {
		Recipes     []models.Recipe
		IsLoggedIn  bool
		CurrentUser *auth.User
		ViewMode    string
		PaginationData
	}{
		Recipes:        recipes,
		IsLoggedIn:     isLoggedIn,
		CurrentUser:    currentUser,
		ViewMode:       viewMode,
		PaginationData: pagination,
	}

	h.Renderer.RenderFragment(w, "recipe-cards", data)
}

func (h *Handler) SetPageSizeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUser, err := auth.GetUserBySession(ctx, h.AuthStore, r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	r.ParseForm()
	pageSizeStr := r.FormValue("page_size")
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil || pageSize < 10 || pageSize > 100 {
		http.Error(w, "Invalid page size", http.StatusBadRequest)
		return
	}

	if err := h.UserPreferencesStore.SetPageSize(ctx, currentUser.ID, pageSize); err != nil {
		logging.AddError(ctx, err, "Failed to save page size preference")
		http.Error(w, "Failed to save preference", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) SetViewModeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUser, err := auth.GetUserBySession(ctx, h.AuthStore, r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	r.ParseForm()
	viewMode := r.FormValue("view_mode")
	if viewMode != models.ViewModeGrid && viewMode != models.ViewModeList {
		http.Error(w, "Invalid view mode", http.StatusBadRequest)
		return
	}

	if err := h.UserPreferencesStore.SetViewMode(ctx, currentUser.ID, viewMode); err != nil {
		logging.AddError(ctx, err, "Failed to save view mode preference")
		http.Error(w, "Failed to save preference", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
