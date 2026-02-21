package handlers

import (
	"database/sql"

	"github.com/mr-flannery/go-recipe-book/src/mail"
	"github.com/mr-flannery/go-recipe-book/src/store"
	"github.com/mr-flannery/go-recipe-book/src/templates"
)

type Handler struct {
	DB              *sql.DB
	RecipeStore     store.RecipeStore
	TagStore        store.TagStore
	UserTagStore    store.UserTagStore
	CommentStore    store.CommentStore
	UserStore       store.UserStore
	AuthStore       store.AuthStore
	IngredientStore store.IngredientStore
	Renderer        templates.Renderer
	MailClient      mail.MailClient
}

func NewHandler(db *sql.DB, recipeStore store.RecipeStore, tagStore store.TagStore, userTagStore store.UserTagStore, commentStore store.CommentStore, userStore store.UserStore, authStore store.AuthStore, ingredientStore store.IngredientStore, renderer templates.Renderer, mailClient mail.MailClient) *Handler {
	return &Handler{
		DB:              db,
		RecipeStore:     recipeStore,
		TagStore:        tagStore,
		UserTagStore:    userTagStore,
		CommentStore:    commentStore,
		UserStore:       userStore,
		AuthStore:       authStore,
		IngredientStore: ingredientStore,
		Renderer:        renderer,
		MailClient:      mailClient,
	}
}
