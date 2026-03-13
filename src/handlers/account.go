package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/mr-flannery/go-recipe-book/src/auth"
	"github.com/mr-flannery/go-recipe-book/src/logging"
	"github.com/mr-flannery/go-recipe-book/src/models"
	"github.com/mr-flannery/go-recipe-book/src/store"
	"github.com/mr-flannery/go-recipe-book/src/utils"
)

type APIKeyDisplay struct {
	store.APIKey
	DecryptedKey string
}

type AccountSettingsData struct {
	UserInfo *auth.UserInfo
	APIKeys  []APIKeyDisplay
	Success  string
	Error    string
}

func (h *Handler) GetAccountSettingsHandler(w http.ResponseWriter, r *http.Request) {
	userInfo := auth.GetUserInfoFromContext(r.Context())
	data := AccountSettingsData{
		UserInfo: userInfo,
	}
	h.Renderer.RenderPage(w, "account-settings.gohtml", data)
}

func (h *Handler) GetAccountAPIKeysHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userInfo := auth.GetUserInfoFromContext(ctx)

	apiKeys, err := h.APIKeyStore.GetByUserID(ctx, userInfo.UserID)
	if err != nil {
		logging.AddError(ctx, err, "Failed to fetch API keys")
		apiKeys = []store.APIKey{}
	}

	displayKeys := make([]APIKeyDisplay, len(apiKeys))
	for i, key := range apiKeys {
		displayKeys[i] = APIKeyDisplay{APIKey: key}
		if key.EncryptedKey != "" && len(h.APIEncryptionKey) > 0 {
			decrypted, err := utils.Decrypt(key.EncryptedKey, h.APIEncryptionKey)
			if err != nil {
				logging.AddError(ctx, err, "Failed to decrypt API key")
			} else {
				displayKeys[i].DecryptedKey = decrypted
			}
		}
	}

	data := AccountSettingsData{
		UserInfo: userInfo,
		APIKeys:  displayKeys,
		Success:  r.URL.Query().Get("success"),
		Error:    r.URL.Query().Get("error"),
	}
	h.Renderer.RenderPage(w, "account-api-keys.gohtml", data)
}

func (h *Handler) GetAccountExportHandler(w http.ResponseWriter, r *http.Request) {
	userInfo := auth.GetUserInfoFromContext(r.Context())
	data := AccountSettingsData{
		UserInfo: userInfo,
	}
	h.Renderer.RenderPage(w, "account-export.gohtml", data)
}

func (h *Handler) GetAccountDeleteHandler(w http.ResponseWriter, r *http.Request) {
	userInfo := auth.GetUserInfoFromContext(r.Context())
	data := AccountSettingsData{
		UserInfo: userInfo,
		Error:    r.URL.Query().Get("error"),
	}
	h.Renderer.RenderPage(w, "account-delete.gohtml", data)
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
	ctx := r.Context()
	userInfo := auth.GetUserInfoFromContext(ctx)
	if !userInfo.IsLoggedIn {
		h.Renderer.RenderError(w, r, http.StatusUnauthorized, "You must be logged in to export your data.")
		return
	}

	user, err := h.AuthStore.GetFullUserByID(ctx, userInfo.UserID)
	if err != nil {
		logging.AddError(ctx, err, "Failed to get user for export")
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

	prefs, err := h.UserPreferencesStore.Get(ctx, userInfo.UserID)
	if err == nil && prefs != nil {
		export.Preferences = prefs
	}

	recipes, err := h.RecipeStore.GetFiltered(ctx, models.FilterParams{AuthorID: userInfo.UserID, Limit: 10000})
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

	comments, err := h.CommentStore.GetByUserID(ctx, userInfo.UserID)
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

	userTags, err := h.UserTagStore.GetByUserID(ctx, userInfo.UserID)
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
		logging.AddError(ctx, err, "Failed to marshal export data")
		h.Renderer.RenderError(w, r, http.StatusInternalServerError, "Failed to export data. Please try again.")
		return
	}

	filename := "recipe-book-data-" + time.Now().Format("2006-01-02") + ".json"
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.Write(jsonData)

	logging.AddMany(ctx, map[string]any{
		"action":               "account.export",
		"export.recipe_count":  len(export.Recipes),
		"export.comment_count": len(export.Comments),
	})
}

