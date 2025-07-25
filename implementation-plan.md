# Implementation Plan (Draft)

## 1. Project Structure
- Organize code into packages: main, handlers, models, templates, static, db, auth, utils.
- Use Go modules for dependency management.
- Store templates and static files in dedicated folders.

## 2. Database Schema
- Users: id, username, password_hash (for mock auth), created_at
- Recipes: id, title, ingredients_md, instructions_md, prep_time, cook_time, calories, author_id, image (bytea), parent_id (nullable), created_at, updated_at
- Labels: id, name
- RecipeLabels: recipe_id, label_id
- Comments: id, recipe_id, author_id, content_md, created_at, updated_at
- ProposedChanges: id, recipe_id, proposer_id, title, ingredients_md, instructions_md, prep_time, cook_time, calories, image, created_at, status (pending/accepted/rejected)

## 3. Authentication
- Implement a local mock authentication provider (session-based, username/password, no registration UI).
- Middleware to enforce authentication for actions (create/edit/delete/propose).

## 4. Core Features
- CRUD for recipes (only author can edit/delete).
- CRUD for comments (only author can edit/delete).
- CRUD for labels (user-created, assign to recipes).
- Image upload for recipes (store in DB as bytea).
- Markdown rendering for ingredients, instructions, and comments.
- Propose changes to recipes; author can accept/reject; rejected changes can be published as new recipes with parent link.

## 5. Search & Pagination
- Implement case-insensitive, full-text, and fuzzy search for recipes (use PostgreSQL features and/or Go fuzzy search libs if needed).
- Paginate recipe overview/search page (default 20 per page, configurable).
- Filter recipes by label.

## 6. Frontend
- Use Go templates for server-side rendered HTML.
- Use htmx for dynamic content (e.g., updating recipe list, submitting forms, accepting/rejecting proposals).
- Basic, clean UI (no advanced styling for MVP).

## 7. Dockerization
- Dockerfile for Go app, using environment variables for DB connection.
- Docker Compose file for local dev (Go app + Postgres).

## 8. Extensibility
- Structure code to allow swapping auth provider and HTTP framework in the future.
- Make pagination and search options configurable.

## 9. Testing & Documentation
- Add basic tests for core logic (models, handlers).
- Document API endpoints, DB schema, and setup in README.

---

This plan is a draft and can be refined further based on feedback or as implementation progresses.
