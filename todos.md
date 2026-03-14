- UX improvements!!
  - when adding a tag, the focus isn't on the input component any more, but it should
  - also when in the tag filter component, when pressing backspace, it should remove the latest tag

- the current syntax for recipes and ingredients is not mobile friendly, probably
  - also on desktop some of the default shortcuts of the editor are annoying
  - if an ingredient is mentioned in the instructions, i want there to be syntax to automatically fill in that ingredient into the text. ideally also with the option to put in a fraction of the overall quantity, e.g. if the usage of some ingredient is split over several steps.
  - if the instructions mention some ingredient twice, there should be some section where everythings summed up, e.g. for the purpose of creating a shopping list. maybe a shopping list button would be the actual thing to do here.
  - also the whole thing really only works well with some integration with an ingredient API that also contains stuff like calorie count etc.

- recipe extraction for videos only works if there are subtitles, since this is where the transcript comes from. meaning that if I manage to find videos that fail due to this error, I need to figure out if I can feed the actual video to an LLM, though this be way more expensive.

- redo readme, it's completely outdated.
  - also make srue to include local deps (yt-dlp, ffmpeg) that are not covered by go deps management

- add websocket connections for notifications and updates to extraction job status and stuff like that

- ok so regarding ingredient databases:
  - there's the german BLS one, that seems to have the best data, but is German only
  - I can probably just supplement it with English AI translations for now
  - the questions is what's the UX for manually adding ingredients in recipes, and how should it work when extracting recipes
  - when manually adding, we need to show some kind of search bar, that let's users browse what matches their input
  - then they can select that ingredient
  - if search sucks, they can add an alias to something
    - e.g. we can alias "Reis poliert, roh" to "Reis", "Basmati Reis", etc.
  - when searching, we can prioritize aliases
  - otherwise, search results should probably be ranked by how much of the search team makes up the match results
    - e.g. when searching for "Reis", "Reis poliert, roh" should be ranked higher than "Schwein Vordereisbein/Vorderhaxe, gebraten ohne Fett (Ofen)"
  - search should be fuzzy, i.e. "Reis roh" should match "Reis poliert, roh"
  - assuming I have all of that, I can make the extracted recipes use the same search and pick the best match
    - if this turns out to suck, I need to manually tune it, so that searching for things returns the correct results
    - maybe I should store machine matches somewhere, so that I can review them, and then also backcorrect them somehow
  - also it might be useful, if a machine match sucks, to store what has initially been input as ingredient, so that I have the OG data that's required to correct it

- OWASP top10 checking would probably also be a good idea

- the vscode launch/debug setting is kinda annoying right now

- openrouter might be a good starting point to compare models for my use cases

- email notification on account approval is either missing or not working
  - it's because Maileroo is using a sandbox domain currently, to configure it correctly I need the actual domain, which I first have to buy
  - for which I need to checkout whether Railway supports external domains, and/or whether I just ditch the .de and use whatever is reasonable and cheap from their list

- I should take a close look at how many DB queries I currently need per read, given that the recipe read is rather inefficient
  - use concurrency as much as possible
  - try to unify queries of the result of one query depends on the result of a different one

- low prio: consider decoupling email sending via some async job/transactional outbox thing

- performance testing

- browser tests in CI
  - fix mobile tests

- login should work with both username and email address
  - which requires username to be unique

- cmd enter doesnt' seem to work any more since reworking the action buttons
  - the fuck. it has created the recipe 3 times.
  - either the redirecting doesn't work, or it's a problem with not blocking the button once the request has been sent already
  - probably I should do both: lock the button, show a loading indicator

- performance testing

- localization
- automated depenendcy updates

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
  - though this would need to be adjusted for ingredentis that are missing precise quantitieg
- add a collection feature
  - recipes can be added to collections
  - collections are owned by a user
  - a user can give other users access to said collection
- GDRP
  - export user info
  - delete account
