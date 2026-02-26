package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/mr-flannery/go-recipe-book/src/auth"
	"github.com/mr-flannery/go-recipe-book/src/models"
)

type AccountSettingsData struct {
	UserInfo *auth.UserInfo
	Success  string
	Error    string
}

func (h *Handler) GetAccountSettingsHandler(w http.ResponseWriter, r *http.Request) {
	data := AccountSettingsData{
		UserInfo: auth.GetUserInfoFromContext(r.Context()),
		Success:  r.URL.Query().Get("success"),
		Error:    r.URL.Query().Get("error"),
	}
	h.Renderer.RenderPage(w, "account-settings.gohtml", data)
}

type UserDataExport struct {
	ExportedAt  string                  `json:"exported_at"`
	Account     UserDataAccount         `json:"account"`
	Preferences *models.UserPreferences `json:"preferences,omitempty"`
	Recipes     []UserDataRecipe        `json:"recipes"`
	Comments    []UserDataComment       `json:"comments"`
	UserTags    []UserDataUserTag       `json:"user_tags"`
}

type UserDataAccount struct {
	Username  string  `json:"username"`
	Email     string  `json:"email"`
	CreatedAt string  `json:"created_at"`
	LastLogin *string `json:"last_login,omitempty"`
}

type UserDataRecipe struct {
	ID           int      `json:"id"`
	Title        string   `json:"title"`
	Ingredients  string   `json:"ingredients"`
	Instructions string   `json:"instructions"`
	PrepTime     int      `json:"prep_time_minutes"`
	CookTime     int      `json:"cook_time_minutes"`
	Calories     int      `json:"calories"`
	Tags         []string `json:"tags"`
	CreatedAt    string   `json:"created_at"`
	UpdatedAt    string   `json:"updated_at"`
}

type UserDataComment struct {
	RecipeID  int    `json:"recipe_id"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type UserDataUserTag struct {
	RecipeID int    `json:"recipe_id"`
	Name     string `json:"name"`
}

func (h *Handler) ExportUserDataHandler(w http.ResponseWriter, r *http.Request) {
	userInfo := auth.GetUserInfoFromContext(r.Context())
	if !userInfo.IsLoggedIn {
		h.Renderer.RenderError(w, r, http.StatusUnauthorized, "You must be logged in to export your data.")
		return
	}

	user, err := h.AuthStore.GetFullUserByID(userInfo.UserID)
	if err != nil {
		slog.Error("Failed to get user for export", "error", err, "user_id", userInfo.UserID)
		h.Renderer.RenderError(w, r, http.StatusInternalServerError, "Failed to export data. Please try again.")
		return
	}

	export := UserDataExport{
		ExportedAt: time.Now().UTC().Format(time.RFC3339),
		Account: UserDataAccount{
			Username:  user.Username,
			Email:     user.Email,
			CreatedAt: user.CreatedAt.Format(time.RFC3339),
		},
		Recipes:  []UserDataRecipe{},
		Comments: []UserDataComment{},
		UserTags: []UserDataUserTag{},
	}

	if user.LastLogin != nil {
		formatted := user.LastLogin.Format(time.RFC3339)
		export.Account.LastLogin = &formatted
	}

	prefs, err := h.UserPreferencesStore.Get(userInfo.UserID)
	if err == nil && prefs != nil {
		export.Preferences = prefs
	}

	recipes, err := h.RecipeStore.GetFiltered(models.FilterParams{AuthorID: userInfo.UserID, Limit: 10000})
	if err == nil {
		for _, recipe := range recipes {
			tags := make([]string, len(recipe.Tags))
			for i, t := range recipe.Tags {
				tags[i] = t.Name
			}
			export.Recipes = append(export.Recipes, UserDataRecipe{
				ID:           recipe.ID,
				Title:        recipe.Title,
				Ingredients:  recipe.IngredientsMD,
				Instructions: recipe.InstructionsMD,
				PrepTime:     recipe.PrepTime,
				CookTime:     recipe.CookTime,
				Calories:     recipe.Calories,
				Tags:         tags,
				CreatedAt:    recipe.CreatedAt.Format(time.RFC3339),
				UpdatedAt:    recipe.UpdatedAt.Format(time.RFC3339),
			})
		}
	}

	comments, err := h.CommentStore.GetByUserID(userInfo.UserID)
	if err == nil {
		for _, comment := range comments {
			export.Comments = append(export.Comments, UserDataComment{
				RecipeID:  comment.RecipeID,
				Content:   comment.ContentMD,
				CreatedAt: comment.CreatedAt.Format(time.RFC3339),
				UpdatedAt: comment.UpdatedAt.Format(time.RFC3339),
			})
		}
	}

	userTags, err := h.UserTagStore.GetByUserID(userInfo.UserID)
	if err == nil {
		for _, tag := range userTags {
			export.UserTags = append(export.UserTags, UserDataUserTag{
				RecipeID: tag.RecipeID,
				Name:     tag.Name,
			})
		}
	}

	jsonData, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		slog.Error("Failed to marshal export data", "error", err, "user_id", userInfo.UserID)
		h.Renderer.RenderError(w, r, http.StatusInternalServerError, "Failed to export data. Please try again.")
		return
	}

	filename := "recipe-book-data-" + time.Now().Format("2006-01-02") + ".json"
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.Write(jsonData)

	slog.Info("User data exported", "user_id", userInfo.UserID, "username", user.Username)
}

func (h *Handler) DeleteOwnAccountHandler(w http.ResponseWriter, r *http.Request) {
	userInfo := auth.GetUserInfoFromContext(r.Context())
	if !userInfo.IsLoggedIn {
		h.Renderer.RenderError(w, r, http.StatusUnauthorized, "You must be logged in to delete your account.")
		return
	}

	if userInfo.IsAdmin {
		http.Redirect(w, r, "/account?error=Admin accounts cannot be deleted through self-service. Please contact another administrator.", http.StatusSeeOther)
		return
	}

	r.ParseForm()
	password := r.FormValue("password")
	confirmDelete := r.FormValue("confirm_delete")

	if confirmDelete != "DELETE" {
		http.Redirect(w, r, "/account?error=Please type DELETE to confirm account deletion.", http.StatusSeeOther)
		return
	}

	user, err := h.AuthStore.GetUserByID(userInfo.UserID)
	if err != nil {
		slog.Error("Failed to get user for deletion", "error", err, "user_id", userInfo.UserID)
		http.Redirect(w, r, "/account?error=Failed to verify account. Please try again.", http.StatusSeeOther)
		return
	}

	_, err = auth.Authenticate(h.AuthStore, user.Email, password)
	if err != nil {
		http.Redirect(w, r, "/account?error=Incorrect password. Please try again.", http.StatusSeeOther)
		return
	}

	err = auth.DeleteUser(h.AuthStore, userInfo.UserID)
	if err != nil {
		slog.Error("Failed to delete user account", "error", err, "user_id", userInfo.UserID)
		http.Redirect(w, r, "/account?error=Failed to delete account. Please try again.", http.StatusSeeOther)
		return
	}

	auth.ClearSessionCookie(w)

	slog.Info("User self-deleted account", "user_id", userInfo.UserID, "username", user.Username)

	http.Redirect(w, r, "/?account_deleted=true", http.StatusSeeOther)
}
