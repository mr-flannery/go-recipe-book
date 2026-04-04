package handlers

import (
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/mr-flannery/go-recipe-book/src/auth"
	"github.com/mr-flannery/go-recipe-book/src/logging"
	"github.com/mr-flannery/go-recipe-book/src/store"
)

const maxImageSize = 10 * 1024 * 1024 // 10MB

type ExtractPageData struct {
	UserInfo *auth.UserInfo
	Error    string
	Success  string
}

func (h *Handler) GetExtractHandler(w http.ResponseWriter, r *http.Request) {
	userInfo := auth.GetUserInfoFromContext(r.Context())
	data := ExtractPageData{
		UserInfo: userInfo,
		Error:    r.URL.Query().Get("error"),
		Success:  r.URL.Query().Get("success"),
	}
	h.Renderer.RenderPage(w, "extract.gohtml", data)
}

func (h *Handler) PostExtractWebsiteHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userInfo := auth.GetUserInfoFromContext(ctx)

	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/extract?error=Invalid form data", http.StatusSeeOther)
		return
	}

	websiteURL := strings.TrimSpace(r.FormValue("url"))
	if websiteURL == "" {
		http.Redirect(w, r, "/extract?error=Please enter a website URL", http.StatusSeeOther)
		return
	}

	parsedURL, err := url.Parse(websiteURL)
	if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
		http.Redirect(w, r, "/extract?error=Please enter a valid URL (starting with http:// or https://)", http.StatusSeeOther)
		return
	}

	jobID, err := h.ExtractionJobStore.Create(ctx, userInfo.UserID, "website", &websiteURL, nil)
	if err != nil {
		logging.AddError(ctx, err, "Failed to create extraction job")
		http.Redirect(w, r, "/extract?error=Failed to create extraction job", http.StatusSeeOther)
		return
	}

	logging.AddMany(ctx, map[string]any{
		"action":   "extraction.submit",
		"job_id":   jobID,
		"job_type": "website",
	})

	http.Redirect(w, r, "/account/jobs/"+strconv.Itoa(jobID), http.StatusSeeOther)
}

func (h *Handler) PostExtractVideoHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userInfo := auth.GetUserInfoFromContext(ctx)

	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/extract?error=Invalid form data", http.StatusSeeOther)
		return
	}

	videoURL := strings.TrimSpace(r.FormValue("url"))
	if videoURL == "" {
		http.Redirect(w, r, "/extract?error=Please enter a YouTube URL", http.StatusSeeOther)
		return
	}

	if !isYouTubeURL(videoURL) {
		http.Redirect(w, r, "/extract?error=Please enter a valid YouTube URL", http.StatusSeeOther)
		return
	}

	jobID, err := h.ExtractionJobStore.Create(ctx, userInfo.UserID, "video", &videoURL, nil)
	if err != nil {
		logging.AddError(ctx, err, "Failed to create extraction job")
		http.Redirect(w, r, "/extract?error=Failed to create extraction job", http.StatusSeeOther)
		return
	}

	logging.AddMany(ctx, map[string]any{
		"action":   "extraction.submit",
		"job_id":   jobID,
		"job_type": "video",
	})

	http.Redirect(w, r, "/account/jobs/"+strconv.Itoa(jobID), http.StatusSeeOther)
}

func (h *Handler) PostExtractImageHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userInfo := auth.GetUserInfoFromContext(ctx)

	if err := r.ParseMultipartForm(maxImageSize); err != nil {
		http.Redirect(w, r, "/extract?error=Image too large (max 10MB)", http.StatusSeeOther)
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		http.Redirect(w, r, "/extract?error=Please select an image file", http.StatusSeeOther)
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if !isAllowedImageType(contentType) {
		http.Redirect(w, r, "/extract?error=Please upload a JPEG, PNG, GIF, or WebP image", http.StatusSeeOther)
		return
	}

	imageData, err := io.ReadAll(io.LimitReader(file, maxImageSize+1))
	if err != nil {
		http.Redirect(w, r, "/extract?error=Failed to read image", http.StatusSeeOther)
		return
	}

	if len(imageData) > maxImageSize {
		http.Redirect(w, r, "/extract?error=Image too large (max 10MB)", http.StatusSeeOther)
		return
	}

	jobID, err := h.ExtractionJobStore.Create(ctx, userInfo.UserID, "image", nil, imageData)
	if err != nil {
		logging.AddError(ctx, err, "Failed to create extraction job")
		http.Redirect(w, r, "/extract?error=Failed to create extraction job", http.StatusSeeOther)
		return
	}

	logging.AddMany(ctx, map[string]any{
		"action":     "extraction.submit",
		"job_id":     jobID,
		"job_type":   "image",
		"image_size": len(imageData),
	})

	http.Redirect(w, r, "/account/jobs/"+strconv.Itoa(jobID), http.StatusSeeOther)
}

