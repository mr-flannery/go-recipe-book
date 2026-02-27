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
		"SELECT user_id, page_size, COALESCE(view_mode, $2) FROM user_preferences WHERE user_id = $1",
		userID, models.DefaultViewMode,
	).Scan(&prefs.UserID, &prefs.PageSize, &prefs.ViewMode)

	if err == sql.ErrNoRows {
		return &models.UserPreferences{
			UserID:   userID,
			PageSize: models.DefaultPageSize,
			ViewMode: models.DefaultViewMode,
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

func (s *UserPreferencesStore) SetViewMode(userID int, viewMode string) error {
	_, err := s.db.Exec(
		`INSERT INTO user_preferences (user_id, page_size, view_mode, updated_at)
		VALUES ($1, $2, $3, NOW())
		ON CONFLICT (user_id) DO UPDATE SET view_mode = $3, updated_at = NOW()`,
		userID, models.DefaultPageSize, viewMode,
	)
	return err
}
