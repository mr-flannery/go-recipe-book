package postgres

import (
	"context"
	"database/sql"
)

type UserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

func (s *UserStore) GetUsernameByID(ctx context.Context, userID int) (string, error) {
	var username string

	err := s.db.QueryRowContext(ctx, "SELECT username FROM users WHERE id = $1", userID).Scan(&username)
	if err != nil {
		return "", err
	}

	return username, nil
}
