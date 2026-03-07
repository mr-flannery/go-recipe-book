package models

import (
	"encoding/base64"
	"strings"
	"time"
	"unicode/utf8"
)

type User struct {
	ID           int
	Username     string
	Email        string
	PasswordHash string
	IsAdmin      bool
	IsActive     bool
	LastLogin    *time.Time
	CreatedAt    time.Time
}

type Recipe struct {
	ID             int
	Title          string
	IngredientsMD  string
	InstructionsMD string
	PrepTime       int
	CookTime       int
	Calories       int
	AuthorID       int
	Image          []byte
	ParentID       *int
	CreatedAt      time.Time
	UpdatedAt      time.Time
	Tags           []Tag
	UserTags       []UserTag
}

func (r Recipe) ImageBase64() string {
	if len(r.Image) == 0 {
		return ""
	}
	return base64.StdEncoding.EncodeToString(r.Image)
}

func (r Recipe) TotalTime() int {
	return r.PrepTime + r.CookTime
}

func (r Recipe) IngredientCount() int {
	if r.IngredientsMD == "" {
		return 0
	}
	count := 0
	lines := strings.Split(r.IngredientsMD, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "-") || strings.HasPrefix(trimmed, "*") {
			count++
		} else if len(trimmed) > 0 && trimmed[0] >= '0' && trimmed[0] <= '9' {
			count++
		}
	}
	return count
}

func (r Recipe) Summary() string {
	if r.InstructionsMD == "" {
		return ""
	}
	text := strings.TrimSpace(r.InstructionsMD)
	lines := strings.SplitN(text, "\n", 2)
	firstLine := strings.TrimSpace(lines[0])
	firstLine = strings.TrimPrefix(firstLine, "#")
	firstLine = strings.TrimPrefix(firstLine, "-")
	firstLine = strings.TrimPrefix(firstLine, "*")
	firstLine = strings.TrimSpace(firstLine)
	if len(firstLine) > 0 && firstLine[0] >= '0' && firstLine[0] <= '9' {
		if idx := strings.Index(firstLine, "."); idx != -1 && idx < 4 {
			firstLine = strings.TrimSpace(firstLine[idx+1:])
		}
	}
	if utf8.RuneCountInString(firstLine) <= 80 {
		return firstLine
	}
	runes := []rune(firstLine)
	truncated := string(runes[:80])
	if lastSpace := strings.LastIndex(truncated, " "); lastSpace > 60 {
		truncated = truncated[:lastSpace]
	}
	return truncated + "..."
}

type Tag struct {
	ID   int
	Name string
}

type UserTag struct {
	ID       int
	UserID   int
	RecipeID int
	Name     string
}

type Comment struct {
	ID        int
	RecipeID  int
	AuthorID  int
	ContentMD string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ProposedChange struct {
	ID             int
	RecipeID       int
	ProposerID     int
	Title          string
	IngredientsMD  string
	InstructionsMD string
	PrepTime       int
	CookTime       int
	Calories       int
	Image          []byte
	CreatedAt      time.Time
	Status         string // pending, accepted, rejected
}

type FilterParams struct {
	Search        string
	CaloriesOp    string
	CaloriesValue int
	PrepTimeOp    string
	PrepTimeValue int
	CookTimeOp    string
	CookTimeValue int
	Tags          []string
	UserID        int
	UserTags      []string
	AuthorID      int
	Limit         int
	Offset        int
}

type RecipeSearchResult struct {
	ID    int
	Title string
}

type UserPreferences struct {
	UserID   int
	PageSize int
	ViewMode string
}

const (
	DefaultPageSize = 20
	ViewModeGrid    = "grid"
	ViewModeList    = "list"
	DefaultViewMode = ViewModeGrid
)
