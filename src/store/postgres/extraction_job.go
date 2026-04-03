package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/mr-flannery/go-recipe-book/src/store"
)

type ExtractionJobStore struct {
	db *sql.DB
}

func NewExtractionJobStore(db *sql.DB) *ExtractionJobStore {
	return &ExtractionJobStore{db: db}
}

func (s *ExtractionJobStore) Create(ctx context.Context, userID int, jobType string, inputURL *string, inputData []byte) (int, error) {
	query := `
		INSERT INTO extraction_jobs (user_id, job_type, input_url, input_data)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	var id int
	err := s.db.QueryRowContext(ctx, query, userID, jobType, inputURL, inputData).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to create extraction job: %w", err)
	}

	return id, nil
}

func (s *ExtractionJobStore) GetByID(ctx context.Context, id int) (*store.ExtractionJob, error) {
	query := `
		SELECT 
			ej.id, ej.user_id, u.username, ej.job_type, ej.input_url, ej.input_data,
			ej.status, ej.error_message, ej.llm_input, ej.llm_output,
			ej.recipe_id, r.title, ej.attempt_count, ej.created_at, ej.updated_at, ej.completed_at
		FROM extraction_jobs ej
		JOIN users u ON ej.user_id = u.id
		LEFT JOIN recipes r ON ej.recipe_id = r.id
		WHERE ej.id = $1`

	var job store.ExtractionJob
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&job.ID, &job.UserID, &job.Username, &job.JobType, &job.InputURL, &job.InputData,
		&job.Status, &job.ErrorMessage, &job.LLMInput, &job.LLMOutput,
		&job.RecipeID, &job.RecipeTitle, &job.AttemptCount, &job.CreatedAt, &job.UpdatedAt, &job.CompletedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get extraction job: %w", err)
	}

	return &job, nil
}

func (s *ExtractionJobStore) GetByUserID(ctx context.Context, userID int, limit, offset int) ([]store.ExtractionJob, error) {
	query := `
		SELECT 
			ej.id, ej.user_id, u.username, ej.job_type, ej.input_url, NULL,
			ej.status, ej.error_message, NULL, NULL,
			ej.recipe_id, r.title, ej.attempt_count, ej.created_at, ej.updated_at, ej.completed_at
		FROM extraction_jobs ej
		JOIN users u ON ej.user_id = u.id
		LEFT JOIN recipes r ON ej.recipe_id = r.id
		WHERE ej.user_id = $1
		ORDER BY ej.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := s.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query extraction jobs: %w", err)
	}
	defer rows.Close()

	return scanJobs(rows)
}

func (s *ExtractionJobStore) CountByUserID(ctx context.Context, userID int) (int, error) {
	query := `SELECT COUNT(*) FROM extraction_jobs WHERE user_id = $1`
	var count int
	err := s.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count extraction jobs: %w", err)
	}
	return count, nil
}

func (s *ExtractionJobStore) GetAll(ctx context.Context, limit, offset int) ([]store.ExtractionJob, error) {
	query := `
		SELECT 
			ej.id, ej.user_id, u.username, ej.job_type, ej.input_url, NULL,
			ej.status, ej.error_message, NULL, NULL,
			ej.recipe_id, r.title, ej.attempt_count, ej.created_at, ej.updated_at, ej.completed_at
		FROM extraction_jobs ej
		JOIN users u ON ej.user_id = u.id
		LEFT JOIN recipes r ON ej.recipe_id = r.id
		ORDER BY ej.created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query extraction jobs: %w", err)
	}
	defer rows.Close()

	return scanJobs(rows)
}

func (s *ExtractionJobStore) CountAll(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM extraction_jobs`
	var count int
	err := s.db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count extraction jobs: %w", err)
	}
	return count, nil
}

