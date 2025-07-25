package models

import "time"

// User represents a user in the system
type User struct {
	ID           int
	Username     string
	PasswordHash string
	CreatedAt    time.Time
}

// Recipe represents a recipe
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
}

// Label represents a label
type Label struct {
	ID   int
	Name string
}

// RecipeLabel represents the many-to-many relationship
type RecipeLabel struct {
	RecipeID int
	LabelID  int
}

// Comment represents a comment on a recipe
type Comment struct {
	ID        int
	RecipeID  int
	AuthorID  int
	ContentMD string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ProposedChange represents a proposed change to a recipe
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
