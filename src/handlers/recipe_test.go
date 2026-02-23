package handlers

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/mr-flannery/go-recipe-book/src/auth"
	"github.com/mr-flannery/go-recipe-book/src/models"
	"github.com/mr-flannery/go-recipe-book/src/store"
	"github.com/mr-flannery/go-recipe-book/src/store/mocks"
	tmocks "github.com/mr-flannery/go-recipe-book/src/templates/mocks"
)

func TestListRecipesHandler_ReturnsRecipesWithTags(t *testing.T) {
	mockRecipeStore := &mocks.MockRecipeStore{
		GetFilteredFunc: func(params models.FilterParams) ([]models.Recipe, error) {
			return []models.Recipe{
				{ID: 1, Title: "Test Recipe 1", AuthorID: 1},
				{ID: 2, Title: "Test Recipe 2", AuthorID: 1},
			}, nil
		},
		CountFilteredFunc: func(params models.FilterParams) (int, error) {
			return 2, nil
		},
	}

	mockTagStore := &mocks.MockTagStore{
		GetForRecipesFunc: func(recipeIDs []int) (map[int][]models.Tag, error) {
			return map[int][]models.Tag{
				1: {{ID: 1, Name: "dinner"}},
				2: {{ID: 2, Name: "lunch"}},
			}, nil
		},
	}

	var capturedData any
	mockRenderer := &tmocks.MockRenderer{
		RenderPageFunc: func(w http.ResponseWriter, name string, data any) {
			capturedData = data
			w.WriteHeader(http.StatusOK)
		},
	}

	h := &Handler{
		RecipeStore: mockRecipeStore,
		TagStore:    mockTagStore,
		Renderer:    mockRenderer,
	}

	req := httptest.NewRequest(http.MethodGet, "/recipes", nil)
	userInfo := &auth.UserInfo{IsLoggedIn: false}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.ListRecipesHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if capturedData == nil {
		t.Fatal("expected data to be captured by renderer")
	}
}

func TestListRecipesHandler_ReturnsErrorWhenStoreFails(t *testing.T) {
	mockRecipeStore := &mocks.MockRecipeStore{
		GetFilteredFunc: func(params models.FilterParams) ([]models.Recipe, error) {
			return nil, errors.New("database error")
		},
	}

	mockRenderer := &tmocks.MockRenderer{
		RenderErrorFunc: func(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
			w.WriteHeader(statusCode)
		},
	}

	h := &Handler{
		RecipeStore: mockRecipeStore,
		Renderer:    mockRenderer,
	}

	req := httptest.NewRequest(http.MethodGet, "/recipes", nil)
	rec := httptest.NewRecorder()

	h.ListRecipesHandler(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}
}

func TestViewRecipeHandler_ReturnsRecipeWhenFound(t *testing.T) {
	mockRecipeStore := &mocks.MockRecipeStore{
		GetByIDFunc: func(id string) (models.Recipe, error) {
			return models.Recipe{
				ID:             1,
				Title:          "Test Recipe",
				IngredientsMD:  "- flour",
				InstructionsMD: "Mix",
				AuthorID:       1,
			}, nil
		},
	}

	mockTagStore := &mocks.MockTagStore{
		GetByRecipeIDFunc: func(recipeID int) ([]models.Tag, error) {
			return []models.Tag{{ID: 1, Name: "dinner"}}, nil
		},
	}

	mockCommentStore := &mocks.MockCommentStore{
		GetByRecipeIDFunc: func(recipeID string) ([]models.Comment, error) {
			return []models.Comment{}, nil
		},
	}

	mockAuthStore := &mocks.MockAuthStore{
		GetSessionFunc: func(sessionID string) (*store.Session, error) {
			return nil, errors.New("no session")
		},
	}

	var capturedData any
	mockRenderer := &tmocks.MockRenderer{
		RenderPageFunc: func(w http.ResponseWriter, name string, data any) {
			capturedData = data
			w.WriteHeader(http.StatusOK)
		},
	}

	h := &Handler{
		RecipeStore:  mockRecipeStore,
		TagStore:     mockTagStore,
		CommentStore: mockCommentStore,
		AuthStore:    mockAuthStore,
		Renderer:     mockRenderer,
	}

	req := httptest.NewRequest(http.MethodGet, "/recipes/1", nil)
	req.SetPathValue("id", "1")
	userInfo := &auth.UserInfo{IsLoggedIn: false}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.ViewRecipeHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if capturedData == nil {
		t.Fatal("expected data to be captured by renderer")
	}
}

