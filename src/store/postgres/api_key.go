package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/mr-flannery/go-recipe-book/src/store"
)

type APIKeyStore struct {
	db *sql.DB
}

func NewAPIKeyStore(db *sql.DB) *APIKeyStore {
	return &APIKeyStore{db: db}
}

func (s *APIKeyStore) Create(ctx context.Context, userID int, name string, keyHash string, keyPrefix string, encryptedKey string) (int, error) {
	query := `
		INSERT INTO api_keys (user_id, name, key_hash, key_prefix, encrypted_key)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	var id int
	err := s.db.QueryRowContext(ctx, query, userID, name, keyHash, keyPrefix, encryptedKey).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create API key: %w", err)
	}

	return id, nil
}

func (s *APIKeyStore) GetByKeyHash(ctx context.Context, keyHash string) (*store.APIKey, error) {
	query := `
		SELECT ak.id, ak.user_id, ak.name, ak.key_prefix, ak.created_at, ak.last_used_at
		FROM api_keys ak
		JOIN users u ON ak.user_id = u.id
		WHERE ak.key_hash = $1 AND u.is_active = true`

	var key store.APIKey
	err := s.db.QueryRowContext(ctx, query, keyHash).Scan(
		&key.ID, &key.UserID, &key.Name, &key.KeyPrefix, &key.CreatedAt, &key.LastUsedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	return &key, nil
}

func (s *APIKeyStore) GetByUserID(ctx context.Context, userID int) ([]store.APIKey, error) {
	query := `
		SELECT id, user_id, name, key_prefix, COALESCE(encrypted_key, ''), created_at, last_used_at
		FROM api_keys
		WHERE user_id = $1
		ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query API keys: %w", err)
	}
	defer rows.Close()

	var keys []store.APIKey
	for rows.Next() {
		var key store.APIKey
		err := rows.Scan(&key.ID, &key.UserID, &key.Name, &key.KeyPrefix, &key.EncryptedKey, &key.CreatedAt, &key.LastUsedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan API key: %w", err)
		}
		keys = append(keys, key)
	}

	return keys, nil
}

func (s *APIKeyStore) Delete(ctx context.Context, userID int, keyID int) error {
	query := `DELETE FROM api_keys WHERE id = $1 AND user_id = $2`
	result, err := s.db.ExecContext(ctx, query, keyID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete API key: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("API key not found")
	}

	return nil
}

func (s *APIKeyStore) UpdateLastUsed(ctx context.Context, keyID int) error {
	query := `UPDATE api_keys SET last_used_at = NOW() WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, keyID)
	if err != nil {
		return fmt.Errorf("failed to update last used: %w", err)
	}
	return nil
}
