package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/mr-flannery/go-recipe-book/src/models"
	"github.com/mr-flannery/go-recipe-book/src/store"
	"github.com/mr-flannery/go-recipe-book/src/store/mocks"
)

func TestSearchTagsHandler_SearchesTagsByQuery(t *testing.T) {
	t.Run("returns matching tags when query matches", func(t *testing.T) {
		mockTagStore := &mocks.MockTagStore{
			SearchFunc: func(query string) ([]models.Tag, error) {
				return []models.Tag{
					{ID: 1, Name: "breakfast"},
					{ID: 2, Name: "brunch"},
				}, nil
			},
		}

		h := &Handler{TagStore: mockTagStore}
		req := httptest.NewRequest(http.MethodGet, "/api/tags/search?q=br", nil)
		rec := httptest.NewRecorder()

		h.SearchTagsHandler(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var response TagSearchResponse
		json.NewDecoder(rec.Body).Decode(&response)

		if len(response.Tags) != 2 {
			t.Errorf("expected 2 tags, got %d", len(response.Tags))
		}
		if response.Tags[0] != "breakfast" {
			t.Errorf("expected first tag 'breakfast', got '%s'", response.Tags[0])
		}
	})

	t.Run("returns empty list when no tags match query", func(t *testing.T) {
		mockTagStore := &mocks.MockTagStore{
			SearchFunc: func(query string) ([]models.Tag, error) {
				return []models.Tag{}, nil
			},
		}

		h := &Handler{TagStore: mockTagStore}
		req := httptest.NewRequest(http.MethodGet, "/api/tags/search?q=xyz", nil)
		rec := httptest.NewRecorder()

		h.SearchTagsHandler(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var response TagSearchResponse
		json.NewDecoder(rec.Body).Decode(&response)

		if len(response.Tags) != 0 {
			t.Errorf("expected 0 tags, got %d", len(response.Tags))
		}
	})

	t.Run("returns error when store fails", func(t *testing.T) {
		mockTagStore := &mocks.MockTagStore{
			SearchFunc: func(query string) ([]models.Tag, error) {
				return nil, errors.New("database error")
			},
		}

		h := &Handler{TagStore: mockTagStore}
		req := httptest.NewRequest(http.MethodGet, "/api/tags/search?q=test", nil)
		rec := httptest.NewRecorder()

		h.SearchTagsHandler(rec, req)

		if rec.Code != http.StatusInternalServerError {
			t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
		}

		var response TagResponse
		json.NewDecoder(rec.Body).Decode(&response)

		if response.Success {
			t.Error("expected success to be false")
		}
	})
}

func TestSearchUserTagsHandler_SearchesUserSpecificTags(t *testing.T) {
	t.Run("returns unauthorized when session is invalid", func(t *testing.T) {
		mockAuthStore := &mocks.MockAuthStore{
			GetSessionFunc: func(sessionID string) (*store.Session, error) {
				return nil, errors.New("session not found")
			},
		}

		h := &Handler{AuthStore: mockAuthStore}
		req := httptest.NewRequest(http.MethodGet, "/api/tags/user/search?q=test", nil)
		rec := httptest.NewRecorder()

		h.SearchUserTagsHandler(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}
	})

	t.Run("returns matching user tags when authenticated", func(t *testing.T) {
		mockAuthStore := &mocks.MockAuthStore{
			GetSessionFunc: func(sessionID string) (*store.Session, error) {
				return &store.Session{ID: sessionID, UserID: 1}, nil
			},
			GetUserByIDFunc: func(userID int) (*store.AuthUser, error) {
				return &store.AuthUser{ID: userID, Username: "testuser"}, nil
			},
		}
		mockUserTagStore := &mocks.MockUserTagStore{
			SearchFunc: func(userID int, query string) ([]string, error) {
				return []string{"favorite", "family"}, nil
			},
		}

		h := &Handler{AuthStore: mockAuthStore, UserTagStore: mockUserTagStore}
		req := httptest.NewRequest(http.MethodGet, "/api/tags/user/search?q=fa", nil)
		req.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
		rec := httptest.NewRecorder()

		h.SearchUserTagsHandler(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var response TagSearchResponse
		json.NewDecoder(rec.Body).Decode(&response)

		if len(response.Tags) != 2 {
			t.Errorf("expected 2 tags, got %d", len(response.Tags))
		}
	})
}

func TestAddTagToRecipeHandler_AddsTagToRecipe(t *testing.T) {
	t.Run("returns error when recipe ID is missing", func(t *testing.T) {
		h := &Handler{}
		req := httptest.NewRequest(http.MethodPost, "/recipes//tags", nil)
		req.SetPathValue("id", "")
		rec := httptest.NewRecorder()

		h.AddTagToRecipeHandler(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("returns error when recipe ID is invalid", func(t *testing.T) {
		h := &Handler{}
		req := httptest.NewRequest(http.MethodPost, "/recipes/abc/tags", nil)
		req.SetPathValue("id", "abc")
		rec := httptest.NewRecorder()

		h.AddTagToRecipeHandler(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("returns unauthorized when session is invalid", func(t *testing.T) {
		mockAuthStore := &mocks.MockAuthStore{
			GetSessionFunc: func(sessionID string) (*store.Session, error) {
				return nil, errors.New("session not found")
			},
		}

		h := &Handler{AuthStore: mockAuthStore}
		req := httptest.NewRequest(http.MethodPost, "/recipes/1/tags", nil)
		req.SetPathValue("id", "1")
		rec := httptest.NewRecorder()

		h.AddTagToRecipeHandler(rec, req)

		if rec.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
		}
	})

	t.Run("returns not found when recipe does not exist", func(t *testing.T) {
		mockAuthStore := &mocks.MockAuthStore{
			GetSessionFunc: func(sessionID string) (*store.Session, error) {
				return &store.Session{ID: sessionID, UserID: 1}, nil
			},
			GetUserByIDFunc: func(userID int) (*store.AuthUser, error) {
				return &store.AuthUser{ID: userID, Username: "testuser"}, nil
			},
		}
		mockRecipeStore := &mocks.MockRecipeStore{
			GetByIDFunc: func(id string) (models.Recipe, error) {
				return models.Recipe{}, errors.New("not found")
			},
		}

		h := &Handler{AuthStore: mockAuthStore, RecipeStore: mockRecipeStore}
		req := httptest.NewRequest(http.MethodPost, "/recipes/999/tags", nil)
		req.SetPathValue("id", "999")
		req.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
		rec := httptest.NewRecorder()

		h.AddTagToRecipeHandler(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}
	})

	t.Run("returns forbidden when user is not the recipe author", func(t *testing.T) {
		mockAuthStore := &mocks.MockAuthStore{
			GetSessionFunc: func(sessionID string) (*store.Session, error) {
				return &store.Session{ID: sessionID, UserID: 2}, nil
			},
			GetUserByIDFunc: func(userID int) (*store.AuthUser, error) {
				return &store.AuthUser{ID: userID, Username: "testuser"}, nil
			},
		}
		mockRecipeStore := &mocks.MockRecipeStore{
			GetByIDFunc: func(id string) (models.Recipe, error) {
				return models.Recipe{ID: 1, AuthorID: 1}, nil
			},
		}

		h := &Handler{AuthStore: mockAuthStore, RecipeStore: mockRecipeStore}
		req := httptest.NewRequest(http.MethodPost, "/recipes/1/tags", nil)
		req.SetPathValue("id", "1")
		req.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
		rec := httptest.NewRecorder()

		h.AddTagToRecipeHandler(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Errorf("expected status %d, got %d", http.StatusForbidden, rec.Code)
		}
	})

	t.Run("returns error when tag name is empty", func(t *testing.T) {
		mockAuthStore := &mocks.MockAuthStore{
			GetSessionFunc: func(sessionID string) (*store.Session, error) {
				return &store.Session{ID: sessionID, UserID: 1}, nil
			},
			GetUserByIDFunc: func(userID int) (*store.AuthUser, error) {
				return &store.AuthUser{ID: userID, Username: "testuser"}, nil
			},
		}
		mockRecipeStore := &mocks.MockRecipeStore{
			GetByIDFunc: func(id string) (models.Recipe, error) {
				return models.Recipe{ID: 1, AuthorID: 1}, nil
			},
		}

		h := &Handler{AuthStore: mockAuthStore, RecipeStore: mockRecipeStore}
		form := url.Values{}
		form.Add("tag", "")
		req := httptest.NewRequest(http.MethodPost, "/recipes/1/tags", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.SetPathValue("id", "1")
		req.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
		rec := httptest.NewRecorder()

		h.AddTagToRecipeHandler(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("adds tag and returns success when all checks pass", func(t *testing.T) {
		mockAuthStore := &mocks.MockAuthStore{
			GetSessionFunc: func(sessionID string) (*store.Session, error) {
				return &store.Session{ID: sessionID, UserID: 1}, nil
			},
			GetUserByIDFunc: func(userID int) (*store.AuthUser, error) {
				return &store.AuthUser{ID: userID, Username: "testuser"}, nil
			},
		}
		mockRecipeStore := &mocks.MockRecipeStore{
			GetByIDFunc: func(id string) (models.Recipe, error) {
				return models.Recipe{ID: 1, AuthorID: 1}, nil
			},
		}
		mockTagStore := &mocks.MockTagStore{
			GetOrCreateFunc: func(name string) (models.Tag, error) {
				return models.Tag{ID: 5, Name: name}, nil
			},
			AddToRecipeFunc: func(recipeID, tagID int) error {
				return nil
			},
		}

		h := &Handler{
			AuthStore:   mockAuthStore,
			RecipeStore: mockRecipeStore,
			TagStore:    mockTagStore,
		}
		form := url.Values{}
		form.Add("tag", "breakfast")
		req := httptest.NewRequest(http.MethodPost, "/recipes/1/tags", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.SetPathValue("id", "1")
		req.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
		rec := httptest.NewRecorder()

		h.AddTagToRecipeHandler(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var response TagResponse
		json.NewDecoder(rec.Body).Decode(&response)
		if !response.Success {
			t.Error("expected success to be true")
		}
	})
}

func TestRemoveTagFromRecipeHandler_RemovesTagFromRecipe(t *testing.T) {
	t.Run("returns error when IDs are missing", func(t *testing.T) {
		h := &Handler{}
		req := httptest.NewRequest(http.MethodDelete, "/recipes/1/tags/", nil)
		req.SetPathValue("id", "1")
		req.SetPathValue("tagId", "")
		rec := httptest.NewRecorder()

		h.RemoveTagFromRecipeHandler(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("removes tag and returns success when authorized", func(t *testing.T) {
		mockAuthStore := &mocks.MockAuthStore{
			GetSessionFunc: func(sessionID string) (*store.Session, error) {
				return &store.Session{ID: sessionID, UserID: 1}, nil
			},
			GetUserByIDFunc: func(userID int) (*store.AuthUser, error) {
				return &store.AuthUser{ID: userID, Username: "testuser"}, nil
			},
		}
		mockRecipeStore := &mocks.MockRecipeStore{
			GetByIDFunc: func(id string) (models.Recipe, error) {
				return models.Recipe{ID: 1, AuthorID: 1}, nil
			},
		}
		mockTagStore := &mocks.MockTagStore{
			RemoveFromRecipeFunc: func(recipeID, tagID int) error {
				return nil
			},
		}

		h := &Handler{
			AuthStore:   mockAuthStore,
			RecipeStore: mockRecipeStore,
			TagStore:    mockTagStore,
		}
		req := httptest.NewRequest(http.MethodDelete, "/recipes/1/tags/5", nil)
		req.SetPathValue("id", "1")
		req.SetPathValue("tagId", "5")
		req.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
		rec := httptest.NewRecorder()

		h.RemoveTagFromRecipeHandler(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		var response TagResponse
		json.NewDecoder(rec.Body).Decode(&response)
		if !response.Success {
			t.Error("expected success to be true")
		}
	})
}

func TestAddUserTagToRecipeHandler_AddsPersonalTagToRecipe(t *testing.T) {
	t.Run("adds user tag and returns success when authenticated", func(t *testing.T) {
		mockAuthStore := &mocks.MockAuthStore{
			GetSessionFunc: func(sessionID string) (*store.Session, error) {
				return &store.Session{ID: sessionID, UserID: 1}, nil
			},
			GetUserByIDFunc: func(userID int) (*store.AuthUser, error) {
				return &store.AuthUser{ID: userID, Username: "testuser"}, nil
			},
		}
		mockRecipeStore := &mocks.MockRecipeStore{
			GetByIDFunc: func(id string) (models.Recipe, error) {
				return models.Recipe{ID: 1, AuthorID: 2}, nil
			},
		}
		mockUserTagStore := &mocks.MockUserTagStore{
			GetOrCreateFunc: func(userID, recipeID int, name string) (models.UserTag, error) {
				return models.UserTag{ID: 10, UserID: userID, RecipeID: recipeID, Name: name}, nil
			},
		}

		h := &Handler{
			AuthStore:    mockAuthStore,
			RecipeStore:  mockRecipeStore,
			UserTagStore: mockUserTagStore,
		}
		form := url.Values{}
		form.Add("tag", "must-try")
		req := httptest.NewRequest(http.MethodPost, "/recipes/1/user-tags", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.SetPathValue("id", "1")
		req.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
		rec := httptest.NewRecorder()

		h.AddUserTagToRecipeHandler(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
	})
}

func TestRemoveUserTagHandler_RemovesPersonalTagFromRecipe(t *testing.T) {
	t.Run("returns error when tag ID is missing", func(t *testing.T) {
		h := &Handler{}
		req := httptest.NewRequest(http.MethodDelete, "/user-tags/", nil)
		req.SetPathValue("tagId", "")
		rec := httptest.NewRecorder()

		h.RemoveUserTagHandler(rec, req)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("removes user tag and returns success when authenticated", func(t *testing.T) {
		mockAuthStore := &mocks.MockAuthStore{
			GetSessionFunc: func(sessionID string) (*store.Session, error) {
				return &store.Session{ID: sessionID, UserID: 1}, nil
			},
			GetUserByIDFunc: func(userID int) (*store.AuthUser, error) {
				return &store.AuthUser{ID: userID, Username: "testuser"}, nil
			},
		}
		var removedUserID, removedTagID int
		mockUserTagStore := &mocks.MockUserTagStore{
			RemoveFunc: func(userID, tagID int) error {
				removedUserID = userID
				removedTagID = tagID
				return nil
			},
		}

		h := &Handler{
			AuthStore:    mockAuthStore,
			UserTagStore: mockUserTagStore,
		}
		req := httptest.NewRequest(http.MethodDelete, "/user-tags/10", nil)
		req.SetPathValue("tagId", "10")
		req.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
		rec := httptest.NewRecorder()

		h.RemoveUserTagHandler(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}

		if removedUserID != 1 {
			t.Errorf("expected user ID 1, got %d", removedUserID)
		}
		if removedTagID != 10 {
			t.Errorf("expected tag ID 10, got %d", removedTagID)
		}
	})
}

func TestNewHandler_InitializesHandlerWithAllStores(t *testing.T) {
	mockRecipeStore := &mocks.MockRecipeStore{}
	mockTagStore := &mocks.MockTagStore{}
	mockUserTagStore := &mocks.MockUserTagStore{}
	mockCommentStore := &mocks.MockCommentStore{}
	mockUserStore := &mocks.MockUserStore{}
	mockAuthStore := &mocks.MockAuthStore{}

	h := NewHandler(nil, mockRecipeStore, mockTagStore, mockUserTagStore, mockCommentStore, mockUserStore, mockAuthStore)

	if h == nil {
		t.Fatal("expected handler, got nil")
	}
	if h.RecipeStore == nil {
		t.Error("expected RecipeStore to be set")
	}
	if h.TagStore == nil {
		t.Error("expected TagStore to be set")
	}
	if h.UserTagStore == nil {
		t.Error("expected UserTagStore to be set")
	}
	if h.CommentStore == nil {
		t.Error("expected CommentStore to be set")
	}
	if h.UserStore == nil {
		t.Error("expected UserStore to be set")
	}
	if h.AuthStore == nil {
		t.Error("expected AuthStore to be set")
	}
}
