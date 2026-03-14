package extraction

import "time"

type JobType string

const (
	JobTypeWebsite JobType = "website"
	JobTypeVideo   JobType = "video"
	JobTypeImage   JobType = "image"
)

type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
)

type FeedbackType string

const (
	FeedbackTypeGood        FeedbackType = "good"
	FeedbackTypeMissingInfo FeedbackType = "missing_info"
	FeedbackTypeInaccurate  FeedbackType = "inaccurate"
	FeedbackTypeOther       FeedbackType = "other"
)

type Job struct {
	ID           int
	UserID       int
	JobType      JobType
	InputURL     *string
	InputData    []byte
	Status       JobStatus
	ErrorMessage *string
	LLMInput     *string
	LLMOutput    *string
	RecipeID     *int
	AttemptCount int
	CreatedAt    time.Time
	UpdatedAt    time.Time
	CompletedAt  *time.Time
}

type Feedback struct {
	ID           int
	JobID        int
	UserID       int
	Rating       int
	FeedbackType FeedbackType
	Comment      *string
	CreatedAt    time.Time
}

type ExtractedRecipe struct {
	Title              string   `json:"title"`
	Description        string   `json:"description"`
	IngredientsMD      string   `json:"ingredients_md"`
	InstructionsMD     string   `json:"instructions_md"`
	PrepTimeMinutes    *int     `json:"prep_time_minutes"`
	CookTimeMinutes    *int     `json:"cook_time_minutes"`
	CaloriesPerServing *int     `json:"calories_per_serving"`
	SuggestedTags      []string `json:"suggested_tags"`
	Confidence         float64  `json:"confidence"`
	ConfidenceNotes    string   `json:"confidence_notes"`
}