func (h *Handler) DeleteOwnAccountHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userInfo := auth.GetUserInfoFromContext(ctx)
	if !userInfo.IsLoggedIn {
		h.Renderer.RenderError(w, r, http.StatusUnauthorized, "You must be logged in to delete your account.")
		return
	}

	if userInfo.IsAdmin {
		http.Redirect(w, r, "/account/delete?error=Admin accounts cannot be deleted through self-service. Please contact another administrator.", http.StatusSeeOther)
		return
	}

	r.ParseForm()
	password := r.FormValue("password")
	confirmDelete := r.FormValue("confirm_delete")

	if confirmDelete != "DELETE" {
		http.Redirect(w, r, "/account/delete?error=Please type DELETE to confirm account deletion.", http.StatusSeeOther)
		return
	}

	user, err := h.AuthStore.GetUserByID(ctx, userInfo.UserID)
	if err != nil {
		logging.AddError(ctx, err, "Failed to get user for deletion")
		http.Redirect(w, r, "/account/delete?error=Failed to verify account. Please try again.", http.StatusSeeOther)
		return
	}

	_, err = auth.Authenticate(ctx, h.AuthStore, user.Email, password)
	if err != nil {
		http.Redirect(w, r, "/account/delete?error=Incorrect password. Please try again.", http.StatusSeeOther)
		return
	}

	err = auth.DeleteUser(ctx, h.AuthStore, userInfo.UserID)
	if err != nil {
		logging.AddError(ctx, err, "Failed to delete user account")
		http.Redirect(w, r, "/account/delete?error=Failed to delete account. Please try again.", http.StatusSeeOther)
		return
	}

	auth.ClearSessionCookie(w)

	logging.AddMany(ctx, map[string]any{
		"action":           "account.delete_self",
		"deleted.user_id":  userInfo.UserID,
		"deleted.username": user.Username,
	})

	http.Redirect(w, r, "/?account_deleted=true", http.StatusSeeOther)
}

func generateAPIKey() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "rb_" + hex.EncodeToString(bytes), nil
}

func (h *Handler) CreateAPIKeyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userInfo := auth.GetUserInfoFromContext(ctx)
	if !userInfo.IsLoggedIn {
		h.Renderer.RenderError(w, r, http.StatusUnauthorized, "You must be logged in to create API keys.")
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/account/api-keys?error=Invalid form data.", http.StatusSeeOther)
		return
	}

	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		http.Redirect(w, r, "/account/api-keys?error=API key name is required.", http.StatusSeeOther)
		return
	}
	if len(name) > 100 {
		http.Redirect(w, r, "/account/api-keys?error=API key name must be 100 characters or less.", http.StatusSeeOther)
		return
	}

	rawKey, err := generateAPIKey()
	if err != nil {
		logging.AddError(ctx, err, "Failed to generate API key")
		http.Redirect(w, r, "/account/api-keys?error=Failed to generate API key.", http.StatusSeeOther)
		return
	}

	keyHash := auth.HashAPIKey(rawKey)
	keyPrefix := rawKey[:10]

	var encryptedKey string
	if len(h.APIEncryptionKey) > 0 {
		encryptedKey, err = utils.Encrypt(rawKey, h.APIEncryptionKey)
		if err != nil {
			logging.AddError(ctx, err, "Failed to encrypt API key")
			http.Redirect(w, r, "/account/api-keys?error=Failed to create API key.", http.StatusSeeOther)
			return
		}
	}

	_, err = h.APIKeyStore.Create(ctx, userInfo.UserID, name, keyHash, keyPrefix, encryptedKey)
	if err != nil {
		logging.AddError(ctx, err, "Failed to save API key")
		http.Redirect(w, r, "/account/api-keys?error=Failed to create API key.", http.StatusSeeOther)
		return
	}

	logging.AddMany(ctx, map[string]any{
		"action":         "account.api_key.create",
		"api_key.name":   name,
		"api_key.prefix": keyPrefix,
	})

	http.Redirect(w, r, "/account/api-keys?success="+rawKey, http.StatusSeeOther)
}

func (h *Handler) DeleteAPIKeyHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userInfo := auth.GetUserInfoFromContext(ctx)
	if !userInfo.IsLoggedIn {
		h.Renderer.RenderError(w, r, http.StatusUnauthorized, "You must be logged in to delete API keys.")
		return
	}

	keyIDStr := r.PathValue("id")
	logging.AddMany(ctx, map[string]any{
		"debug.path":       r.URL.Path,
		"debug.key_id_str": keyIDStr,
		"debug.method":     r.Method,
	})
	keyID, err := strconv.Atoi(keyIDStr)
	if err != nil {
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`<tr><td colspan="5" style="color: red; padding: 16px;">Invalid API key ID</td></tr>`))
			return
		}
		http.Redirect(w, r, "/account/api-keys?error=Invalid API key ID.", http.StatusSeeOther)
		return
	}

	err = h.APIKeyStore.Delete(ctx, userInfo.UserID, keyID)
	if err != nil {
		logging.AddError(ctx, err, "Failed to delete API key")
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("Content-Type", "text/html")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`<tr><td colspan="5" style="color: red; padding: 16px;">Failed to delete API key</td></tr>`))
			return
		}
		http.Redirect(w, r, "/account/api-keys?error=Failed to delete API key.", http.StatusSeeOther)
		return
	}

	logging.AddMany(ctx, map[string]any{
		"action":     "account.api_key.delete",
		"api_key.id": keyID,
	})

	if r.Header.Get("HX-Request") == "true" {
		w.WriteHeader(http.StatusOK)
		return
	}

	http.Redirect(w, r, "/account/api-keys?success=API key deleted.", http.StatusSeeOther)
}

