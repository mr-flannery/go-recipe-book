# Architecture: Store/Repository Pattern

This document describes the store/repository pattern used in this codebase for data access.

## Overview

The application uses a layered architecture that separates HTTP handling from data access:

```
┌─────────────────────────────────────────────────────────────────┐
│                         main.go                                 │
│   - Initializes DB pool                                         │
│   - Creates store implementations                               │
│   - Injects stores into Handler                                 │
│   - Wires up routes                                             │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                      handlers.Handler                           │
│   - Holds references to all stores (as interfaces)              │
│   - All handler methods are receivers on this struct            │
└─────────────────────────────────────────────────────────────────┘
                              │
            ┌─────────────────┼─────────────────┐
            ▼                 ▼                 ▼
┌───────────────────┐ ┌───────────────┐ ┌───────────────┐
│ store.RecipeStore │ │store.TagStore │ │store.UserStore│  (interfaces)
└───────────────────┘ └───────────────┘ └───────────────┘
            │                 │                 │
            ▼                 ▼                 ▼
┌────────────────────┐┌────────────────┐┌────────────────┐
│postgres.RecipeStore││postgres.TagStore││postgres.UserStore│ (implementations)
└────────────────────┘└────────────────┘└────────────────┘
                              │
                              ▼
                    ┌─────────────────┐
                    │   *sql.DB pool  │
                    └─────────────────┘
```

## Directory Structure

```
src/
├── store/
│   ├── interfaces.go       # Store interface definitions
│   └── postgres/           # PostgreSQL implementations
│       ├── recipe.go
│       ├── tag.go
│       ├── user_tag.go
│       ├── comment.go
│       └── user.go
├── handlers/
│   ├── handler.go          # Handler struct with dependencies
│   ├── recipe.go           # Recipe-related handlers
│   ├── tags.go             # Tag-related handlers
│   ├── auth.go             # Authentication handlers
│   └── api.go              # API handlers
├── models/
│   └── models.go           # Data structures only (no SQL)
└── db/
    └── db.go               # Connection pool management
```

## Store Interfaces

Interfaces define the contract for data access without implementation details:

```go
// src/store/interfaces.go
type RecipeStore interface {
    Save(recipe models.Recipe) (int, error)
    GetByID(id string) (models.Recipe, error)
    Update(recipe models.Recipe) error
    Delete(id string) error
    GetAll() ([]models.Recipe, error)
    GetFiltered(params models.FilterParams) ([]models.Recipe, error)
    GetRandomID() (int, error)
}

type TagStore interface {
    GetOrCreate(name string) (models.Tag, error)
    Search(query string) ([]models.Tag, error)
    GetByRecipeID(recipeID int) ([]models.Tag, error)
    GetForRecipes(recipeIDs []int) (map[int][]models.Tag, error)
    AddToRecipe(recipeID, tagID int) error
    RemoveFromRecipe(recipeID, tagID int) error
    SetRecipeTags(recipeID int, tagNames []string) error
}

type UserTagStore interface {
    GetOrCreate(userID, recipeID int, name string) (models.UserTag, error)
    Search(userID int, query string) ([]string, error)
    GetByRecipeID(userID, recipeID int) ([]models.UserTag, error)
    Remove(userID, tagID int) error
}

type CommentStore interface {
    GetByRecipeID(recipeID string) ([]models.Comment, error)
    Save(comment models.Comment) error
    GetLatestByUserAndRecipe(userID, recipeID int) (models.Comment, error)
}

type UserStore interface {
    GetUsernameByID(userID int) (string, error)
}
```

## PostgreSQL Implementations

Each store interface has a corresponding PostgreSQL implementation:

```go
// src/store/postgres/recipe.go
type RecipeStore struct {
    db *sql.DB
}

func NewRecipeStore(db *sql.DB) *RecipeStore {
    return &RecipeStore{db: db}
}

func (s *RecipeStore) Save(recipe models.Recipe) (int, error) {
    query := `INSERT INTO recipes (...) VALUES ($1, $2, ...) RETURNING id`
    var id int
    err := s.db.QueryRow(query, recipe.Title, ...).Scan(&id)
    return id, err
}

func (s *RecipeStore) GetByID(id string) (models.Recipe, error) {
    var recipe models.Recipe
    err := s.db.QueryRow("SELECT ... FROM recipes WHERE id = $1", id).
        Scan(&recipe.ID, &recipe.Title, ...)
    return recipe, err
}
```

