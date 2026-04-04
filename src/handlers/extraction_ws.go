package handlers

import (
	"bytes"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mr-flannery/go-recipe-book/src/auth"
	"github.com/mr-flannery/go-recipe-book/src/logging"
	"github.com/mr-flannery/go-recipe-book/src/store"
)

type jobStatusFragment struct {
	Job      *store.ExtractionJob
	Feedback *store.ExtractionFeedback
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Origin is already validated by the session cookie auth that precedes this handler.
		return true
	},
}

// GetJobStatusWSHandler upgrades the connection to a WebSocket and streams
// HTMX HTML fragments whenever the extraction job status changes.
// It closes the connection once the job reaches a terminal state
// (completed or failed) or the client disconnects.
func (h *Handler) GetJobStatusWSHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userInfo := auth.GetUserInfoFromContext(ctx)

	jobIDStr := r.PathValue("id")
	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil {
		http.Error(w, "Invalid job ID", http.StatusBadRequest)
		return
	}

	// Authorisation: verify the job belongs to this user before upgrading.
	job, err := h.ExtractionJobStore.GetByID(ctx, jobID)
	if err != nil || job == nil {
		http.Error(w, "Job not found", http.StatusNotFound)
		return
	}
	if job.UserID != userInfo.UserID && !userInfo.IsAdmin {
		http.Error(w, "Access denied", http.StatusForbidden)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logging.AddError(ctx, err, "WebSocket upgrade failed")
		return
	}
	defer conn.Close()

	logging.AddMany(ctx, map[string]any{
		"action": "job.ws.connect",
		"job_id": jobID,
	})

	// Pump: detect disconnects from the client (browser tab close, navigation).
	disconnected := make(chan struct{})
	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				close(disconnected)
				return
			}
		}
	}()

	lastStatus := job.Status
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	isTerminal := func(s string) bool {
		return s == "completed" || s == "failed"
	}

	sendFragment := func(current *jobStatusFragment) error {
		var buf bytes.Buffer
		if err := h.Renderer.Render(&buf, "job-status-fragment", current); err != nil {
			return err
		}
		return conn.WriteMessage(websocket.TextMessage, buf.Bytes())
	}

	// Send the initial state immediately so the page is always in sync.
	feedback, _ := h.ExtractionFeedbackStore.GetByJobID(ctx, jobID)
	if err := sendFragment(&jobStatusFragment{Job: job, Feedback: feedback}); err != nil {
		slog.Debug("WS: initial write failed", "job_id", jobID, "error", err)
		return
	}

	if isTerminal(lastStatus) {
		return
	}

	for {
		select {
		case <-disconnected:
			logging.AddMany(ctx, map[string]any{
				"action": "job.ws.disconnect",
				"job_id": jobID,
			})
			return

		case <-ticker.C:
			latest, err := h.ExtractionJobStore.GetByID(ctx, jobID)
			if err != nil || latest == nil {
				slog.Warn("WS: failed to fetch job", "job_id", jobID, "error", err)
				continue
			}

			if latest.Status != lastStatus {
				lastStatus = latest.Status
				feedback, _ := h.ExtractionFeedbackStore.GetByJobID(ctx, jobID)
				if err := sendFragment(&jobStatusFragment{Job: latest, Feedback: feedback}); err != nil {
					slog.Debug("WS: write failed", "job_id", jobID, "error", err)
					return
				}
				logging.AddMany(ctx, map[string]any{
					"action":     "job.ws.update",
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
