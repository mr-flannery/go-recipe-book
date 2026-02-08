package handlers

import (
	"database/sql"

	"github.com/mr-flannery/go-recipe-book/src/store"
)

type Handler struct {
	DB           *sql.DB
	RecipeStore  store.RecipeStore
	TagStore     store.TagStore
	UserTagStore store.UserTagStore
	CommentStore store.CommentStore
	UserStore    store.UserStore
	AuthStore    store.AuthStore
}

func NewHandler(db *sql.DB, recipeStore store.RecipeStore, tagStore store.TagStore, userTagStore store.UserTagStore, commentStore store.CommentStore, userStore store.UserStore, authStore store.AuthStore) *Handler {
	return &Handler{
		DB:           db,
		RecipeStore:  recipeStore,
		TagStore:     tagStore,
		UserTagStore: userTagStore,
		CommentStore: commentStore,
		UserStore:    userStore,
		AuthStore:    authStore,
	}
}
