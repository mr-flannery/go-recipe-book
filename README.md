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

### Live Reloading with Reflex
For development with automatic server restarts on file changes:

```sh
# Using the dev script
./dev.sh

# Or directly with Reflex
reflex -r '\.go$|\.gohtml$|\.html$|\.tmpl$|\.tpl$' -s -- sh -c 'cd src && go run main.go'

# Or using VSCode launch configuration
# Select "Launch Server with Live Reload (Reflex)" from the debug panel
```

### Project Structure
- Go code is in `src/` directory
- Templates in `src/templates/`
- Database migrations in `src/db/migrations/`
- Live reloading with Reflex watches Go and template files

### Manual Development
For development without live reloading:
```sh
cd src
go run main.go
```

## Users (Mock Auth)
- Predefined users: `alice` (password: `password1`), `bob` (password: `password2`)

## TODO
- Implement handlers, templates, and core features as described in `requirements.md` and `implementation-plan.md`.
