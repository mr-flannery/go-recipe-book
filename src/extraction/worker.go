package extraction

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/mr-flannery/go-recipe-book/src/logging"
	"github.com/mr-flannery/go-recipe-book/src/mail"
	"github.com/mr-flannery/go-recipe-book/src/models"
	"github.com/mr-flannery/go-recipe-book/src/store"
)

var tracer = otel.Tracer("extraction")

const (
	maxAutoRetries  = 1
	technicalRetry  = 1 * time.Hour
)

// TechnicalError wraps errors that are caused by transient infrastructure
// issues (e.g. model unavailable, API down). Jobs that fail with a
// TechnicalError are automatically rescheduled rather than permanently failed.
type TechnicalError struct {
	cause error
}

func (e *TechnicalError) Error() string { return e.cause.Error() }
func (e *TechnicalError) Unwrap() error { return e.cause }

func technicalErrorf(format string, args ...any) error {
	return &TechnicalError{cause: fmt.Errorf(format, args...)}
}

type WorkerConfig struct {
	Concurrency      int
	PollInterval     time.Duration
	OpenRouterAPIKey string
	BaseURL          string
}

type Worker struct {
	config      WorkerConfig
	jobStore    store.ExtractionJobStore
	recipeStore store.RecipeStore
	tagStore    store.TagStore
	authStore   store.AuthStore
	mailClient  mail.MailClient
	llmClient   *LLMClient
	stopCh      chan struct{}
	wg          sync.WaitGroup
}

func NewWorker(
	config WorkerConfig,
	jobStore store.ExtractionJobStore,
	recipeStore store.RecipeStore,
	tagStore store.TagStore,
	authStore store.AuthStore,
	mailClient mail.MailClient,
) *Worker {
	return &Worker{
		config:      config,
		jobStore:    jobStore,
		recipeStore: recipeStore,
		tagStore:    tagStore,
		authStore:   authStore,
		mailClient:  mailClient,
		llmClient:   NewLLMClient(config.OpenRouterAPIKey),
		stopCh:      make(chan struct{}),
	}
}

func (w *Worker) Start() {
	slog.Info("Starting extraction workers", "concurrency", w.config.Concurrency)

	for i := 0; i < w.config.Concurrency; i++ {
		w.wg.Add(1)
		go w.workerLoop(i)
	}
}

func (w *Worker) Stop() {
	slog.Info("Stopping extraction workers...")
	close(w.stopCh)
	w.wg.Wait()
	slog.Info("Extraction workers stopped")
}

func (w *Worker) workerLoop(workerID int) {
	defer w.wg.Done()

	ticker := time.NewTicker(w.config.PollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.stopCh:
			return
		case <-ticker.C:
			w.processNextJob(workerID)
		}
	}
}

func (w *Worker) processNextJob(workerID int) {
	ctx := context.Background()

	job, err := w.jobStore.ClaimPendingJob(ctx)
	if err != nil {
		slog.Error("Failed to claim job", "worker", workerID, "error", err)
		return
	}

	if job == nil {
		return
	}

	ctx, span := tracer.Start(ctx, "extraction.process_job")
	defer span.End()

	span.SetAttributes(
		attribute.Int("job.id", job.ID),
		attribute.String("job.type", job.JobType),
		attribute.Int("job.attempt", job.AttemptCount+1),
		attribute.Int("worker.id", workerID),
	)

	slog.Info("Processing extraction job",
		"worker", workerID,
		"job_id", job.ID,
		"job_type", job.JobType,
		"attempt", job.AttemptCount+1,
	)

	if err := w.jobStore.IncrementAttemptCount(ctx, job.ID); err != nil {
		slog.Error("Failed to increment attempt count", "job_id", job.ID, "error", err)
	}

	err = w.processJob(ctx, job)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		w.handleJobFailure(ctx, job, err)
		return
	}

	span.SetStatus(codes.Ok, "")
	slog.Info("Job completed successfully", "job_id", job.ID)
}

