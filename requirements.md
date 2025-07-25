# Requirements

## Core Features
- A recipe must have a title, a list of ingredients, and instructions, all written in free text.
- Users should be able to use markdown syntax to write the ingredients and instructions.
- A recipe can have labels.
- A label is a user-created entity that contains arbitrary text.
- Labels and recipes have a many-to-many relationship.
- A recipe can also have preparation time, cook time, and estimated calorie count.
- Users should be able to submit new recipes, update existing ones, and delete them again.
- Users should be able to leave comments on recipes.
- There should be an overview page that lists all recipes, on which users should be able to search for recipes and filter them by label.
- Recipes can have images, which are stored in the database.

## User & Auth
- There should be user authentication, implemented with a local mock provider for now (to be swapped out later).
- Each recipe has an author (the user who created it).
- Comments are associated with users.
- Users can only edit or delete their own recipes and comments.
- Users can propose changes to a recipe. The author can accept or reject these changes.
- If changes are rejected, the proposer can create a new recipe as an alternative version, with the original as its parent.

## Search & Pagination
- Search should be case-insensitive and support both full-text and fuzzy search (e.g., 'spag carb' finds 'spaghetti carbonara').
- The recipe overview/search page should be paginated, with a default limit of 20 per page (configurable in the future, possibly by users).

## Architecture
- Backend is written in Go.
- Uses PostgreSQL as the database.
- Frontend uses strictly server-side rendered HTML pages with Go's built-in templating.
- htmx is used for dynamic content (e.g., updating recipe lists, submitting forms without full page reloads).
- The application should be runnable via a Docker container.
- The Docker image for the Go app should take PostgreSQL connection details as environment variables.
- For local development, a dockerized Postgres instance will be used, but the DB should be swappable for a cloud instance in the future.
- Use only the Go standard library for HTTP, but structure code so a framework can be adopted later.

## Other
- No special requirements for markdown rendering for now.
- No user management (registration, roles) for now.
