package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mr-flannery/go-recipe-book/src/auth"
	"github.com/mr-flannery/go-recipe-book/src/store"
	tmocks "github.com/mr-flannery/go-recipe-book/src/templates/mocks"
)

type mockExtractionJobStore struct {
	getByIDFunc    func(ctx context.Context, id int) (*store.ExtractionJob, error)
	resetForRetry  func(ctx context.Context, id int) error
	resetCallCount int
}

func (m *mockExtractionJobStore) GetByID(ctx context.Context, id int) (*store.ExtractionJob, error) {
	return m.getByIDFunc(ctx, id)
}
func (m *mockExtractionJobStore) ResetForRetry(ctx context.Context, id int) error {
	m.resetCallCount++
	if m.resetForRetry != nil {
		return m.resetForRetry(ctx, id)
	}
	return nil
}
func (m *mockExtractionJobStore) Create(ctx context.Context, userID int, jobType string, inputURL *string, inputData []byte) (int, error) {
	return 0, nil
}
func (m *mockExtractionJobStore) GetByUserID(ctx context.Context, userID int, limit, offset int) ([]store.ExtractionJob, error) {
	return nil, nil
}
func (m *mockExtractionJobStore) CountByUserID(ctx context.Context, userID int) (int, error) {
	return 0, nil
}
func (m *mockExtractionJobStore) GetAll(ctx context.Context, limit, offset int) ([]store.ExtractionJob, error) {
	return nil, nil
}
func (m *mockExtractionJobStore) CountAll(ctx context.Context) (int, error) { return 0, nil }
func (m *mockExtractionJobStore) ClaimPendingJob(ctx context.Context) (*store.ExtractionJob, error) {
	return nil, nil
}
func (m *mockExtractionJobStore) UpdateStatus(ctx context.Context, id int, status string, errorMessage *string) error {
	return nil
}
func (m *mockExtractionJobStore) UpdateLLMData(ctx context.Context, id int, llmInput, llmOutput string) error {
	return nil
}
func (m *mockExtractionJobStore) SetRecipeID(ctx context.Context, id int, recipeID int) error {
	return nil
}
func (m *mockExtractionJobStore) MarkCompleted(ctx context.Context, id int) error { return nil }
func (m *mockExtractionJobStore) IncrementAttemptCount(ctx context.Context, id int) error {
	return nil
}
func (m *mockExtractionJobStore) ScheduleRetry(ctx context.Context, id int, retryAfter time.Time) error {
	return nil
}

func TestPostJobRetryHandler_StatusGuard(t *testing.T) {
	userInfo := &auth.UserInfo{IsLoggedIn: true, UserID: 1}

	tests := []struct {
		name             string
		job              *store.ExtractionJob
		wantRetry        bool
		wantRedirectPath string
	}{
		{
			name: "failed job is retried",
			job: &store.ExtractionJob{
				ID: 1, UserID: 1, Status: "failed",
				UpdatedAt: time.Now().Add(-10 * time.Minute),
			},
			wantRetry:        true,
			wantRedirectPath: "/account/jobs/1?success=Job queued for retry",
		},
		{
			name: "processing job stuck > 5min is retried",
			job: &store.ExtractionJob{
				ID: 1, UserID: 1, Status: "processing",
				UpdatedAt: time.Now().Add(-6 * time.Minute),
			},
			wantRetry:        true,
			wantRedirectPath: "/account/jobs/1?success=Job queued for retry",
		},
		{
			name: "processing job under 5min is rejected",
			job: &store.ExtractionJob{
				ID: 1, UserID: 1, Status: "processing",
				UpdatedAt: time.Now().Add(-4 * time.Minute),
			},
			wantRetry:        false,
			wantRedirectPath: "/account/jobs/1?error=Only failed or stuck processing jobs can be retried",
		},
		{
			name: "pending job is rejected",
			job: &store.ExtractionJob{
				ID: 1, UserID: 1, Status: "pending",
				UpdatedAt: time.Now().Add(-10 * time.Minute),
			},
			wantRetry:        false,
			wantRedirectPath: "/account/jobs/1?error=Only failed or stuck processing jobs can be retried",
		},
		{
			name: "completed job is rejected",
			job: &store.ExtractionJob{
				ID: 1, UserID: 1, Status: "completed",
				UpdatedAt: time.Now().Add(-10 * time.Minute),
			},
			wantRetry:        false,
			wantRedirectPath: "/account/jobs/1?error=Only failed or stuck processing jobs can be retried",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jobStore := &mockExtractionJobStore{
				getByIDFunc: func(_ context.Context, _ int) (*store.ExtractionJob, error) {
					return tt.job, nil
				},
			}

			h := &Handler{
				ExtractionJobStore: jobStore,
				Renderer:           &tmocks.MockRenderer{},
			}

			req := httptest.NewRequest(http.MethodPost, "/account/jobs/1/retry", nil)
			req.SetPathValue("id", "1")
			req = req.WithContext(auth.ContextWithUserInfo(req.Context(), userInfo))
			rec := httptest.NewRecorder()

			h.PostJobRetryHandler(rec, req)

			if tt.wantRetry && jobStore.resetCallCount != 1 {
				t.Errorf("expected ResetForRetry to be called once, got %d", jobStore.resetCallCount)
			}
			if !tt.wantRetry && jobStore.resetCallCount != 0 {
				t.Errorf("expected ResetForRetry not to be called, got %d calls", jobStore.resetCallCount)
			}

			location := rec.Header().Get("Location")
			if location != tt.wantRedirectPath {
				t.Errorf("redirect location = %q, want %q", location, tt.wantRedirectPath)
			}
		})
	}
}
