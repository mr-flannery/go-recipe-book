package store

import (
	"context"
	"time"

	"github.com/mr-flannery/go-recipe-book/src/models"
)

type RecipeStore interface {
	Save(ctx context.Context, recipe models.Recipe) (int, error)
	GetByID(ctx context.Context, id string) (models.Recipe, error)
	Update(ctx context.Context, recipe models.Recipe) error
	Delete(ctx context.Context, id string) error
	GetAll(ctx context.Context) ([]models.Recipe, error)
	GetFiltered(ctx context.Context, params models.FilterParams) ([]models.Recipe, error)
	CountFiltered(ctx context.Context, params models.FilterParams) (int, error)
	GetRandomID(ctx context.Context) (int, error)
	SearchByTitle(ctx context.Context, query string, limit int) ([]models.RecipeSearchResult, error)
}

type TagStore interface {
	GetOrCreate(ctx context.Context, name string) (models.Tag, error)
	Search(ctx context.Context, query string) ([]models.Tag, error)
	GetByRecipeID(ctx context.Context, recipeID int) ([]models.Tag, error)
	GetForRecipes(ctx context.Context, recipeIDs []int) (map[int][]models.Tag, error)
	AddToRecipe(ctx context.Context, recipeID, tagID int) error
	RemoveFromRecipe(ctx context.Context, recipeID, tagID int) error
	SetRecipeTags(ctx context.Context, recipeID int, tagNames []string) error
}

type UserTagStore interface {
	GetOrCreate(ctx context.Context, userID, recipeID int, name string) (models.UserTag, error)
	Search(ctx context.Context, userID int, query string) ([]string, error)
	GetByRecipeID(ctx context.Context, userID, recipeID int) ([]models.UserTag, error)
	GetByUserID(ctx context.Context, userID int) ([]models.UserTag, error)
	GetForRecipes(ctx context.Context, userID int, recipeIDs []int) (map[int][]models.UserTag, error)
	Remove(ctx context.Context, userID, tagID int) error
}

type CommentStore interface {
	GetByRecipeID(ctx context.Context, recipeID string) ([]models.Comment, error)
	GetByID(ctx context.Context, commentID int) (models.Comment, error)
	GetByUserID(ctx context.Context, userID int) ([]models.Comment, error)
	Save(ctx context.Context, comment models.Comment) error
	Update(ctx context.Context, commentID int, content string) error
	Delete(ctx context.Context, commentID int) error
	GetLatestByUserAndRecipe(ctx context.Context, userID, recipeID int) (models.Comment, error)
}

type UserStore interface {
	GetUsernameByID(ctx context.Context, userID int) (string, error)
}

type IngredientStore interface {
	Search(ctx context.Context, query string, limit int) ([]string, error)
	GetOrCreate(ctx context.Context, name string) (int, error)
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
	RequestedAt  time.Time
	Status       string
}

type FullAuthUser struct {
	ID        int
	Username  string
	Email     string
	IsAdmin   bool
	IsActive  bool
	CreatedAt time.Time
	LastLogin *time.Time
}

type AuthStore interface {
	GetUserByEmail(ctx context.Context, email string) (*AuthUser, string, error)
	UpdateLastLogin(ctx context.Context, userID int) error
	GetUserByID(ctx context.Context, userID int) (*AuthUser, error)
	GetFullUserByID(ctx context.Context, userID int) (*FullAuthUser, error)
	GetUserIDByUsername(ctx context.Context, username string) (int, error)

	CreateSession(ctx context.Context, session *Session) error
	GetSession(ctx context.Context, sessionID string) (*Session, error)
	DeleteSession(ctx context.Context, sessionID string) error
	DeleteExpiredSessions(ctx context.Context) (int64, error)
	DeleteUserSessions(ctx context.Context, userID int) error
	GetActiveSessionCount(ctx context.Context, userID int) (int, error)
	ExtendSession(ctx context.Context, sessionID string) error

	CreateRegistrationRequest(ctx context.Context, username, email, passwordHash string) error
	GetPendingRegistrations(ctx context.Context) ([]RegistrationRequest, error)
	GetAllRegistrations(ctx context.Context) ([]RegistrationRequest, error)
	GetAllRegistrationsPaginated(ctx context.Context, limit, offset int) ([]RegistrationRequest, error)
	CountAllRegistrations(ctx context.Context) (int, error)
	ApproveRegistration(ctx context.Context, requestID, adminID int) error
	RejectRegistration(ctx context.Context, requestID, adminID int) error

	CreateUser(ctx context.Context, username, email, passwordHash string, isAdmin bool) error
	UserExists(ctx context.Context, username string) (bool, error)

	GetAllUsers(ctx context.Context) ([]AuthUser, error)
	DeleteUser(ctx context.Context, userID int) error

	CreatePasswordResetToken(ctx context.Context, userID int, tokenHash string, expiresAt time.Time) error
	GetPasswordResetToken(ctx context.Context, tokenHash string) (*PasswordResetToken, error)
	MarkPasswordResetTokenUsed(ctx context.Context, tokenHash string) error
	DeleteExpiredPasswordResetTokens(ctx context.Context) (int64, error)
	UpdateUserPassword(ctx context.Context, userID int, passwordHash string) error
	ResetPasswordWithToken(ctx context.Context, tokenHash string, newPasswordHash string) (int, error)
}

type UserPreferencesStore interface {
	Get(ctx context.Context, userID int) (*models.UserPreferences, error)
	SetPageSize(ctx context.Context, userID, pageSize int) error
	SetViewMode(ctx context.Context, userID int, viewMode string) error
	SetTheme(ctx context.Context, userID int, theme string) error
}

type PasswordResetToken struct {
	ID        int
	UserID    int
	ExpiresAt time.Time
	UsedAt    *time.Time
}

type APIKey struct {
	ID           int
	UserID       int
	Name         string
	KeyPrefix    string
	EncryptedKey string
	CreatedAt    time.Time
	LastUsedAt   *time.Time
}

type APIKeyStore interface {
	Create(ctx context.Context, userID int, name string, keyHash string, keyPrefix string, encryptedKey string) (int, error)
	GetByKeyHash(ctx context.Context, keyHash string) (*APIKey, error)
	GetByUserID(ctx context.Context, userID int) ([]APIKey, error)
	Delete(ctx context.Context, userID int, keyID int) error
	UpdateLastUsed(ctx context.Context, keyID int) error
}