type ThemeOption struct {
	ID          string
	Name        string
	Description string
}

type ThemeSettingsData struct {
	UserInfo     *auth.UserInfo
	CurrentTheme string
	Themes       []ThemeOption
	Success      string
	Error        string
}

var AvailableThemes = []ThemeOption{
	{ID: models.ThemeEditorial, Name: "Editorial", Description: "Clean, magazine-inspired design"},
	{ID: models.ThemeClassic, Name: "Classic", Description: "Simple, traditional layout"},
	{ID: models.ThemeDiner, Name: "Roadside Diner", Description: "1950s Americana with teal booths and chrome accents"},
	{ID: models.ThemeTrattoria, Name: "Rustic Trattoria", Description: "Italian countryside warmth with terracotta and olive"},
	{ID: models.ThemeKuche, Name: "German Farmhouse", Description: "Cozy Alpine kitchen with forest green and natural wood"},
	{ID: models.ThemeNightowl, Name: "Night Owl Cafe", Description: "Dark theme with amber glow for late-night cooking"},
	{ID: models.ThemeMilkbar, Name: "Retro Milk Bar", Description: "Cheerful pastels inspired by 1950s soda fountains"},
	{ID: models.ThemeBodega, Name: "Spanish Bodega", Description: "Wine cellar elegance with deep reds and warm golds"},
	{ID: models.ThemeMarket, Name: "Farmers Market", Description: "Fresh chalkboard-style with handwritten character"},
	{ID: models.ThemeBistro, Name: "Parisian Bistro", Description: "Classic French elegance in black, white, and gold"},
	{ID: models.ThemeComfort, Name: "Grandma's Kitchen", Description: "Nostalgic comfort with gingham and worn recipe cards"},
	{ID: models.ThemeSpeakeasy, Name: "Art Deco Speakeasy", Description: "1920s glamour with geometric gold and deep black"},
	{ID: models.ThemeCuchifritos, Name: "NYC Lunch Counter", Description: "Neon signs, pink tiles, and neighborhood warmth"},
	{ID: models.ThemePizzeria, Name: "NY Slice Shop", Description: "Bold typography on pink with that New York slice energy"},
}

func (h *Handler) GetThemeSettingsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userInfo := auth.GetUserInfoFromContext(ctx)

	currentTheme := models.DefaultTheme
	if prefs, err := h.UserPreferencesStore.Get(ctx, userInfo.UserID); err == nil && prefs.Theme != "" {
		currentTheme = prefs.Theme
	}

	data := ThemeSettingsData{
		UserInfo:     userInfo,
		CurrentTheme: currentTheme,
		Themes:       AvailableThemes,
		Success:      r.URL.Query().Get("success"),
		Error:        r.URL.Query().Get("error"),
	}
	h.Renderer.RenderPage(w, "account-theme.gohtml", data)
}

func (h *Handler) SetThemeHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userInfo := auth.GetUserInfoFromContext(ctx)
	if !userInfo.IsLoggedIn {
		h.Renderer.RenderError(w, r, http.StatusUnauthorized, "You must be logged in to change theme.")
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/account/theme?error=Invalid form data.", http.StatusSeeOther)
		return
	}

	theme := r.FormValue("theme")

	validTheme := false
	for _, t := range AvailableThemes {
		if t.ID == theme {
			validTheme = true
			break
		}
	}
	if !validTheme {
		http.Redirect(w, r, "/account/theme?error=Invalid theme selection.", http.StatusSeeOther)
		return
	}

	err := h.UserPreferencesStore.SetTheme(ctx, userInfo.UserID, theme)
	if err != nil {
		logging.AddError(ctx, err, "Failed to save theme preference")
		http.Redirect(w, r, "/account/theme?error=Failed to save theme.", http.StatusSeeOther)
		return
	}

	logging.AddMany(ctx, map[string]any{
		"action": "account.theme.set",
		"theme":  theme,
	})

	http.Redirect(w, r, "/account/theme?success=Theme updated.", http.StatusSeeOther)
}