func (w *Worker) processJob(ctx context.Context, job *store.ExtractionJob) error {
	var content string
	var llmInput string
	var recipe *ExtractedRecipe
	var err error

	switch job.JobType {
	case "website":
		if job.InputURL == nil {
			return errors.New("website job missing input URL")
		}

		_, fetchSpan := tracer.Start(ctx, "extraction.fetch_website")
		var usedURL string
		content, usedURL, err = FetchWebsiteContent(*job.InputURL)
		fetchSpan.End()
		if err != nil {
			return fmt.Errorf("failed to fetch website: %w", err)
		}
		if usedURL != *job.InputURL {
			slog.Info("Used Wayback Machine archive for website extraction",
				"job_id", job.ID,
				"original_url", *job.InputURL,
				"archive_url", usedURL,
			)
			logging.AddMany(ctx, map[string]any{
				"extraction.wayback_fallback": true,
				"extraction.archive_url":      usedURL,
			})
		}

		llmCtx, llmSpan := tracer.Start(ctx, "extraction.llm_extract")
		llmSpan.SetAttributes(attribute.String("extraction.source", "website"))
		llmInput, recipe, err = w.llmClient.ExtractRecipeFromText(llmCtx, "website", content)
		llmSpan.End()

	case "video":
		if job.InputURL == nil {
			return errors.New("video job missing input URL")
		}

		_, metaSpan := tracer.Start(ctx, "extraction.fetch_video_metadata")
		metadata, metaErr := FetchVideoMetadata(*job.InputURL)
		metaSpan.End()
		if metaErr != nil && !errors.Is(metaErr, ErrVideoUnavailable) {
			slog.Warn("Failed to fetch video metadata", "job_id", job.ID, "error", metaErr)
		}

		var additionalContext string
		if metadata != nil {
			additionalContext = w.buildVideoContext(ctx, metadata)
		}

		_, transcriptSpan := tracer.Start(ctx, "extraction.fetch_transcript")
		content, err = FetchYouTubeTranscript(*job.InputURL, []string{"en", "de"})
		transcriptSpan.End()
		if err != nil {
			if errors.Is(err, ErrNoCaptions) {
				slog.Info("No captions available, falling back to audio extraction", "job_id", job.ID)

				audioCtx, audioSpan := tracer.Start(ctx, "extraction.audio_extract")
				llmInput, recipe, err = w.extractFromAudio(audioCtx, job, additionalContext)
				audioSpan.End()
				if err != nil {
					return fmt.Errorf("audio extraction failed: %w", err)
				}
			} else {
				return fmt.Errorf("failed to fetch transcript: %w", err)
			}
		} else {
			if additionalContext != "" {
				content = content + "\n\n---\n\nAdditional context from video description:\n" + additionalContext
			}
			llmCtx, llmSpan := tracer.Start(ctx, "extraction.llm_extract")
			llmSpan.SetAttributes(attribute.String("extraction.source", "transcript"))
			llmInput, recipe, err = w.llmClient.ExtractRecipeFromText(llmCtx, "transcript", content)
			llmSpan.End()
		}

	case "image":
		if len(job.InputData) == 0 {
			return errors.New("image job missing input data")
		}
		llmCtx, llmSpan := tracer.Start(ctx, "extraction.llm_extract")
		llmSpan.SetAttributes(attribute.String("extraction.source", "image"))
		llmInput, recipe, err = w.llmClient.ExtractRecipeFromImage(llmCtx, job.InputData, "")
		llmSpan.End()

	default:
		return fmt.Errorf("unknown job type: %s", job.JobType)
	}

	if err != nil {
		if llmInput != "" {
			_ = w.jobStore.UpdateLLMData(ctx, job.ID, llmInput, "")
		}
		return fmt.Errorf("LLM extraction failed: %w", err)
	}

	llmOutput := fmt.Sprintf(`{"title": %q, "description": %q, "confidence": %.2f}`,
		recipe.Title, recipe.Description, recipe.Confidence)
	if err := w.jobStore.UpdateLLMData(ctx, job.ID, llmInput, llmOutput); err != nil {
		slog.Error("Failed to update LLM data", "job_id", job.ID, "error", err)
	}

	recipeModel := models.Recipe{
		Title:          recipe.Title,
		Description:    recipe.Description,
		IngredientsMD:  recipe.IngredientsMD,
		InstructionsMD: recipe.InstructionsMD,
		AuthorID:       job.UserID,
	}

	if job.InputURL != nil {
		recipeModel.Source = *job.InputURL
	}

	if recipe.PrepTimeMinutes != nil {
		recipeModel.PrepTime = *recipe.PrepTimeMinutes
	}
	if recipe.CookTimeMinutes != nil {
		recipeModel.CookTime = *recipe.CookTimeMinutes
	}
	if recipe.CaloriesPerServing != nil {
		recipeModel.Calories = *recipe.CaloriesPerServing
	}

	saveCtx, saveSpan := tracer.Start(ctx, "extraction.save_recipe")
	recipeID, err := w.recipeStore.Save(saveCtx, recipeModel)
	saveSpan.End()
	if err != nil {
		return fmt.Errorf("failed to save recipe: %w", err)
	}

	logging.Add(ctx, "recipe.id", recipeID)
	logging.Add(ctx, "recipe.title", recipe.Title)
	logging.Add(ctx, "recipe.confidence", recipe.Confidence)

	if len(recipe.SuggestedTags) > 0 {
		if err := w.tagStore.SetRecipeTags(ctx, recipeID, recipe.SuggestedTags); err != nil {
			slog.Error("Failed to set recipe tags", "recipe_id", recipeID, "error", err)
		}
	}

	if err := w.jobStore.SetRecipeID(ctx, job.ID, recipeID); err != nil {
		slog.Error("Failed to set recipe ID on job", "job_id", job.ID, "error", err)
	}

	if err := w.jobStore.MarkCompleted(ctx, job.ID); err != nil {
		slog.Error("Failed to mark job completed", "job_id", job.ID, "error", err)
	}

	w.sendSuccessNotification(ctx, job, recipe.Title, recipeID)

	return nil
}

