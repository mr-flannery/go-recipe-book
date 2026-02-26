package mocks

import "github.com/mr-flannery/go-recipe-book/src/models"

type MockRecipeStore struct {
	SaveFunc          func(recipe models.Recipe) (int, error)
	GetByIDFunc       func(id string) (models.Recipe, error)
	UpdateFunc        func(recipe models.Recipe) error
	DeleteFunc        func(id string) error
	GetAllFunc        func() ([]models.Recipe, error)
	GetFilteredFunc   func(params models.FilterParams) ([]models.Recipe, error)
	CountFilteredFunc func(params models.FilterParams) (int, error)
	GetRandomIDFunc   func() (int, error)
	SearchByTitleFunc func(query string, limit int) ([]models.RecipeSearchResult, error)
}

func (m *MockRecipeStore) Save(recipe models.Recipe) (int, error) {
	if m.SaveFunc != nil {
		return m.SaveFunc(recipe)
	}
	return 0, nil
}

func (m *MockRecipeStore) GetByID(id string) (models.Recipe, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(id)
	}
	return models.Recipe{}, nil
}

func (m *MockRecipeStore) Update(recipe models.Recipe) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(recipe)
	}
	return nil
}

func (m *MockRecipeStore) Delete(id string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(id)
	}
	return nil
}

func (m *MockRecipeStore) GetAll() ([]models.Recipe, error) {
	if m.GetAllFunc != nil {
		return m.GetAllFunc()
	}
	return nil, nil
}

func (m *MockRecipeStore) GetFiltered(params models.FilterParams) ([]models.Recipe, error) {
	if m.GetFilteredFunc != nil {
		return m.GetFilteredFunc(params)
	}
	return nil, nil
}

func (m *MockRecipeStore) CountFiltered(params models.FilterParams) (int, error) {
	if m.CountFilteredFunc != nil {
		return m.CountFilteredFunc(params)
	}
	return 0, nil
}

func (m *MockRecipeStore) GetRandomID() (int, error) {
	if m.GetRandomIDFunc != nil {
		return m.GetRandomIDFunc()
	}
	return 0, nil
}

func (m *MockRecipeStore) SearchByTitle(query string, limit int) ([]models.RecipeSearchResult, error) {
	if m.SearchByTitleFunc != nil {
		return m.SearchByTitleFunc(query, limit)
	}
	return nil, nil
}

type MockTagStore struct {
	GetOrCreateFunc      func(name string) (models.Tag, error)
	SearchFunc           func(query string) ([]models.Tag, error)
	GetByRecipeIDFunc    func(recipeID int) ([]models.Tag, error)
	GetForRecipesFunc    func(recipeIDs []int) (map[int][]models.Tag, error)
	AddToRecipeFunc      func(recipeID, tagID int) error
	RemoveFromRecipeFunc func(recipeID, tagID int) error
	SetRecipeTagsFunc    func(recipeID int, tagNames []string) error
}

func (m *MockTagStore) GetOrCreate(name string) (models.Tag, error) {
	if m.GetOrCreateFunc != nil {
		return m.GetOrCreateFunc(name)
	}
	return models.Tag{}, nil
}

func (m *MockTagStore) Search(query string) ([]models.Tag, error) {
	if m.SearchFunc != nil {
		return m.SearchFunc(query)
	}
	return nil, nil
}

func (m *MockTagStore) GetByRecipeID(recipeID int) ([]models.Tag, error) {
	if m.GetByRecipeIDFunc != nil {
		return m.GetByRecipeIDFunc(recipeID)
	}
	return nil, nil
}

func (m *MockTagStore) GetForRecipes(recipeIDs []int) (map[int][]models.Tag, error) {
	if m.GetForRecipesFunc != nil {
		return m.GetForRecipesFunc(recipeIDs)
	}
	return nil, nil
}

func (m *MockTagStore) AddToRecipe(recipeID, tagID int) error {
	if m.AddToRecipeFunc != nil {
		return m.AddToRecipeFunc(recipeID, tagID)
	}
	return nil
}

func (m *MockTagStore) RemoveFromRecipe(recipeID, tagID int) error {
	if m.RemoveFromRecipeFunc != nil {
		return m.RemoveFromRecipeFunc(recipeID, tagID)
	}
	return nil
}

func (m *MockTagStore) SetRecipeTags(recipeID int, tagNames []string) error {
	if m.SetRecipeTagsFunc != nil {
		return m.SetRecipeTagsFunc(recipeID, tagNames)
	}
	return nil
}

