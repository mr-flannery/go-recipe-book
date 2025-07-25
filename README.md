# agent-coding-recipe-book

## Setup

1. Copy `db/schema.sql` to your Postgres instance and run it to create the tables.
2. Build and run the app with Docker Compose:

```sh
docker-compose up --build
```

The app will be available at http://localhost:8080

## Environment Variables
- `DB_HOST` (default: db)
- `DB_PORT` (default: 5432)
- `DB_USER` (default: recipeuser)
- `DB_PASSWORD` (default: recipepass)
- `DB_NAME` (default: recipebook)
- `PORT` (default: 8080)

## Development
- Go code is in `cmd/`, `internal/`, `db/`, `auth/`
- Templates in `templates/`, static files in `static/`
- Database schema in `db/schema.sql`

## Users (Mock Auth)
- Predefined users: `alice` (password: `password1`), `bob` (password: `password2`)

## TODO
- Implement handlers, templates, and core features as described in `requirements.md` and `implementation-plan.md`.
