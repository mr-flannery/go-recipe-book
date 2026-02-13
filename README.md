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

### Live Reloading with Air
For development with automatic server restarts on file changes:

```sh
make dev
```

This uses [Air](https://github.com/air-verse/air) for live reloading. Configuration is in `.air.toml`.

### Project Structure
- Go code is in `src/` directory
- Templates in `src/templates/`
- Database migrations in `src/db/migrations/`

### Architecture
See [docs/architecture.md](docs/architecture.md) for details on the store/repository pattern used for data access.

### Manual Development
For development without live reloading:
```sh
cd src
go run main.go
```
