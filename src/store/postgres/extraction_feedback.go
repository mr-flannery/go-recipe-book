package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/mr-flannery/go-recipe-book/src/store"
)

type ExtractionFeedbackStore struct {
	db *sql.DB
}

func NewExtractionFeedbackStore(db *sql.DB) *ExtractionFeedbackStore {
	return &ExtractionFeedbackStore{db: db}
}

func (s *ExtractionFeedbackStore) Create(ctx context.Context, jobID, userID int, rating int, feedbackType string, comment *string) error {
	query := `
		INSERT INTO extraction_feedback (job_id, user_id, rating, feedback_type, comment)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (job_id, user_id) DO UPDATE SET
			rating = EXCLUDED.rating,
			feedback_type = EXCLUDED.feedback_type,
			comment = EXCLUDED.comment`

	_, err := s.db.ExecContext(ctx, query, jobID, userID, rating, feedbackType, comment)
	if err != nil {
		return fmt.Errorf("failed to create extraction feedback: %w", err)
	}

	return nil
}

func (s *ExtractionFeedbackStore) GetByJobID(ctx context.Context, jobID int) (*store.ExtractionFeedback, error) {
	query := `
		SELECT ef.id, ef.job_id, ef.user_id, u.username, ef.rating, ef.feedback_type, ef.comment, ef.created_at
		FROM extraction_feedback ef
		JOIN users u ON ef.user_id = u.id
		WHERE ef.job_id = $1`

	var feedback store.ExtractionFeedback
	err := s.db.QueryRowContext(ctx, query, jobID).Scan(
		&feedback.ID, &feedback.JobID, &feedback.UserID, &feedback.Username,
		&feedback.Rating, &feedback.FeedbackType, &feedback.Comment, &feedback.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get extraction feedback: %w", err)
	}

	return &feedback, nil
}

func (s *ExtractionFeedbackStore) GetAll(ctx context.Context, limit, offset int) ([]store.ExtractionFeedback, error) {
	query := `
		SELECT ef.id, ef.job_id, ef.user_id, u.username, ef.rating, ef.feedback_type, ef.comment, ef.created_at
		FROM extraction_feedback ef
		JOIN users u ON ef.user_id = u.id
		ORDER BY ef.created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query extraction feedback: %w", err)
	}
	defer rows.Close()

	var feedbacks []store.ExtractionFeedback
	for rows.Next() {
		var feedback store.ExtractionFeedback
		err := rows.Scan(
			&feedback.ID, &feedback.JobID, &feedback.UserID, &feedback.Username,
			&feedback.Rating, &feedback.FeedbackType, &feedback.Comment, &feedback.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan extraction feedback: %w", err)
		}
		feedbacks = append(feedbacks, feedback)
	}

	return feedbacks, nil
}

func (s *ExtractionFeedbackStore) CountAll(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM extraction_feedback`
	var count int
	err := s.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count extraction feedback: %w", err)
	}
	return count, nil
}