type JobsListData struct {
	UserInfo   *auth.UserInfo
	Jobs       []store.ExtractionJob
	TotalCount int
	Page       int
	PageSize   int
	TotalPages int
}

func (h *Handler) GetAccountJobsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userInfo := auth.GetUserInfoFromContext(ctx)

	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	pageSize := 20
	offset := (page - 1) * pageSize

	jobs, err := h.ExtractionJobStore.GetByUserID(ctx, userInfo.UserID, pageSize, offset)
	if err != nil {
		logging.AddError(ctx, err, "Failed to fetch jobs")
		jobs = []store.ExtractionJob{}
	}

	totalCount, err := h.ExtractionJobStore.CountByUserID(ctx, userInfo.UserID)
	if err != nil {
		logging.AddError(ctx, err, "Failed to count jobs")
	}

	totalPages := (totalCount + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}

	data := JobsListData{
		UserInfo:   userInfo,
		Jobs:       jobs,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
	h.Renderer.RenderPage(w, "account-jobs.gohtml", data)
}

type JobDetailData struct {
	UserInfo *auth.UserInfo
	Job      *store.ExtractionJob
	Feedback *store.ExtractionFeedback
	Error    string
	Success  string
}

func (h *Handler) GetAccountJobDetailHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userInfo := auth.GetUserInfoFromContext(ctx)

	jobIDStr := r.PathValue("id")
	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil {
		h.Renderer.RenderError(w, r, http.StatusBadRequest, "Invalid job ID")
		return
	}

	job, err := h.ExtractionJobStore.GetByID(ctx, jobID)
	if err != nil {
		logging.AddError(ctx, err, "Failed to fetch job")
		h.Renderer.RenderError(w, r, http.StatusInternalServerError, "Failed to load job")
		return
	}

	if job == nil {
		h.Renderer.RenderError(w, r, http.StatusNotFound, "Job not found")
		return
	}

	if job.UserID != userInfo.UserID && !userInfo.IsAdmin {
		h.Renderer.RenderError(w, r, http.StatusForbidden, "Access denied")
		return
	}

	feedback, _ := h.ExtractionFeedbackStore.GetByJobID(ctx, jobID)

	data := JobDetailData{
		UserInfo: userInfo,
		Job:      job,
		Feedback: feedback,
		Error:    r.URL.Query().Get("error"),
		Success:  r.URL.Query().Get("success"),
	}
	h.Renderer.RenderPage(w, "account-job-detail.gohtml", data)
}

func (h *Handler) PostJobFeedbackHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userInfo := auth.GetUserInfoFromContext(ctx)

	jobIDStr := r.PathValue("id")
	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil {
		http.Redirect(w, r, "/account/jobs", http.StatusSeeOther)
		return
	}

	job, err := h.ExtractionJobStore.GetByID(ctx, jobID)
	if err != nil || job == nil {
		http.Redirect(w, r, "/account/jobs", http.StatusSeeOther)
		return
	}

	if job.UserID != userInfo.UserID && !userInfo.IsAdmin {
		http.Redirect(w, r, "/account/jobs", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/account/jobs/"+jobIDStr+"?error=Invalid form data", http.StatusSeeOther)
		return
	}

	rating, err := strconv.Atoi(r.FormValue("rating"))
	if err != nil || rating < 1 || rating > 5 {
		http.Redirect(w, r, "/account/jobs/"+jobIDStr+"?error=Invalid rating", http.StatusSeeOther)
		return
	}

	feedbackType := r.FormValue("feedback_type")
	validTypes := map[string]bool{"good": true, "missing_info": true, "inaccurate": true, "other": true}
	if !validTypes[feedbackType] {
		http.Redirect(w, r, "/account/jobs/"+jobIDStr+"?error=Invalid feedback type", http.StatusSeeOther)
		return
	}

	var comment *string
	if c := strings.TrimSpace(r.FormValue("comment")); c != "" {
		comment = &c
	}

	err = h.ExtractionFeedbackStore.Create(ctx, jobID, userInfo.UserID, rating, feedbackType, comment)
	if err != nil {
		logging.AddError(ctx, err, "Failed to save feedback")
		http.Redirect(w, r, "/account/jobs/"+jobIDStr+"?error=Failed to save feedback", http.StatusSeeOther)
		return
	}

	logging.AddMany(ctx, map[string]any{
		"action":        "extraction.feedback",
		"job_id":        jobID,
		"rating":        rating,
		"feedback_type": feedbackType,
	})

	http.Redirect(w, r, "/account/jobs/"+jobIDStr+"?success=Feedback submitted", http.StatusSeeOther)
}

