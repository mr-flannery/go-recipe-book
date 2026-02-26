package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/mr-flannery/go-recipe-book/src/models"
)

type CommentStore struct {
	db *sql.DB
}

func NewCommentStore(db *sql.DB) *CommentStore {
	return &CommentStore{db: db}
}

func (s *CommentStore) GetByRecipeID(recipeID string) ([]models.Comment, error) {
	rows, err := s.db.Query("SELECT id, recipe_id, author_id, content_md, created_at, updated_at FROM comments WHERE recipe_id = $1 ORDER BY created_at DESC", recipeID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch comments: %v", err)
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		if err := rows.Scan(&comment.ID, &comment.RecipeID, &comment.AuthorID, &comment.ContentMD, &comment.CreatedAt, &comment.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan comment: %v", err)
		}
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over comments: %v", err)
	}

	return comments, nil
}

func (s *CommentStore) Save(comment models.Comment) error {
	query := `INSERT INTO comments (recipe_id, author_id, content_md, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5)`

	_, err := s.db.Exec(query, comment.RecipeID, comment.AuthorID, comment.ContentMD, time.Now(), time.Now())
	return err
}

func (s *CommentStore) GetLatestByUserAndRecipe(userID int, recipeID int) (models.Comment, error) {
	var comment models.Comment

	err := s.db.QueryRow(
		"SELECT id, recipe_id, author_id, content_md, created_at, updated_at FROM comments WHERE author_id = $1 AND recipe_id = $2 ORDER BY created_at DESC LIMIT 1",
		userID, recipeID,
	).Scan(&comment.ID, &comment.RecipeID, &comment.AuthorID, &comment.ContentMD, &comment.CreatedAt, &comment.UpdatedAt)

	if err != nil {
		return models.Comment{}, err
	}

	return comment, nil
}

func (s *CommentStore) GetByID(commentID int) (models.Comment, error) {
	var comment models.Comment

	err := s.db.QueryRow(
		"SELECT id, recipe_id, author_id, content_md, created_at, updated_at FROM comments WHERE id = $1",
		commentID,
	).Scan(&comment.ID, &comment.RecipeID, &comment.AuthorID, &comment.ContentMD, &comment.CreatedAt, &comment.UpdatedAt)

	if err != nil {
		return models.Comment{}, err
	}

	return comment, nil
}

func (s *CommentStore) GetByUserID(userID int) ([]models.Comment, error) {
	rows, err := s.db.Query(
		"SELECT id, recipe_id, author_id, content_md, created_at, updated_at FROM comments WHERE author_id = $1 ORDER BY created_at DESC",
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch comments: %v", err)
	}
	defer rows.Close()

	var comments []models.Comment
	for rows.Next() {
		var comment models.Comment
		if err := rows.Scan(&comment.ID, &comment.RecipeID, &comment.AuthorID, &comment.ContentMD, &comment.CreatedAt, &comment.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan comment: %v", err)
		}
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over comments: %v", err)
	}

	return comments, nil
}

func (s *CommentStore) Update(commentID int, content string) error {
	_, err := s.db.Exec(
		"UPDATE comments SET content_md = $1, updated_at = $2 WHERE id = $3",
		content, time.Now(), commentID,
	)
	return err
}

func (s *CommentStore) Delete(commentID int) error {
	_, err := s.db.Exec("DELETE FROM comments WHERE id = $1", commentID)
	return err
}
