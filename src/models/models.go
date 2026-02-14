package models

import (
	"encoding/base64"
	"time"
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
	Limit         int
	Offset        int
}
