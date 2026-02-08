package store

import "github.com/mr-flannery/go-recipe-book/src/models"

type RecipeStore interface {
	Save(recipe models.Recipe) (int, error)
	GetByID(id string) (models.Recipe, error)
	Update(recipe models.Recipe) error
	Delete(id string) error
	GetAll() ([]models.Recipe, error)
	GetFiltered(params models.FilterParams) ([]models.Recipe, error)
	GetRandomID() (int, error)
}

type TagStore interface {
	GetOrCreate(name string) (models.Tag, error)
	Search(query string) ([]models.Tag, error)
	GetByRecipeID(recipeID int) ([]models.Tag, error)
	GetForRecipes(recipeIDs []int) (map[int][]models.Tag, error)
	AddToRecipe(recipeID, tagID int) error
	RemoveFromRecipe(recipeID, tagID int) error
	SetRecipeTags(recipeID int, tagNames []string) error
}

type UserTagStore interface {
	GetOrCreate(userID, recipeID int, name string) (models.UserTag, error)
	Search(userID int, query string) ([]string, error)
	GetByRecipeID(userID, recipeID int) ([]models.UserTag, error)
	Remove(userID, tagID int) error
}

type CommentStore interface {
	GetByRecipeID(recipeID string) ([]models.Comment, error)
	Save(comment models.Comment) error
	GetLatestByUserAndRecipe(userID, recipeID int) (models.Comment, error)
}

type UserStore interface {
	GetUsernameByID(userID int) (string, error)
}

type AuthUser struct {
	ID       int
	Username string
	Email    string
	IsAdmin  bool
	IsActive bool
}

type Session struct {
	ID        string
	UserID    int
	IPAddress string
	UserAgent string
}

type RegistrationRequest struct {
	ID           int
	Username     string
	Email        string
	PasswordHash string
	Status       string
}

type AuthStore interface {
	GetUserByEmail(email string) (*AuthUser, string, error) // returns user, passwordHash, error
	UpdateLastLogin(userID int) error
	GetUserByID(userID int) (*AuthUser, error)
	GetUserIDByUsername(username string) (int, error)

	CreateSession(session *Session) error
	GetSession(sessionID string) (*Session, error)
	DeleteSession(sessionID string) error
	DeleteExpiredSessions() (int64, error)
	DeleteUserSessions(userID int) error
	GetActiveSessionCount(userID int) (int, error)
	ExtendSession(sessionID string) error

	CreateRegistrationRequest(username, email, passwordHash string) error
	GetPendingRegistrations() ([]RegistrationRequest, error)
	ApproveRegistration(requestID, adminID int) error
	RejectRegistration(requestID, adminID int, reason string) error

	CreateUser(username, email, passwordHash string, isAdmin bool) error
	UserExists(username string) (bool, error)
}