type MockUserTagStore struct {
	GetOrCreateFunc   func(userID, recipeID int, name string) (models.UserTag, error)
	SearchFunc        func(userID int, query string) ([]string, error)
	GetByRecipeIDFunc func(userID, recipeID int) ([]models.UserTag, error)
	GetByUserIDFunc   func(userID int) ([]models.UserTag, error)
	GetForRecipesFunc func(userID int, recipeIDs []int) (map[int][]models.UserTag, error)
	RemoveFunc        func(userID, tagID int) error
}

func (m *MockUserTagStore) GetOrCreate(userID, recipeID int, name string) (models.UserTag, error) {
	if m.GetOrCreateFunc != nil {
		return m.GetOrCreateFunc(userID, recipeID, name)
	}
	return models.UserTag{}, nil
}

func (m *MockUserTagStore) Search(userID int, query string) ([]string, error) {
	if m.SearchFunc != nil {
		return m.SearchFunc(userID, query)
	}
	return nil, nil
}

func (m *MockUserTagStore) GetByRecipeID(userID, recipeID int) ([]models.UserTag, error) {
	if m.GetByRecipeIDFunc != nil {
		return m.GetByRecipeIDFunc(userID, recipeID)
	}
	return nil, nil
}

func (m *MockUserTagStore) GetByUserID(userID int) ([]models.UserTag, error) {
	if m.GetByUserIDFunc != nil {
		return m.GetByUserIDFunc(userID)
	}
	return nil, nil
}

func (m *MockUserTagStore) GetForRecipes(userID int, recipeIDs []int) (map[int][]models.UserTag, error) {
	if m.GetForRecipesFunc != nil {
		return m.GetForRecipesFunc(userID, recipeIDs)
	}
	return nil, nil
}

func (m *MockUserTagStore) Remove(userID, tagID int) error {
	if m.RemoveFunc != nil {
		return m.RemoveFunc(userID, tagID)
	}
	return nil
}

type MockCommentStore struct {
	GetByRecipeIDFunc            func(recipeID string) ([]models.Comment, error)
	GetByIDFunc                  func(commentID int) (models.Comment, error)
	GetByUserIDFunc              func(userID int) ([]models.Comment, error)
	SaveFunc                     func(comment models.Comment) error
	UpdateFunc                   func(commentID int, content string) error
	DeleteFunc                   func(commentID int) error
	GetLatestByUserAndRecipeFunc func(userID, recipeID int) (models.Comment, error)
}

func (m *MockCommentStore) GetByRecipeID(recipeID string) ([]models.Comment, error) {
	if m.GetByRecipeIDFunc != nil {
		return m.GetByRecipeIDFunc(recipeID)
	}
	return nil, nil
}

func (m *MockCommentStore) GetByID(commentID int) (models.Comment, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(commentID)
	}
	return models.Comment{}, nil
}

func (m *MockCommentStore) GetByUserID(userID int) ([]models.Comment, error) {
	if m.GetByUserIDFunc != nil {
		return m.GetByUserIDFunc(userID)
	}
	return nil, nil
}

func (m *MockCommentStore) Save(comment models.Comment) error {
	if m.SaveFunc != nil {
		return m.SaveFunc(comment)
	}
	return nil
}

func (m *MockCommentStore) Update(commentID int, content string) error {
	if m.UpdateFunc != nil {
		return m.UpdateFunc(commentID, content)
	}
	return nil
}

func (m *MockCommentStore) Delete(commentID int) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(commentID)
	}
	return nil
}

func (m *MockCommentStore) GetLatestByUserAndRecipe(userID, recipeID int) (models.Comment, error) {
	if m.GetLatestByUserAndRecipeFunc != nil {
		return m.GetLatestByUserAndRecipeFunc(userID, recipeID)
	}
	return models.Comment{}, nil
}

type MockUserStore struct {
	GetUsernameByIDFunc func(userID int) (string, error)
}

func (m *MockUserStore) GetUsernameByID(userID int) (string, error) {
	if m.GetUsernameByIDFunc != nil {
		return m.GetUsernameByIDFunc(userID)
	}
	return "", nil
}

type MockIngredientStore struct {
	SearchFunc      func(query string, limit int) ([]string, error)
	GetOrCreateFunc func(name string) (int, error)
}

func (m *MockIngredientStore) Search(query string, limit int) ([]string, error) {
	if m.SearchFunc != nil {
		return m.SearchFunc(query, limit)
	}
	return nil, nil
}

func (m *MockIngredientStore) GetOrCreate(name string) (int, error) {
	if m.GetOrCreateFunc != nil {
		return m.GetOrCreateFunc(name)
	}
	return 0, nil
}
