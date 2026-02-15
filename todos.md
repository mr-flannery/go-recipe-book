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
- UX improvements!!
  - the admin registration UX could be improved, by listing denied and approved registrations, and also by make the experience more dynamic, showing success messages or something like this.
  - when adding a tag, the focus isn't on the input component any more, but it should
  - also when in the tag filter component, when pressing backspace, it should remove the latest tag

- cmd-enter should trigger certain actions on certain pages

- the vscode launch/debug setting is kinda annoying right now
- mail tests?
- cookie banner

- performance testing

- localization

- automated depenendcy updates

- observability
  - logging
  - metrics
  - alerts
  - traces?

- integration tests with test container are currently serial
  - should I need more speed, I can investigate concurrency
  - this probabaly requires a pool of databases, or even separate containers

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