func (w *Worker) handleJobFailure(ctx context.Context, job *store.ExtractionJob, jobErr error) {
	slog.Error("Job failed",
		"job_id", job.ID,
		"attempt", job.AttemptCount+1,
		"error", jobErr,
	)

	var techErr *TechnicalError
	if errors.As(jobErr, &techErr) {
		retryAt := time.Now().Add(technicalRetry)
		slog.Info("Technical failure, scheduling hourly retry",
			"job_id", job.ID,
			"retry_at", retryAt,
		)
		if err := w.jobStore.ScheduleRetry(ctx, job.ID, retryAt); err != nil {
			slog.Error("Failed to schedule retry", "job_id", job.ID, "error", err)
		}
		return
	}

	currentAttempt := job.AttemptCount + 1

	if currentAttempt <= maxAutoRetries {
		slog.Info("Scheduling automatic retry", "job_id", job.ID, "attempt", currentAttempt)
		if err := w.jobStore.ResetForRetry(ctx, job.ID); err != nil {
			slog.Error("Failed to reset job for retry", "job_id", job.ID, "error", err)
		}
		return
	}

	errMsg := jobErr.Error()
	if err := w.jobStore.UpdateStatus(ctx, job.ID, "failed", &errMsg); err != nil {
		slog.Error("Failed to update job status", "job_id", job.ID, "error", err)
	}

	w.sendFailureNotification(ctx, job, errMsg)
}

func (w *Worker) sendSuccessNotification(ctx context.Context, job *store.ExtractionJob, recipeTitle string, recipeID int) {
	user, err := w.authStore.GetUserByID(ctx, job.UserID)
	if err != nil {
		slog.Error("Failed to get user for notification", "user_id", job.UserID, "error", err)
		return
	}

	recipeURL := fmt.Sprintf("%s/recipes/%d", w.config.BaseURL, recipeID)
	if err := mail.SendExtractionSuccessNotification(ctx, w.mailClient, user.Email, user.Username, recipeTitle, recipeURL); err != nil {
		slog.Error("Failed to send success notification", "job_id", job.ID, "error", err)
	}
}

func (w *Worker) sendFailureNotification(ctx context.Context, job *store.ExtractionJob, errorMessage string) {
	user, err := w.authStore.GetUserByID(ctx, job.UserID)
	if err != nil {
		slog.Error("Failed to get user for notification", "user_id", job.UserID, "error", err)
		return
	}

	jobURL := fmt.Sprintf("%s/account/jobs/%d", w.config.BaseURL, job.ID)
	if err := mail.SendExtractionFailureNotification(ctx, w.mailClient, user.Email, user.Username, errorMessage, jobURL); err != nil {
		slog.Error("Failed to send failure notification", "job_id", job.ID, "error", err)
	}
}

func (w *Worker) buildVideoContext(ctx context.Context, metadata *VideoMetadata) string {
	var parts []string

	if metadata.Description != "" {
		parts = append(parts, "Video description:\n"+metadata.Description)
	}

	if len(metadata.RecipeLinks) > 0 {
		for _, link := range metadata.RecipeLinks {
			html, _, err := FetchWebsiteContent(link)
			if err != nil {
				slog.Warn("Failed to fetch recipe link", "url", link, "error", err)
				continue
			}
			parts = append(parts, fmt.Sprintf("Recipe from linked page (%s):\n%s", link, html))
		}
	}

	return strings.Join(parts, "\n\n---\n\n")
}

func (w *Worker) extractFromAudio(ctx context.Context, job *store.ExtractionJob, additionalContext string) (string, *ExtractedRecipe, error) {
	audioResult, err := DownloadYouTubeAudio(*job.InputURL)
	if err != nil {
		return "", nil, fmt.Errorf("failed to download audio: %w", err)
	}
	defer func() {
		if cleanupErr := audioResult.Cleanup(); cleanupErr != nil {
			slog.Warn("Failed to cleanup audio file", "job_id", job.ID, "error", cleanupErr)
		}
	}()

	audioData, err := os.ReadFile(audioResult.FilePath)
	if err != nil {
		return "", nil, technicalErrorf("failed to read audio file: %w", err)
	}

	return w.llmClient.ExtractRecipeFromAudio(ctx, audioData, additionalContext)
}
