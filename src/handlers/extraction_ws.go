package handlers

import (
	"bytes"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/mr-flannery/go-recipe-book/src/auth"
	"github.com/mr-flannery/go-recipe-book/src/logging"
	"github.com/mr-flannery/go-recipe-book/src/store"
)

type jobStatusFragment struct {
	Job      *store.ExtractionJob
	Feedback *store.ExtractionFeedback
}

// GetJobStatusSSEHandler streams job status updates to the client via
// Server-Sent Events. It sends an event whenever the status changes and
// closes the stream once the job reaches a terminal state.
func (h *Handler) GetJobStatusSSEHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userInfo := auth.GetUserInfoFromContext(ctx)

	jobIDStr := r.PathValue("id")
	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil {
		http.Error(w, "Invalid job ID", http.StatusBadRequest)
		return
	}

	job, err := h.ExtractionJobStore.GetByID(ctx, jobID)
	if err != nil || job == nil {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}
	if job.UserID != userInfo.UserID && !userInfo.IsAdmin {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no") // disable nginx buffering
	w.WriteHeader(http.StatusOK)

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	logging.AddMany(ctx, map[string]any{
		"action": "job.sse.connect",
		"job_id": jobID,
	})

	isTerminal := func(s string) bool {
		return s == "completed" || s == "failed"
	}

	sendFragment := func(current *jobStatusFragment) error {
		var buf bytes.Buffer
		if err := h.Renderer.Render(&buf, "job-status-fragment", current); err != nil {
			return err
		}
		// SSE format: each line of data prefixed with "data: ", terminated by two newlines.
		for _, line := range bytes.Split(buf.Bytes(), []byte("\n")) {
			if _, err := fmt.Fprintf(w, "data: %s\n", line); err != nil {
				return err
			}
		}
		_, err := fmt.Fprint(w, "\n")
		flusher.Flush()
		return err
	}

	lastStatus := job.Status
	feedback, _ := h.ExtractionFeedbackStore.GetByJobID(ctx, jobID)
	if err := sendFragment(&jobStatusFragment{Job: job, Feedback: feedback}); err != nil {
		slog.Debug("SSE: initial write failed", "job_id", jobID, "error", err)
		return
	}

	if isTerminal(lastStatus) {
		return
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.Context().Done():
			logging.AddMany(ctx, map[string]any{
				"action": "job.sse.disconnect",
				"job_id": jobID,
			})
			return

		case <-ticker.C:
			latest, err := h.ExtractionJobStore.GetByID(ctx, jobID)
			if err != nil || latest == nil {
				slog.Warn("SSE: failed to fetch job", "job_id", jobID, "error", err)
				continue
			}

			if latest.Status != lastStatus {
				lastStatus = latest.Status
				feedback, _ := h.ExtractionFeedbackStore.GetByJobID(ctx, jobID)
				if err := sendFragment(&jobStatusFragment{Job: latest, Feedback: feedback}); err != nil {
					slog.Debug("SSE: write failed", "job_id", jobID, "error", err)
					return
				}
				logging.AddMany(ctx, map[string]any{
					"action":     "job.sse.update",
					"job_id":     jobID,
					"new_status": latest.Status,
				})
			}

			if isTerminal(lastStatus) {
				return
			}
		}
	}
}
