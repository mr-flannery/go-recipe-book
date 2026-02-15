# Testing Guide

This document describes the testing strategy and how to run tests for the recipe book application.

## Test Types

### Unit Tests
Fast tests that don't require external dependencies. These test individual functions and components in isolation.

**Run unit tests:**
```bash
go test ./... -short -count=1
```

### Integration Tests
Tests that require a PostgreSQL database. These use [testcontainers-go](https://golang.testcontainers.org/) to spin up ephemeral PostgreSQL containers.

**Prerequisites:**
- Docker must be running

**Run integration tests:**
```bash
go test ./src/store/postgres/... -v -count=1
```

### All Tests
Run both unit and integration tests:
```bash
go test ./... -count=1
```

## Test Organization

```
src/
├── auth/
│   └── *_test.go           # Auth utility unit tests
├── config/
│   └── config_test.go      # Configuration parsing tests
├── handlers/
│   ├── auth_test.go        # Auth handler tests (~30 tests)
│   └── recipe_test.go      # Recipe handler tests (~25 tests)
├── store/postgres/
│   ├── auth_test.go        # Auth store integration tests (~25 tests)
│   ├── comment_test.go     # Comment store integration tests (~8 tests)
│   ├── recipe_test.go      # Recipe store integration tests (~12 tests)
│   ├── tag_test.go         # Tag store integration tests (~12 tests)
│   ├── user_tag_test.go    # User tag store integration tests (~10 tests)
│   └── user_test.go        # User store integration tests (~2 tests)
└── testutil/
    └── database.go         # Test database helper with seed functions
```

## Test Utilities

### Test Database Helper (`src/testutil/database.go`)

Provides a `TestDatabase` struct with:

- `SetupTestDatabase(t)` - Creates a PostgreSQL container and runs migrations
- `Cleanup(t)` - Stops and removes the container
- `TruncateTables(t, tables...)` - Clears specific tables
- `ResetAllTables(t)` - Clears all tables

**Seed helpers:**
- `SeedUser(t, username, email, passwordHash, isAdmin)` - Creates a test user
- `SeedRecipe(t, title, ingredients, instructions, authorID)` - Creates a test recipe
- `SeedTag(t, name)` - Creates a test tag
- `SeedRecipeTag(t, recipeID, tagID)` - Associates tag with recipe
- `SeedUserTag(t, userID, recipeID, name)` - Creates a user-specific tag
- `SeedComment(t, recipeID, authorID, content)` - Creates a test comment
- `SeedSession(t, sessionID, userID, expiresAt)` - Creates a test session

### Skipping Integration Tests

Integration tests should call `testutil.SkipIfShort(t)` at the start to skip when running with `-short`:

```go
func TestSomething_IntegrationTest(t *testing.T) {
    testutil.SkipIfShort(t)
    // ... test code
}
```

## Test Naming Convention

Tests follow the pattern: `TestThingUnderTest_BehavesLikeXWhenY`

Examples:
- `TestAuthenticate_ReturnsUserWhenCredentialsAreValid`
- `TestHashPassword_ReturnsErrorWhenPasswordIsWeak`
- `TestRecipeStore_GetByID_ReturnsErrorWhenNotFound`

## Writing Handler Tests

Handler tests use `httptest` and mock stores:

```go
func TestHandler_ExpectedBehavior(t *testing.T) {
    store := &mockStore{}
    handler := NewHandler(store)
    
    req := httptest.NewRequest("GET", "/path", nil)
    w := httptest.NewRecorder()
    
    handler.ServeHTTP(w, req)
    
    if w.Code != http.StatusOK {
        t.Errorf("expected status 200, got %d", w.Code)
    }
}
```

## Writing Integration Tests

Integration tests use testcontainers:

```go
func TestStore_ExpectedBehavior(t *testing.T) {
    testutil.SkipIfShort(t)
    
    testDB := testutil.SetupTestDatabase(t)
    defer testDB.Cleanup(t)
    
    // Seed test data
    userID := testDB.SeedUser(t, "testuser", "test@example.com", "hash", false)
    
    // Create store and test
    store := NewStore(testDB.DB)
    result, err := store.SomeMethod(userID)
    
    // Assertions
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }
}
```

## CI/CD

Tests run automatically on push via GitHub Actions (`.github/workflows/test.yml`):

1. **Unit tests** - Run on every push
2. **Integration tests** - Run with Docker service
3. **Coverage** - Generated and uploaded as artifacts

## Makefile Targets

```bash
make test-unit        # Run unit tests only
make test-integration # Run integration tests (requires Docker)
make test-coverage    # Generate coverage report
```

## Notes

- Password validation requires strong passwords (12+ chars, mixed case, numbers, symbols)
- Test passwords should use something like `"Correct#Pass1"` not `"password123"`