func (s *ExtractionJobStore) ClaimPendingJob(ctx context.Context) (*store.ExtractionJob, error) {
	query := `
		UPDATE extraction_jobs
		SET status = 'processing', updated_at = NOW()
		WHERE id = (
			SELECT id FROM extraction_jobs
			WHERE status = 'pending'
			  AND (retry_after IS NULL OR retry_after <= NOW())
			ORDER BY created_at ASC
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id, user_id, job_type, input_url, input_data, status, error_message,
		          llm_input, llm_output, recipe_id, attempt_count, created_at, updated_at, completed_at, retry_after`

	var job store.ExtractionJob
	err := s.db.QueryRowContext(ctx, query).Scan(
		&job.ID, &job.UserID, &job.JobType, &job.InputURL, &job.InputData,
		&job.Status, &job.ErrorMessage, &job.LLMInput, &job.LLMOutput,
		&job.RecipeID, &job.AttemptCount, &job.CreatedAt, &job.UpdatedAt, &job.CompletedAt, &job.RetryAfter,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to claim pending job: %w", err)
	}

	return &job, nil
}

func (s *ExtractionJobStore) UpdateStatus(ctx context.Context, id int, status string, errorMessage *string) error {
	query := `UPDATE extraction_jobs SET status = $2, error_message = $3, updated_at = NOW() WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, id, status, errorMessage)
	if err != nil {
		return fmt.Errorf("failed to update job status: %w", err)
	}
	return nil
}

func (s *ExtractionJobStore) UpdateLLMData(ctx context.Context, id int, llmInput, llmOutput string) error {
	query := `UPDATE extraction_jobs SET llm_input = $2, llm_output = $3, updated_at = NOW() WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, id, llmInput, llmOutput)
	if err != nil {
		return fmt.Errorf("failed to update LLM data: %w", err)
	}
	return nil
}

func (s *ExtractionJobStore) SetRecipeID(ctx context.Context, id int, recipeID int) error {
	query := `UPDATE extraction_jobs SET recipe_id = $2, updated_at = NOW() WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, id, recipeID)
	if err != nil {
		return fmt.Errorf("failed to set recipe ID: %w", err)
	}
	return nil
}

func (s *ExtractionJobStore) MarkCompleted(ctx context.Context, id int) error {
	query := `UPDATE extraction_jobs SET status = 'completed', completed_at = NOW(), updated_at = NOW() WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to mark job completed: %w", err)
	}
	return nil
}

func (s *ExtractionJobStore) IncrementAttemptCount(ctx context.Context, id int) error {
	query := `UPDATE extraction_jobs SET attempt_count = attempt_count + 1, updated_at = NOW() WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to increment attempt count: %w", err)
	}
	return nil
}

func (s *ExtractionJobStore) ResetForRetry(ctx context.Context, id int) error {
	query := `UPDATE extraction_jobs SET status = 'pending', error_message = NULL, retry_after = NULL, updated_at = NOW() WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to reset job for retry: %w", err)
	}
	return nil
}

func (s *ExtractionJobStore) ScheduleRetry(ctx context.Context, id int, retryAfter time.Time) error {
	query := `UPDATE extraction_jobs SET status = 'pending', error_message = NULL, retry_after = $2, updated_at = NOW() WHERE id = $1`
	_, err := s.db.ExecContext(ctx, query, id, retryAfter)
	if err != nil {
		return fmt.Errorf("failed to schedule job retry: %w", err)
	}
	return nil
}

func scanJobs(rows *sql.Rows) ([]store.ExtractionJob, error) {
	var jobs []store.ExtractionJob
	for rows.Next() {
		var job store.ExtractionJob
		err := rows.Scan(
			&job.ID, &job.UserID, &job.Username, &job.JobType, &job.InputURL, &job.InputData,
			&job.Status, &job.ErrorMessage, &job.LLMInput, &job.LLMOutput,
			&job.RecipeID, &job.RecipeTitle, &job.AttemptCount, &job.CreatedAt, &job.UpdatedAt, &job.CompletedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan extraction job: %w", err)
		}
		jobs = append(jobs, job)
	}
	return jobs, nil
}