func TestViewRecipeHandler_ReturnsNotFoundWhenRecipeDoesNotExist(t *testing.T) {
	mockRecipeStore := &mocks.MockRecipeStore{
		GetByIDFunc: func(id string) (models.Recipe, error) {
			return models.Recipe{}, errors.New("not found")
		},
	}

	mockRenderer := &tmocks.MockRenderer{
		RenderErrorFunc: func(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
			w.WriteHeader(statusCode)
		},
	}

	h := &Handler{
		RecipeStore: mockRecipeStore,
		Renderer:    mockRenderer,
	}

	req := httptest.NewRequest(http.MethodGet, "/recipes/999", nil)
	req.SetPathValue("id", "999")
	rec := httptest.NewRecorder()

	h.ViewRecipeHandler(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestViewRecipeHandler_ReturnsBadRequestWhenIDMissing(t *testing.T) {
	mockRenderer := &tmocks.MockRenderer{
		RenderErrorFunc: func(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
			w.WriteHeader(statusCode)
		},
	}

	h := &Handler{
		Renderer: mockRenderer,
	}

	req := httptest.NewRequest(http.MethodGet, "/recipes/", nil)
	rec := httptest.NewRecorder()

	h.ViewRecipeHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestPostCreateRecipeHandler_RedirectsToLoginWhenNotAuthenticated(t *testing.T) {
	mockAuthStore := &mocks.MockAuthStore{
		GetSessionFunc: func(sessionID string) (*store.Session, error) {
			return nil, errors.New("no session")
		},
	}

	h := &Handler{
		AuthStore: mockAuthStore,
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("title", "Test Recipe")
	writer.WriteField("ingredients", "- flour")
	writer.WriteField("instructions", "Mix")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/recipes/create", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	h.PostCreateRecipeHandler(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rec.Code)
	}

	location := rec.Header().Get("Location")
	if location != "/login" {
		t.Errorf("expected redirect to /login, got %s", location)
	}
}

func TestPostCreateRecipeHandler_CreatesRecipeWhenAuthenticated(t *testing.T) {
	mockAuthStore := &mocks.MockAuthStore{
		GetSessionFunc: func(sessionID string) (*store.Session, error) {
			return &store.Session{ID: sessionID, UserID: 1}, nil
		},
		GetUserByIDFunc: func(userID int) (*store.AuthUser, error) {
			return &store.AuthUser{ID: 1, Username: "testuser", Email: "test@test.com"}, nil
		},
	}

	var capturedRecipe models.Recipe
	mockRecipeStore := &mocks.MockRecipeStore{
		SaveFunc: func(recipe models.Recipe) (int, error) {
			capturedRecipe = recipe
			return 123, nil
		},
	}

	mockTagStore := &mocks.MockTagStore{
		SetRecipeTagsFunc: func(recipeID int, tagNames []string) error {
			return nil
		},
	}

	h := &Handler{
		AuthStore:   mockAuthStore,
		RecipeStore: mockRecipeStore,
		TagStore:    mockTagStore,
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("title", "Test Recipe")
	writer.WriteField("ingredients", "- flour\n- sugar")
	writer.WriteField("instructions", "Mix and bake")
	writer.WriteField("preptime", "10")
	writer.WriteField("cooktime", "20")
	writer.WriteField("calories", "300")
	writer.WriteField("tags", "dinner,easy")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/recipes/create", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
	rec := httptest.NewRecorder()

	h.PostCreateRecipeHandler(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rec.Code)
	}

	location := rec.Header().Get("Location")
	if location != "/recipes/123" {
		t.Errorf("expected redirect to /recipes/123, got %s", location)
	}

	if capturedRecipe.Title != "Test Recipe" {
		t.Errorf("expected title 'Test Recipe', got '%s'", capturedRecipe.Title)
	}

	if capturedRecipe.PrepTime != 10 {
		t.Errorf("expected prep time 10, got %d", capturedRecipe.PrepTime)
	}

	if capturedRecipe.AuthorID != 1 {
		t.Errorf("expected author ID 1, got %d", capturedRecipe.AuthorID)
	}
}

func TestPostCreateRecipeHandler_ReturnsBadRequestWhenFormInvalid(t *testing.T) {
	h := &Handler{}

	req := httptest.NewRequest(http.MethodPost, "/recipes/create", strings.NewReader("invalid"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	h.PostCreateRecipeHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestPostCreateRecipeHandler_ReturnsBadRequestWhenPrepTimeInvalid(t *testing.T) {
	mockAuthStore := &mocks.MockAuthStore{
		GetSessionFunc: func(sessionID string) (*store.Session, error) {
			return &store.Session{ID: sessionID, UserID: 1}, nil
		},
		GetUserByIDFunc: func(userID int) (*store.AuthUser, error) {
			return &store.AuthUser{ID: 1, Username: "testuser"}, nil
		},
	}

	h := &Handler{
		AuthStore: mockAuthStore,
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("title", "Test Recipe")
	writer.WriteField("ingredients", "- flour")
	writer.WriteField("instructions", "Mix")
	writer.WriteField("preptime", "invalid")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/recipes/create", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
	rec := httptest.NewRecorder()

	h.PostCreateRecipeHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestPostUpdateRecipeHandler_UpdatesRecipeWhenUserIsAuthor(t *testing.T) {
	mockAuthStore := &mocks.MockAuthStore{
		GetSessionFunc: func(sessionID string) (*store.Session, error) {
			return &store.Session{ID: sessionID, UserID: 1}, nil
		},
		GetUserByIDFunc: func(userID int) (*store.AuthUser, error) {
			return &store.AuthUser{ID: 1, Username: "testuser"}, nil
		},
	}

	mockRecipeStore := &mocks.MockRecipeStore{
		GetByIDFunc: func(id string) (models.Recipe, error) {
			return models.Recipe{ID: 1, AuthorID: 1, Title: "Original"}, nil
		},
		UpdateFunc: func(recipe models.Recipe) error {
			return nil
		},
	}

	mockTagStore := &mocks.MockTagStore{
		SetRecipeTagsFunc: func(recipeID int, tagNames []string) error {
			return nil
		},
	}

	h := &Handler{
		AuthStore:   mockAuthStore,
		RecipeStore: mockRecipeStore,
		TagStore:    mockTagStore,
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("title", "Updated Recipe")
	writer.WriteField("ingredients", "- updated flour")
	writer.WriteField("instructions", "Updated instructions")
	writer.WriteField("preptime", "15")
	writer.WriteField("cooktime", "25")
	writer.WriteField("calories", "350")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/recipes/1/update", body)
	req.SetPathValue("id", "1")
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
	rec := httptest.NewRecorder()

	h.PostUpdateRecipeHandler(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rec.Code)
	}

	location := rec.Header().Get("Location")
	if location != "/recipes/1" {
		t.Errorf("expected redirect to /recipes/1, got %s", location)
	}
}

func TestPostUpdateRecipeHandler_ReturnsForbiddenWhenUserIsNotAuthor(t *testing.T) {
	mockAuthStore := &mocks.MockAuthStore{
		GetSessionFunc: func(sessionID string) (*store.Session, error) {
			return &store.Session{ID: sessionID, UserID: 2}, nil
		},
		GetUserByIDFunc: func(userID int) (*store.AuthUser, error) {
			return &store.AuthUser{ID: 2, Username: "otheruser"}, nil
		},
	}

	mockRecipeStore := &mocks.MockRecipeStore{
		GetByIDFunc: func(id string) (models.Recipe, error) {
			return models.Recipe{ID: 1, AuthorID: 1, Title: "Original"}, nil
		},
	}

	h := &Handler{
		AuthStore:   mockAuthStore,
		RecipeStore: mockRecipeStore,
	}

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.WriteField("title", "Hacked Recipe")
	writer.Close()

	req := httptest.NewRequest(http.MethodPost, "/recipes/1/update", body)
	req.SetPathValue("id", "1")
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
	rec := httptest.NewRecorder()

	h.PostUpdateRecipeHandler(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestDeleteRecipeHandler_DeletesRecipeWhenUserIsAuthor(t *testing.T) {
	mockAuthStore := &mocks.MockAuthStore{
		GetSessionFunc: func(sessionID string) (*store.Session, error) {
			return &store.Session{ID: sessionID, UserID: 1}, nil
		},
		GetUserByIDFunc: func(userID int) (*store.AuthUser, error) {
			return &store.AuthUser{ID: 1, Username: "testuser"}, nil
		},
	}

	deleteCalled := false
	mockRecipeStore := &mocks.MockRecipeStore{
		GetByIDFunc: func(id string) (models.Recipe, error) {
			return models.Recipe{ID: 1, AuthorID: 1}, nil
		},
		DeleteFunc: func(id string) error {
			deleteCalled = true
			return nil
		},
	}

	h := &Handler{
		AuthStore:   mockAuthStore,
		RecipeStore: mockRecipeStore,
	}

	req := httptest.NewRequest(http.MethodDelete, "/recipes/1/delete", nil)
	req.SetPathValue("id", "1")
	req.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
	rec := httptest.NewRecorder()

	h.DeleteRecipeHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if !deleteCalled {
		t.Error("expected delete to be called")
	}
}

func TestDeleteRecipeHandler_ReturnsForbiddenWhenUserIsNotAuthor(t *testing.T) {
	mockAuthStore := &mocks.MockAuthStore{
		GetSessionFunc: func(sessionID string) (*store.Session, error) {
			return &store.Session{ID: sessionID, UserID: 2}, nil
		},
		GetUserByIDFunc: func(userID int) (*store.AuthUser, error) {
			return &store.AuthUser{ID: 2, Username: "otheruser"}, nil
		},
	}

	mockRecipeStore := &mocks.MockRecipeStore{
		GetByIDFunc: func(id string) (models.Recipe, error) {
			return models.Recipe{ID: 1, AuthorID: 1}, nil
		},
	}

	h := &Handler{
		AuthStore:   mockAuthStore,
		RecipeStore: mockRecipeStore,
	}

	req := httptest.NewRequest(http.MethodDelete, "/recipes/1/delete", nil)
	req.SetPathValue("id", "1")
	req.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
	rec := httptest.NewRecorder()

	h.DeleteRecipeHandler(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestDeleteRecipeHandler_ReturnsUnauthorizedWhenNotLoggedIn(t *testing.T) {
	mockAuthStore := &mocks.MockAuthStore{
		GetSessionFunc: func(sessionID string) (*store.Session, error) {
			return nil, errors.New("no session")
		},
	}

	h := &Handler{
		AuthStore: mockAuthStore,
	}

	req := httptest.NewRequest(http.MethodDelete, "/recipes/1/delete", nil)
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	h.DeleteRecipeHandler(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestCommentHTMXHandler_AddsCommentWhenAuthenticated(t *testing.T) {
	mockAuthStore := &mocks.MockAuthStore{
		GetSessionFunc: func(sessionID string) (*store.Session, error) {
			return &store.Session{ID: sessionID, UserID: 1}, nil
		},
		GetUserByIDFunc: func(userID int) (*store.AuthUser, error) {
			return &store.AuthUser{ID: 1, Username: "testuser"}, nil
		},
	}

	var capturedComment models.Comment
	mockCommentStore := &mocks.MockCommentStore{
		SaveFunc: func(comment models.Comment) error {
			capturedComment = comment
			return nil
		},
		GetLatestByUserAndRecipeFunc: func(userID, recipeID int) (models.Comment, error) {
			return models.Comment{ID: 1, RecipeID: recipeID, AuthorID: userID, ContentMD: "Test comment"}, nil
		},
	}

	mockRenderer := &tmocks.MockRenderer{
		RenderFragmentFunc: func(w http.ResponseWriter, name string, data any) {
			w.WriteHeader(http.StatusOK)
		},
	}

	h := &Handler{
		AuthStore:    mockAuthStore,
		CommentStore: mockCommentStore,
		Renderer:     mockRenderer,
	}

	form := url.Values{}
	form.Set("comment", "Test comment")

	req := httptest.NewRequest(http.MethodPost, "/recipes/1/comment", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", "1")
	req.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
	rec := httptest.NewRecorder()

	h.CommentHTMXHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if capturedComment.ContentMD != "Test comment" {
		t.Errorf("expected comment content 'Test comment', got '%s'", capturedComment.ContentMD)
	}

	if capturedComment.RecipeID != 1 {
		t.Errorf("expected recipe ID 1, got %d", capturedComment.RecipeID)
	}
}

func TestCommentHTMXHandler_ReturnsUnauthorizedWhenNotLoggedIn(t *testing.T) {
	mockAuthStore := &mocks.MockAuthStore{
		GetSessionFunc: func(sessionID string) (*store.Session, error) {
			return nil, errors.New("no session")
		},
	}

	h := &Handler{
		AuthStore: mockAuthStore,
	}

	form := url.Values{}
	form.Set("comment", "Test comment")

	req := httptest.NewRequest(http.MethodPost, "/recipes/1/comment", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", "1")
	rec := httptest.NewRecorder()

	h.CommentHTMXHandler(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestCommentHTMXHandler_ReturnsBadRequestWhenCommentEmpty(t *testing.T) {
	mockAuthStore := &mocks.MockAuthStore{
		GetSessionFunc: func(sessionID string) (*store.Session, error) {
			return &store.Session{ID: sessionID, UserID: 1}, nil
		},
		GetUserByIDFunc: func(userID int) (*store.AuthUser, error) {
			return &store.AuthUser{ID: 1, Username: "testuser"}, nil
		},
	}

	h := &Handler{
		AuthStore: mockAuthStore,
	}

	form := url.Values{}
	form.Set("comment", "")

	req := httptest.NewRequest(http.MethodPost, "/recipes/1/comment", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", "1")
	req.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
	rec := httptest.NewRecorder()

	h.CommentHTMXHandler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestUpdateCommentHandler_UpdatesCommentWhenUserIsAuthor(t *testing.T) {
	mockAuthStore := &mocks.MockAuthStore{
		GetSessionFunc: func(sessionID string) (*store.Session, error) {
			return &store.Session{ID: sessionID, UserID: 1}, nil
		},
		GetUserByIDFunc: func(userID int) (*store.AuthUser, error) {
			return &store.AuthUser{ID: 1, Username: "testuser"}, nil
		},
	}

	mockCommentStore := &mocks.MockCommentStore{
		GetByIDFunc: func(commentID int) (models.Comment, error) {
			return models.Comment{ID: commentID, AuthorID: 1, ContentMD: "Original"}, nil
		},
		UpdateFunc: func(commentID int, content string) error {
			return nil
		},
	}

	mockRenderer := &tmocks.MockRenderer{
		RenderFragmentFunc: func(w http.ResponseWriter, name string, data any) {
			w.WriteHeader(http.StatusOK)
		},
	}

	h := &Handler{
		AuthStore:    mockAuthStore,
		CommentStore: mockCommentStore,
		Renderer:     mockRenderer,
	}

	form := url.Values{}
	form.Set("comment", "Updated comment")

	req := httptest.NewRequest(http.MethodPut, "/comments/1", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", "1")
	req.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
	rec := httptest.NewRecorder()

	h.UpdateCommentHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestUpdateCommentHandler_ReturnsForbiddenWhenUserIsNotAuthor(t *testing.T) {
	mockAuthStore := &mocks.MockAuthStore{
		GetSessionFunc: func(sessionID string) (*store.Session, error) {
			return &store.Session{ID: sessionID, UserID: 2}, nil
		},
		GetUserByIDFunc: func(userID int) (*store.AuthUser, error) {
			return &store.AuthUser{ID: 2, Username: "otheruser"}, nil
		},
	}

	mockCommentStore := &mocks.MockCommentStore{
		GetByIDFunc: func(commentID int) (models.Comment, error) {
			return models.Comment{ID: commentID, AuthorID: 1, ContentMD: "Original"}, nil
		},
	}

	h := &Handler{
		AuthStore:    mockAuthStore,
		CommentStore: mockCommentStore,
	}

	form := url.Values{}
	form.Set("comment", "Hacked comment")

	req := httptest.NewRequest(http.MethodPut, "/comments/1", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetPathValue("id", "1")
	req.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
	rec := httptest.NewRecorder()

	h.UpdateCommentHandler(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestDeleteCommentHandler_DeletesCommentWhenUserIsAuthor(t *testing.T) {
	mockAuthStore := &mocks.MockAuthStore{
		GetSessionFunc: func(sessionID string) (*store.Session, error) {
			return &store.Session{ID: sessionID, UserID: 1}, nil
		},
		GetUserByIDFunc: func(userID int) (*store.AuthUser, error) {
			return &store.AuthUser{ID: 1, Username: "testuser"}, nil
		},
	}

	deleteCalled := false
	mockCommentStore := &mocks.MockCommentStore{
		GetByIDFunc: func(commentID int) (models.Comment, error) {
			return models.Comment{ID: commentID, AuthorID: 1}, nil
		},
		DeleteFunc: func(commentID int) error {
			deleteCalled = true
			return nil
		},
	}

	h := &Handler{
		AuthStore:    mockAuthStore,
		CommentStore: mockCommentStore,
	}

	req := httptest.NewRequest(http.MethodDelete, "/comments/1", nil)
	req.SetPathValue("id", "1")
	req.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
	rec := httptest.NewRecorder()

	h.DeleteCommentHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if !deleteCalled {
		t.Error("expected delete to be called")
	}
}

func TestDeleteCommentHandler_ReturnsForbiddenWhenUserIsNotAuthor(t *testing.T) {
	mockAuthStore := &mocks.MockAuthStore{
		GetSessionFunc: func(sessionID string) (*store.Session, error) {
			return &store.Session{ID: sessionID, UserID: 2}, nil
		},
		GetUserByIDFunc: func(userID int) (*store.AuthUser, error) {
			return &store.AuthUser{ID: 2, Username: "otheruser"}, nil
		},
	}

	mockCommentStore := &mocks.MockCommentStore{
		GetByIDFunc: func(commentID int) (models.Comment, error) {
			return models.Comment{ID: commentID, AuthorID: 1}, nil
		},
	}

	h := &Handler{
		AuthStore:    mockAuthStore,
		CommentStore: mockCommentStore,
	}

	req := httptest.NewRequest(http.MethodDelete, "/comments/1", nil)
	req.SetPathValue("id", "1")
	req.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
	rec := httptest.NewRecorder()

	h.DeleteCommentHandler(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestRandomRecipeHandler_RedirectsToRandomRecipe(t *testing.T) {
	mockRecipeStore := &mocks.MockRecipeStore{
		GetRandomIDFunc: func() (int, error) {
			return 42, nil
		},
	}

	h := &Handler{
		RecipeStore: mockRecipeStore,
	}

	req := httptest.NewRequest(http.MethodGet, "/recipes/random", nil)
	rec := httptest.NewRecorder()

	h.RandomRecipeHandler(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rec.Code)
	}

	location := rec.Header().Get("Location")
	if location != "/recipes/42" {
		t.Errorf("expected redirect to /recipes/42, got %s", location)
	}
}

func TestGetUpdateRecipeHandler_RendersPageWhenUserIsAuthor(t *testing.T) {
	mockAuthStore := &mocks.MockAuthStore{
		GetSessionFunc: func(sessionID string) (*store.Session, error) {
			return &store.Session{ID: sessionID, UserID: 1}, nil
		},
		GetUserByIDFunc: func(userID int) (*store.AuthUser, error) {
			return &store.AuthUser{ID: 1, Username: "testuser"}, nil
		},
	}

	mockRecipeStore := &mocks.MockRecipeStore{
		GetByIDFunc: func(id string) (models.Recipe, error) {
			return models.Recipe{ID: 1, AuthorID: 1, Title: "My Recipe"}, nil
		},
	}

	mockTagStore := &mocks.MockTagStore{
		GetByRecipeIDFunc: func(recipeID int) ([]models.Tag, error) {
			return []models.Tag{}, nil
		},
	}

	var capturedData any
	mockRenderer := &tmocks.MockRenderer{
		RenderPageFunc: func(w http.ResponseWriter, name string, data any) {
			capturedData = data
			w.WriteHeader(http.StatusOK)
		},
	}

	h := &Handler{
		AuthStore:   mockAuthStore,
		RecipeStore: mockRecipeStore,
		TagStore:    mockTagStore,
		Renderer:    mockRenderer,
	}

	req := httptest.NewRequest(http.MethodGet, "/recipes/1/update", nil)
	req.SetPathValue("id", "1")
	req.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
	userInfo := &auth.UserInfo{IsLoggedIn: true, Username: "testuser"}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.GetUpdateRecipeHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if capturedData == nil {
		t.Fatal("expected data to be captured by renderer")
	}
}

func TestGetUpdateRecipeHandler_ReturnsForbiddenWhenUserIsNotAuthor(t *testing.T) {
	mockAuthStore := &mocks.MockAuthStore{
		GetSessionFunc: func(sessionID string) (*store.Session, error) {
			return &store.Session{ID: sessionID, UserID: 2}, nil
		},
		GetUserByIDFunc: func(userID int) (*store.AuthUser, error) {
			return &store.AuthUser{ID: 2, Username: "otheruser"}, nil
		},
	}

	mockRecipeStore := &mocks.MockRecipeStore{
		GetByIDFunc: func(id string) (models.Recipe, error) {
			return models.Recipe{ID: 1, AuthorID: 1, Title: "Someone Else's Recipe"}, nil
		},
	}

	mockRenderer := &tmocks.MockRenderer{
		RenderErrorFunc: func(w http.ResponseWriter, r *http.Request, statusCode int, message string) {
			w.WriteHeader(statusCode)
		},
	}

	h := &Handler{
		AuthStore:   mockAuthStore,
		RecipeStore: mockRecipeStore,
		Renderer:    mockRenderer,
	}

	req := httptest.NewRequest(http.MethodGet, "/recipes/1/update", nil)
	req.SetPathValue("id", "1")
	req.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
	userInfo := &auth.UserInfo{IsLoggedIn: true, Username: "otheruser"}
	req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
	rec := httptest.NewRecorder()

	h.GetUpdateRecipeHandler(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected status %d, got %d", http.StatusForbidden, rec.Code)
	}
}

func TestRandomRecipeHandler_RedirectsToListWhenNoRecipes(t *testing.T) {
	mockRecipeStore := &mocks.MockRecipeStore{
		GetRandomIDFunc: func() (int, error) {
			return 0, errors.New("no recipes")
		},
	}

	h := &Handler{
		RecipeStore: mockRecipeStore,
	}

	req := httptest.NewRequest(http.MethodGet, "/recipes/random", nil)
	rec := httptest.NewRecorder()

	h.RandomRecipeHandler(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Errorf("expected status %d, got %d", http.StatusSeeOther, rec.Code)
	}

	location := rec.Header().Get("Location")
	if location != "/recipes" {
		t.Errorf("expected redirect to /recipes, got %s", location)
	}
}

func TestFilterRecipesHTMXHandler_FiltersRecipesBySearchQuery(t *testing.T) {
	mockAuthStore := &mocks.MockAuthStore{
		GetSessionFunc: func(sessionID string) (*store.Session, error) {
			return nil, errors.New("no session")
		},
	}

	var capturedParams models.FilterParams
	mockRecipeStore := &mocks.MockRecipeStore{
		GetFilteredFunc: func(params models.FilterParams) ([]models.Recipe, error) {
			capturedParams = params
			return []models.Recipe{{ID: 1, Title: "Pasta"}}, nil
		},
		CountFilteredFunc: func(params models.FilterParams) (int, error) {
			return 1, nil
		},
	}

	mockTagStore := &mocks.MockTagStore{
		GetForRecipesFunc: func(recipeIDs []int) (map[int][]models.Tag, error) {
			return map[int][]models.Tag{}, nil
		},
	}

	mockRenderer := &tmocks.MockRenderer{
		RenderFragmentFunc: func(w http.ResponseWriter, name string, data any) {
			w.WriteHeader(http.StatusOK)
		},
	}

	h := &Handler{
		AuthStore:   mockAuthStore,
		RecipeStore: mockRecipeStore,
		TagStore:    mockTagStore,
		Renderer:    mockRenderer,
	}

	form := url.Values{}
	form.Set("search", "pasta")

	req := httptest.NewRequest(http.MethodPost, "/recipes/filter", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	h.FilterRecipesHTMXHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if capturedParams.Search != "pasta" {
		t.Errorf("expected search 'pasta', got '%s'", capturedParams.Search)
	}
}

func TestFilterRecipesHTMXHandler_FiltersRecipesByTags(t *testing.T) {
	mockAuthStore := &mocks.MockAuthStore{
		GetSessionFunc: func(sessionID string) (*store.Session, error) {
			return nil, errors.New("no session")
		},
	}

	var capturedParams models.FilterParams
	mockRecipeStore := &mocks.MockRecipeStore{
		GetFilteredFunc: func(params models.FilterParams) ([]models.Recipe, error) {
			capturedParams = params
			return []models.Recipe{}, nil
		},
		CountFilteredFunc: func(params models.FilterParams) (int, error) {
			return 0, nil
		},
	}

	mockTagStore := &mocks.MockTagStore{
		GetForRecipesFunc: func(recipeIDs []int) (map[int][]models.Tag, error) {
			return map[int][]models.Tag{}, nil
		},
	}

	mockRenderer := &tmocks.MockRenderer{
		RenderFragmentFunc: func(w http.ResponseWriter, name string, data any) {
			w.WriteHeader(http.StatusOK)
		},
	}

	h := &Handler{
		AuthStore:   mockAuthStore,
		RecipeStore: mockRecipeStore,
		TagStore:    mockTagStore,
		Renderer:    mockRenderer,
	}

	form := url.Values{}
	form.Set("tags", "dinner,easy,quick")

	req := httptest.NewRequest(http.MethodPost, "/recipes/filter", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	h.FilterRecipesHTMXHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if len(capturedParams.Tags) != 3 {
		t.Errorf("expected 3 tags, got %d", len(capturedParams.Tags))
	}
}

func TestFilterRecipesHTMXHandler_FiltersRecipesByNumericValues(t *testing.T) {
	mockAuthStore := &mocks.MockAuthStore{
		GetSessionFunc: func(sessionID string) (*store.Session, error) {
			return nil, errors.New("no session")
		},
	}

	var capturedParams models.FilterParams
	mockRecipeStore := &mocks.MockRecipeStore{
		GetFilteredFunc: func(params models.FilterParams) ([]models.Recipe, error) {
			capturedParams = params
			return []models.Recipe{}, nil
		},
		CountFilteredFunc: func(params models.FilterParams) (int, error) {
			return 0, nil
		},
	}

	mockTagStore := &mocks.MockTagStore{
		GetForRecipesFunc: func(recipeIDs []int) (map[int][]models.Tag, error) {
			return map[int][]models.Tag{}, nil
		},
	}

	mockRenderer := &tmocks.MockRenderer{
		RenderFragmentFunc: func(w http.ResponseWriter, name string, data any) {
			w.WriteHeader(http.StatusOK)
		},
	}

	h := &Handler{
		AuthStore:   mockAuthStore,
		RecipeStore: mockRecipeStore,
		TagStore:    mockTagStore,
		Renderer:    mockRenderer,
	}

	form := url.Values{}
	form.Set("calories_value", "500")
	form.Set("calories_op", "lte")
	form.Set("prep_time_value", "30")
	form.Set("prep_time_op", "lte")

	req := httptest.NewRequest(http.MethodPost, "/recipes/filter", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	h.FilterRecipesHTMXHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if capturedParams.CaloriesValue != 500 {
		t.Errorf("expected calories value 500, got %d", capturedParams.CaloriesValue)
	}

	if capturedParams.CaloriesOp != "lte" {
		t.Errorf("expected calories op 'lte', got '%s'", capturedParams.CaloriesOp)
	}

	if capturedParams.PrepTimeValue != 30 {
		t.Errorf("expected prep time value 30, got %d", capturedParams.PrepTimeValue)
	}
}

func TestFilterRecipesHTMXHandler_HandlesPagination(t *testing.T) {
	mockAuthStore := &mocks.MockAuthStore{
		GetSessionFunc: func(sessionID string) (*store.Session, error) {
			return nil, errors.New("no session")
		},
	}

	var capturedParams models.FilterParams
	mockRecipeStore := &mocks.MockRecipeStore{
		GetFilteredFunc: func(params models.FilterParams) ([]models.Recipe, error) {
			capturedParams = params
			return []models.Recipe{}, nil
		},
		CountFilteredFunc: func(params models.FilterParams) (int, error) {
			return 100, nil
		},
	}

	mockTagStore := &mocks.MockTagStore{
		GetForRecipesFunc: func(recipeIDs []int) (map[int][]models.Tag, error) {
			return map[int][]models.Tag{}, nil
		},
	}

	mockRenderer := &tmocks.MockRenderer{
		RenderFragmentFunc: func(w http.ResponseWriter, name string, data any) {
			w.WriteHeader(http.StatusOK)
		},
	}

	h := &Handler{
		AuthStore:   mockAuthStore,
		RecipeStore: mockRecipeStore,
		TagStore:    mockTagStore,
		Renderer:    mockRenderer,
	}

	form := url.Values{}
	form.Set("offset", "20")

	req := httptest.NewRequest(http.MethodPost, "/recipes/filter", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	h.FilterRecipesHTMXHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	if capturedParams.Offset != 20 {
		t.Errorf("expected offset 20, got %d", capturedParams.Offset)
	}

	if capturedParams.Limit != models.DefaultPageSize {
		t.Errorf("expected limit %d, got %d", models.DefaultPageSize, capturedParams.Limit)
	}
}
