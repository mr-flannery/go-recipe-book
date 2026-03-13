package mocks

import (
	"context"

	"github.com/mr-flannery/go-recipe-book/src/models"
	"github.com/mr-flannery/go-recipe-book/src/store"
)

type MockRecipeStore struct {
	SaveFunc          func(ctx context.Context, recipe models.Recipe) (int, error)
	GetByIDFunc       func(ctx context.Context, id string) (models.Recipe, error)
	UpdateFunc        func(ctx context.Context, recipe models.Recipe) error
	DeleteFunc        func(ctx context.Context, id string) error
	GetAllFunc        func(ctx context.Context) ([]models.Recipe, error)
	GetFilteredFunc   func(ctx context.Context, params models.FilterParams) ([]models.Recipe, error)
	CountFilteredFunc func(ctx context.Context, params models.FilterParams) (int, error)
	GetRandomIDFunc   func(ctx context.Context) (int, error)
	SearchByTitleFunc func(ctx context.Context, query string, limit int) ([]models.RecipeSearchResult, error)
}

func (m *MockRecipeStore) Save(ctx context.Context, recipe models.Recipe) (int, error) {
	if m.SaveFunc != nil {
		return m.SaveFunc(ctx, recipe)
	}
	return 0, nil
}

func (m *MockRecipeStore) GetByID(ctx context.Context, id string) (models.Recipe, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return models.Recipe{}, nil
}

func (m *MockRecipeStore) Update(ctx context.Context, recipe models.Recipe) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, recipe)
	}
	return nil
}

func (m *MockRecipeStore) Delete(ctx context.Context, id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, id)
	}
	return nil
}

func (m *MockRecipeStore) GetAll(ctx context.Context) ([]models.Recipe, error) {
	if m.GetAllFunc != nil {
		return m.GetAllFunc(ctx)
	}
	return nil, nil
}

func (m *MockRecipeStore) GetFiltered(ctx context.Context, params models.FilterParams) ([]models.Recipe, error) {
	if m.GetFilteredFunc != nil {
		return m.GetFilteredFunc(ctx, params)
	}
	return nil, nil
}

func (m *MockRecipeStore) CountFiltered(ctx context.Context, params models.FilterParams) (int, error) {
	if m.CountFilteredFunc != nil {
		return m.CountFilteredFunc(ctx, params)
	}
	return 0, nil
}

func (m *MockRecipeStore) GetRandomID(ctx context.Context) (int, error) {
	if m.GetRandomIDFunc != nil {
		return m.GetRandomIDFunc(ctx)
	}
	return 0, nil
}

func (m *MockRecipeStore) SearchByTitle(ctx context.Context, query string, limit int) ([]models.RecipeSearchResult, error) {
	if m.SearchByTitleFunc != nil {
		return m.SearchByTitleFunc(ctx, query, limit)
	}
	return nil, nil
}

type MockTagStore struct {
	GetOrCreateFunc      func(ctx context.Context, name string) (models.Tag, error)
	SearchFunc           func(ctx context.Context, query string) ([]models.Tag, error)
	GetByRecipeIDFunc    func(ctx context.Context, recipeID int) ([]models.Tag, error)
	GetForRecipesFunc    func(ctx context.Context, recipeIDs []int) (map[int][]models.Tag, error)
	AddToRecipeFunc      func(ctx context.Context, recipeID, tagID int) error
	RemoveFromRecipeFunc func(ctx context.Context, recipeID, tagID int) error
	SetRecipeTagsFunc    func(ctx context.Context, recipeID int, tagNames []string) error
}

func (m *MockTagStore) GetOrCreate(ctx context.Context, name string) (models.Tag, error) {
	if m.GetOrCreateFunc != nil {
		return m.GetOrCreateFunc(ctx, name)
	}
	return models.Tag{}, nil
}

func (m *MockTagStore) Search(ctx context.Context, query string) ([]models.Tag, error) {
	if m.SearchFunc != nil {
		return m.SearchFunc(ctx, query)
	}
	return nil, nil
}

func (m *MockTagStore) GetByRecipeID(ctx context.Context, recipeID int) ([]models.Tag, error) {
	if m.GetByRecipeIDFunc != nil {
		return m.GetByRecipeIDFunc(ctx, recipeID)
	}
	return nil, nil
}

func (m *MockTagStore) GetForRecipes(ctx context.Context, recipeIDs []int) (map[int][]models.Tag, error) {
	if m.GetForRecipesFunc != nil {
		return m.GetForRecipesFunc(ctx, recipeIDs)
	}
	return nil, nil
}

