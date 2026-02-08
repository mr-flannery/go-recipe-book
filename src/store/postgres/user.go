package postgres

import (
	"database/sql"
)

type UserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

func (s *UserStore) GetUsernameByID(userID int) (string, error) {
	var username string

	err := s.db.QueryRow("SELECT username FROM users WHERE id = $1", userID).Scan(&username)
	if err != nil {
		return "", err
	}

	return username, nil
}
