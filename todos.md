- SORT TODOs
  - what do I really need to do before deploying
  - what is nice to have but can be done after going live with the intial version

- UX improvements!!
  - also I need to test the mobile layout!
  - the admin registration UX could be improved, by listing denied and approved registrations, and also by make the experience more dynamic, showing success messages or something like this.
  - when adding a tag, the focus isn't on the input component any more, but it should
  - also when in the tag filter component, when pressing backspace, it should remove the latest tag

- the current syntax for recipes and ingredients is not mobile friendly, probably
  - also on desktop some of the default shortcuts of the editor are annoying
  - if an ingredient is mentioned in the instructions, i want there to be syntax to automatically fill in that ingredient into the text. ideally also with the option to put in a fraction of the overall quantity, e.g. if the usage of some ingredient is split over several steps.
  - if the instructions mention some ingredient twice, there should be some section where everythings summed up, e.g. for the purpose of creating a shopping list. maybe a shopping list button would be the actual thing to do here.
  - also the whole thing really only works well with some integration with an ingredient API that also contains stuff like calorie count etc.

- sql injection???

- the vscode launch/debug setting is kinda annoying right now

- cookie banner
- is the imprint up to date?
  - also how do I handle the imprint containing my address without pushing my address to github?
  - do it as a github secret?

- SEO/performance testing
  - can I use gzip and shit like this?

- browser tests in CI
  - currently only test for chromium?
  - check if parallelism can be increased

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