func (m *MockTagStore) AddToRecipe(ctx context.Context, recipeID, tagID int) error {
	if m.AddToRecipeFunc != nil {
		return m.AddToRecipeFunc(ctx, recipeID, tagID)
	}
	return nil
}

func (m *MockTagStore) RemoveFromRecipe(ctx context.Context, recipeID, tagID int) error {
	if m.RemoveFromRecipeFunc != nil {
		return m.RemoveFromRecipeFunc(ctx, recipeID, tagID)
	}
	return nil
}

func (m *MockTagStore) SetRecipeTags(ctx context.Context, recipeID int, tagNames []string) error {
	if m.SetRecipeTagsFunc != nil {
		return m.SetRecipeTagsFunc(ctx, recipeID, tagNames)
	}
	return nil
}

type MockUserTagStore struct {
	GetOrCreateFunc   func(ctx context.Context, userID, recipeID int, name string) (models.UserTag, error)
	SearchFunc        func(ctx context.Context, userID int, query string) ([]string, error)
	GetByRecipeIDFunc func(ctx context.Context, userID, recipeID int) ([]models.UserTag, error)
	GetByUserIDFunc   func(ctx context.Context, userID int) ([]models.UserTag, error)
	GetForRecipesFunc func(ctx context.Context, userID int, recipeIDs []int) (map[int][]models.UserTag, error)
	RemoveFunc        func(ctx context.Context, userID, tagID int) error
}

func (m *MockUserTagStore) GetOrCreate(ctx context.Context, userID, recipeID int, name string) (models.UserTag, error) {
	if m.GetOrCreateFunc != nil {
		return m.GetOrCreateFunc(ctx, userID, recipeID, name)
	}
	return models.UserTag{}, nil
}

func (m *MockUserTagStore) Search(ctx context.Context, userID int, query string) ([]string, error) {
	if m.SearchFunc != nil {
		return m.SearchFunc(ctx, userID, query)
	}
	return nil, nil
}

func (m *MockUserTagStore) GetByRecipeID(ctx context.Context, userID, recipeID int) ([]models.UserTag, error) {
	if m.GetByRecipeIDFunc != nil {
		return m.GetByRecipeIDFunc(ctx, userID, recipeID)
	}
	return nil, nil
}

func (m *MockUserTagStore) GetByUserID(ctx context.Context, userID int) ([]models.UserTag, error) {
	if m.GetByUserIDFunc != nil {
		return m.GetByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockUserTagStore) GetForRecipes(ctx context.Context, userID int, recipeIDs []int) (map[int][]models.UserTag, error) {
	if m.GetForRecipesFunc != nil {
		return m.GetForRecipesFunc(ctx, userID, recipeIDs)
	}
	return nil, nil
}

func (m *MockUserTagStore) Remove(ctx context.Context, userID, tagID int) error {
	if m.RemoveFunc != nil {
		return m.RemoveFunc(ctx, userID, tagID)
	}
	return nil
}

type MockCommentStore struct {
	GetByRecipeIDFunc            func(ctx context.Context, recipeID string) ([]models.Comment, error)
	GetByIDFunc                  func(ctx context.Context, commentID int) (models.Comment, error)
	GetByUserIDFunc              func(ctx context.Context, userID int) ([]models.Comment, error)
	SaveFunc                     func(ctx context.Context, comment models.Comment) error
	UpdateFunc                   func(ctx context.Context, commentID int, content string) error
	DeleteFunc                   func(ctx context.Context, commentID int) error
	GetLatestByUserAndRecipeFunc func(ctx context.Context, userID, recipeID int) (models.Comment, error)
}

func (m *MockCommentStore) GetByRecipeID(ctx context.Context, recipeID string) ([]models.Comment, error) {
	if m.GetByRecipeIDFunc != nil {
		return m.GetByRecipeIDFunc(ctx, recipeID)
	}
	return nil, nil
}

func (m *MockCommentStore) GetByID(ctx context.Context, commentID int) (models.Comment, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, commentID)
	}
	return models.Comment{}, nil
}

func (m *MockCommentStore) GetByUserID(ctx context.Context, userID int) ([]models.Comment, error) {
	if m.GetByUserIDFunc != nil {
		return m.GetByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockCommentStore) Save(ctx context.Context, comment models.Comment) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(ctx, comment)
	}
	return nil
}

