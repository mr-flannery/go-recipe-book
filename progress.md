# Progress Log

## Date: 2025-07-25

### Project Setup Completed
- Created initial project structure: `cmd/`, `internal/`, `db/`, `auth/`, `templates/`, `static/`.
- Added `go.mod` with PostgreSQL driver dependency.
- Added initial `main.go` with basic HTTP server and placeholder home handler.
- Created database schema in `db/schema.sql`.
- Added Dockerfile and docker-compose for Go app and Postgres.
- Created initial models in `internal/models/models.go`.
- Added mock authentication logic in `auth/auth.go`.
- Created a basic home page handler and template.
- Wrote a README with setup instructions and mock user info.


### Implementation Progress (continued)
- Added login and logout handlers (inline in main.go for now, also available in internal/handlers/auth.go).
- Home page now renders the template as intended.
- Auth routes `/login` and `/logout` are functional with mock users (alice/password1, bob/password2).
- Next lint error to fix: move inline handlers to their own package and wire up imports properly.

### Next Steps
- Refactor handlers to use the internal/handlers package directly.
- Implement session checks for authenticated routes.
- Implement recipe CRUD (create, read, update, delete) and main recipe listing page.
- Connect handlers to templates and database.
- Add htmx-powered dynamic content for recipe list and forms.

**Paused here for user questions/clarifications.**

## Date: 2025-08-01

### Recent Progress
- **Database Migrations**:
  - Created `20250801_initial_schema.up.sql` to define the initial database schema.
  - Created `20250801_insert_default_users.up.sql` to insert default users (`alice` and `bob`).
  - Created `20250801_everything.down.sql` to revert all changes made by the initial schema and default users migrations.

- **Authentication**:
  - Implemented `auth.GetUserIDByUsername` to fetch user IDs from the database based on usernames.

- **Recipe Handlers**:
  - Updated `CreateRecipeHandler` to dynamically set the `AuthorID` based on the logged-in user.
  - Updated `ListRecipesHandler` to fetch recipes from the database using the new `models.GetAllRecipes` function.

- **Models**:
  - Added `models.GetAllRecipes` to retrieve all recipes from the database.

### Next Steps
- add browser tests for existing functionality
- refactor!!!
- Test the database migrations and ensure they work as expected.
- Verify the updated handlers (`CreateRecipeHandler` and `ListRecipesHandler`) with real data.
- Implement additional CRUD operations for recipes (update and delete).
- Add unit tests for the new database functions and handlers.
- Add htmx-powered dynamic content for recipe list and forms.
