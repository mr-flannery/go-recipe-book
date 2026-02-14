- write tests
  - primarily integration tests for the handler functions
  - which means I gotta figure out how to mock the DB/do DI or something similar

```
What Wasn't Tested (and Why)
- src/handlers/home.go, src/handlers/recipe.go: These handlers render HTML templates, making them harder to unit test without a full template mocking strategy. They would be better suited for integration tests.
- src/store/postgres/: Database stores require go-sqlmock or a real database for proper testing. This would be integration testing rather than unit testing.
- src/config/: Depends on file system and global state (singleton pattern)
- src/mail/: Requires mocking external Maileroo client
- src/db/: Database connection/migration code - better suited for integration tests
```

- browser tests for tagging and commenting are missing

- UX improvements!!
  - add pagination/dynamic fetching on scroll for the recipes page
  - the admin registration UX could be improved, by listing denied and approved registrations, and also by make the experience more dynamic, showing success messages or something like this.
  - filtering by authored by me should be possible
  - when adding a tag, the focus isn't on the input component any more, but it should

- error pages suck/are non existent
  - e.g. if a user tries to manually navigate to the edit page for a different user's recipe, the resulting page is not a properly rendered page, and it doesn't redirect anywhere
  - 
  
- performance testing

- refactor
  - split models into different files, recipe, tags, comment, etc.
  - openAPI spec + syncing
  - the API endpoints should be restructured into
    - pages `/recipe/`
    - htmx, which will probably be `/htmx/`, gotta think about whether I can make them match the pages or whether they should rather be grouped with models or something
    - API `/api/`, i.e. those endpoints that are actually used either from interactive components or from API clients
- deploy
  - nginx https stuff
  - buy domain
  - register domain with Maileroo
- thhink about transaction handling for various operations, e.g. recipes and tags are separate tables
- SQL prepared statements? is this code vulnerable to SQL injection right now?
- there are a lot of inline structs, also a lot of duplicated ones
  - e.g. `CommentsWithUsername`
- list recipes that share ingredients with the current recipe
  - this would reuqire first implement some form of structured ingredient logging, i.e. instead of just using text, it would require structured inputs that is backed by some database
  - and/or this can be parsed from the freetext input
  - this might be an llm use case
  - probably rather do this at the end
- extract recipes from youtube videos or websites using llms
- add api endpoints for creating recipes
- calculate calorie information from ingredients
  - though this would need to be adjusted for ingredentis that are missing precise quantities
- add a collection feature
  - recipes can be added to collections
  - collections are owned by a user
  - a user can give other users access to said collection
- GDRP
  - export user info
  - delete account