func (h *Handler) PostJobRetryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userInfo := auth.GetUserInfoFromContext(ctx)

	jobIDStr := r.PathValue("id")
	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil {
		http.Redirect(w, r, "/account/jobs", http.StatusSeeOther)
		return
	}

	job, err := h.ExtractionJobStore.GetByID(ctx, jobID)
	if err != nil || job == nil {
		http.Redirect(w, r, "/account/jobs", http.StatusSeeOther)
		return
	}

	if job.UserID != userInfo.UserID && !userInfo.IsAdmin {
		http.Redirect(w, r, "/account/jobs", http.StatusSeeOther)
		return
	}

	if job.Status != "failed" && !(job.Status == "processing" && time.Since(job.UpdatedAt) > 5*time.Minute) {
		http.Redirect(w, r, "/account/jobs/"+jobIDStr+"?error=Only failed or stuck processing jobs can be retried", http.StatusSeeOther)
		return
	}

	err = h.ExtractionJobStore.ResetForRetry(ctx, jobID)
	if err != nil {
		logging.AddError(ctx, err, "Failed to retry job")
		http.Redirect(w, r, "/account/jobs/"+jobIDStr+"?error=Failed to retry job", http.StatusSeeOther)
		return
	}

	logging.AddMany(ctx, map[string]any{
		"action": "extraction.retry",
		"job_id": jobID,
	})

	http.Redirect(w, r, "/account/jobs/"+jobIDStr+"?success=Job queued for retry", http.StatusSeeOther)
}

func isYouTubeURL(urlStr string) bool {
	patterns := []string{
		"youtube.com/watch",
		"youtu.be/",
		"youtube.com/embed/",
		"youtube.com/v/",
	}
	urlLower := strings.ToLower(urlStr)
	for _, pattern := range patterns {
		if strings.Contains(urlLower, pattern) {
			return true
		}
	}
	return false
}

func isAllowedImageType(contentType string) bool {
	allowed := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	return allowed[contentType]
}

type AdminJobsData struct {
	UserInfo   *auth.UserInfo
	Jobs       []store.ExtractionJob
	TotalCount int
	Page       int
	PageSize   int
	TotalPages int
}

func (h *Handler) GetAdminJobsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userInfo := auth.GetUserInfoFromContext(ctx)

	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	pageSize := 50
	offset := (page - 1) * pageSize

	jobs, err := h.ExtractionJobStore.GetAll(ctx, pageSize, offset)
	if err != nil {
		logging.AddError(ctx, err, "Failed to fetch jobs")
		jobs = []store.ExtractionJob{}
	}

	totalCount, err := h.ExtractionJobStore.CountAll(ctx)
	if err != nil {
		logging.AddError(ctx, err, "Failed to count jobs")
	}

	totalPages := (totalCount + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}

	data := AdminJobsData{
		UserInfo:   userInfo,
		Jobs:       jobs,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
	h.Renderer.RenderPage(w, "admin-jobs.gohtml", data)
}

type AdminFeedbackData struct {
	UserInfo   *auth.UserInfo
	Feedback   []store.ExtractionFeedback
	TotalCount int
	Page       int
	PageSize   int
	TotalPages int
}

func (h *Handler) GetAdminFeedbackHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userInfo := auth.GetUserInfoFromContext(ctx)

	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	pageSize := 50
	offset := (page - 1) * pageSize

	feedback, err := h.ExtractionFeedbackStore.GetAll(ctx, pageSize, offset)
	if err != nil {
		logging.AddError(ctx, err, "Failed to fetch feedback")
		feedback = []store.ExtractionFeedback{}
	}

	totalCount, err := h.ExtractionFeedbackStore.CountAll(ctx)
	if err != nil {
		logging.AddError(ctx, err, "Failed to count feedback")
	}

	totalPages := (totalCount + pageSize - 1) / pageSize
	if totalPages < 1 {
		totalPages = 1
	}

	data := AdminFeedbackData{
		UserInfo:   userInfo,
		Feedback:   feedback,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
		TotalPages: totalPages,
	}
	h.Renderer.RenderPage(w, "admin-feedback.gohtml", data)
}