## Handler Dependency Injection

The Handler struct holds all store dependencies as interfaces:

```go
// src/handlers/handler.go
type Handler struct {
    DB           *sql.DB
    RecipeStore  store.RecipeStore
    TagStore     store.TagStore
    UserTagStore store.UserTagStore
    CommentStore store.CommentStore
    UserStore    store.UserStore
}

func NewHandler(
    db *sql.DB,
    recipeStore store.RecipeStore,
    tagStore store.TagStore,
    userTagStore store.UserTagStore,
    commentStore store.CommentStore,
    userStore store.UserStore,
) *Handler {
    return &Handler{
        DB:           db,
        RecipeStore:  recipeStore,
        TagStore:     tagStore,
        UserTagStore: userTagStore,
        CommentStore: commentStore,
        UserStore:    userStore,
    }
}
```

Handler methods use the injected stores:

```go
// src/handlers/recipe.go
func (h *Handler) ListRecipesHandler(w http.ResponseWriter, r *http.Request) {
    recipes, err := h.RecipeStore.GetAll()
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
    // ... render template
}
```

## Wiring in main.go

All dependencies are created and wired together at application startup:

```go
// src/main.go
func main() {
    // Initialize single connection pool
    database, err := db.InitPool()
    defer db.ClosePool()

    // Create concrete store implementations
    recipeStore := postgres.NewRecipeStore(database)
    tagStore := postgres.NewTagStore(database)
    userTagStore := postgres.NewUserTagStore(database)
    commentStore := postgres.NewCommentStore(database)
    userStore := postgres.NewUserStore(database)

    // Inject into Handler
    h := handlers.NewHandler(database, recipeStore, tagStore,
                             userTagStore, commentStore, userStore)

    // Wire routes
    mux.Handle("GET /recipes", userContext(http.HandlerFunc(h.ListRecipesHandler)))
    mux.Handle("POST /recipes/create",
        userContext(requireAuth(http.HandlerFunc(h.PostCreateRecipeHandler))))
}
```

## Connection Pool

The database package manages a single shared connection pool:

```go
// src/db/db.go
var pool *sql.DB

func InitPool() (*sql.DB, error) {
    if pool != nil {
        return pool, nil
    }

    db, err := sql.Open("postgres", connectionString)
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(25)
    pool = db
    return pool, nil
}

func ClosePool() error {
    if pool != nil {
        return pool.Close()
    }
    return nil
}
```

## Benefits

1. **Testability**: Create mock stores that return test data without a database
2. **Separation of concerns**: HTTP handling and data access are cleanly separated
3. **Swappable implementations**: Could add SQLite or other backends
4. **Single source of truth**: One place for each SQL query
5. **Dependency injection**: Clear visibility into handler dependencies

## Adding a New Store

1. Define the interface in `src/store/interfaces.go`
2. Create implementation in `src/store/postgres/newstore.go`
3. Add field to Handler struct in `src/handlers/handler.go`
4. Update `NewHandler` function
5. Create store instance in `main.go` and pass to `NewHandler`

## Testing with Mocks

To test handlers without a database, create mock implementations:

```go
// src/store/mock/recipe.go
type RecipeStore struct {
    Recipes []models.Recipe
    SaveFn  func(models.Recipe) (int, error)
}

func (m *RecipeStore) GetAll() ([]models.Recipe, error) {
    return m.Recipes, nil
}

func (m *RecipeStore) Save(recipe models.Recipe) (int, error) {
    if m.SaveFn != nil {
        return m.SaveFn(recipe)
    }
    return 1, nil
}
```

Then in tests:

```go
func TestListRecipesHandler(t *testing.T) {
    mockRecipeStore := &mock.RecipeStore{
        Recipes: []models.Recipe{{ID: 1, Title: "Test"}},
    }
    mockTagStore := &mock.TagStore{}

    h := handlers.NewHandler(nil, mockRecipeStore, mockTagStore, nil, nil, nil)

    req := httptest.NewRequest("GET", "/recipes", nil)
    w := httptest.NewRecorder()

    h.ListRecipesHandler(w, req)

    // assert response
}
```
