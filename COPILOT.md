# CoPilot

This is the repository for a recipe management application written in Go, using PostgreSQL and htmx.

## Features

- a recipe must have a title, a list of ingredients, and instructions, written in freetext
- users should be able to use markdown syntax to write the ingredients and instructions
- a recipe can have labels
- a label is a user created entity that contains arbitrary text
- labels and recipes have a n:m relationship
- a recipe can also have preparation time, cook time, and estimated calorie count
- user should be able to submit new recipes, update existing ones, and delete them again
- users should be able to leave comments on recipes
- there should be an overview page that lists all recipes, on which users should be able to search for recipes and filter them by label

## Architecture

- the backend of the application is written in Go
- for the frontend, it uses server-side rendered html pages using go's built-in templating plus htmx for dynamic content, like the list recipes page
- the application should be runnable via a docker container