func (m *MockCommentStore) Update(ctx context.Context, commentID int, content string) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(ctx, commentID, content)
	}
	return nil
}

func (m *MockCommentStore) Delete(ctx context.Context, commentID int) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, commentID)
	}
	return nil
}

func (m *MockCommentStore) GetLatestByUserAndRecipe(ctx context.Context, userID, recipeID int) (models.Comment, error) {
	if m.GetLatestByUserAndRecipeFunc != nil {
		return m.GetLatestByUserAndRecipeFunc(ctx, userID, recipeID)
	}
	return models.Comment{}, nil
}

type MockUserStore struct {
	GetUsernameByIDFunc func(ctx context.Context, userID int) (string, error)
}

func (m *MockUserStore) GetUsernameByID(ctx context.Context, userID int) (string, error) {
	if m.GetUsernameByIDFunc != nil {
		return m.GetUsernameByIDFunc(ctx, userID)
	}
	return "", nil
}

type MockIngredientStore struct {
	SearchFunc      func(ctx context.Context, query string, limit int) ([]string, error)
	GetOrCreateFunc func(ctx context.Context, name string) (int, error)
}

func (m *MockIngredientStore) Search(ctx context.Context, query string, limit int) ([]string, error) {
	if m.SearchFunc != nil {
		return m.SearchFunc(ctx, query, limit)
	}
	return nil, nil
}

func (m *MockIngredientStore) GetOrCreate(ctx context.Context, name string) (int, error) {
	if m.GetOrCreateFunc != nil {
		return m.GetOrCreateFunc(ctx, name)
	}
	return 0, nil
}

type MockUserPreferencesStore struct {
	GetFunc         func(ctx context.Context, userID int) (*models.UserPreferences, error)
	SetPageSizeFunc func(ctx context.Context, userID, pageSize int) error
	SetViewModeFunc func(ctx context.Context, userID int, viewMode string) error
	SetThemeFunc    func(ctx context.Context, userID int, theme string) error
}

func (m *MockUserPreferencesStore) Get(ctx context.Context, userID int) (*models.UserPreferences, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockUserPreferencesStore) SetPageSize(ctx context.Context, userID, pageSize int) error {
	if m.SetPageSizeFunc != nil {
		return m.SetPageSizeFunc(ctx, userID, pageSize)
	}
	return nil
}

func (m *MockUserPreferencesStore) SetViewMode(ctx context.Context, userID int, viewMode string) error {
	if m.SetViewModeFunc != nil {
		return m.SetViewModeFunc(ctx, userID, viewMode)
	}
	return nil
}

func (m *MockUserPreferencesStore) SetTheme(ctx context.Context, userID int, theme string) error {
	if m.SetThemeFunc != nil {
		return m.SetThemeFunc(ctx, userID, theme)
	}
	return nil
}

type MockAPIKeyStore struct {
	CreateFunc         func(ctx context.Context, userID int, name string, keyHash string, keyPrefix string, encryptedKey string) (int, error)
	GetByKeyHashFunc   func(ctx context.Context, keyHash string) (*store.APIKey, error)
	GetByUserIDFunc    func(ctx context.Context, userID int) ([]store.APIKey, error)
	DeleteFunc         func(ctx context.Context, userID int, keyID int) error
	UpdateLastUsedFunc func(ctx context.Context, keyID int) error
}

func (m *MockAPIKeyStore) Create(ctx context.Context, userID int, name string, keyHash string, keyPrefix string, encryptedKey string) (int, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, userID, name, keyHash, keyPrefix, encryptedKey)
	}
	return 0, nil
}

func (m *MockAPIKeyStore) GetByKeyHash(ctx context.Context, keyHash string) (*store.APIKey, error) {
	if m.GetByKeyHashFunc != nil {
		return m.GetByKeyHashFunc(ctx, keyHash)
	}
	return nil, nil
}

func (m *MockAPIKeyStore) GetByUserID(ctx context.Context, userID int) ([]store.APIKey, error) {
	if m.GetByUserIDFunc != nil {
		return m.GetByUserIDFunc(ctx, userID)
	}
	return nil, nil
}

func (m *MockAPIKeyStore) Delete(ctx context.Context, userID int, keyID int) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, userID, keyID)
	}
	return nil
}

func (m *MockAPIKeyStore) UpdateLastUsed(ctx context.Context, keyID int) error {
	if m.UpdateLastUsedFunc != nil {
		return m.UpdateLastUsedFunc(ctx, keyID)
	}
	return nil
}
