package postgres

import (
	"database/sql"

	"github.com/mr-flannery/go-recipe-book/src/models"
)

type UserPreferencesStore struct {
	db *sql.DB
}

func NewUserPreferencesStore(db *sql.DB) *UserPreferencesStore {
	return &UserPreferencesStore{db: db}
}

func (s *UserPreferencesStore) Get(userID int) (*models.UserPreferences, error) {
	var prefs models.UserPreferences
	err := s.db.QueryRow(
		"SELECT user_id, page_size FROM user_preferences WHERE user_id = $1",
		userID,
	).Scan(&prefs.UserID, &prefs.PageSize)

	if err == sql.ErrNoRows {
		return &models.UserPreferences{
			UserID:   userID,
			PageSize: models.DefaultPageSize,
		}, nil
	}
	if err != nil {
		return nil, err
	}
	return &prefs, nil
}

func (s *UserPreferencesStore) SetPageSize(userID, pageSize int) error {
	_, err := s.db.Exec(
		`INSERT INTO user_preferences (user_id, page_size, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (user_id) DO UPDATE SET page_size = $2, updated_at = NOW()`,
		userID, pageSize,
	)
	return err
}
