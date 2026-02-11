package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"errors"
	"fmt"
	"regexp"
	"unicode"

	"golang.org/x/crypto/argon2"
)

const (
	// Argon2id parameters - recommended for password hashing
	argon2Time    = 3
	argon2Memory  = 64 * 1024 // 64MB
	argon2Threads = 4
	argon2KeyLen  = 32
	saltLength    = 16
)

var (
	ErrPasswordTooShort    = errors.New("password must be at least 12 characters long")
	ErrPasswordTooWeak     = errors.New("password must contain uppercase, lowercase, numbers, and symbols")
	ErrInvalidPasswordHash = errors.New("invalid password hash format")
)

type PasswordHash struct {
	Hash    []byte
	Salt    []byte
	Time    uint32
	Memory  uint32
	Threads uint8
	KeyLen  uint32
}

func HashPassword(password string) (string, error) {
	if err := ValidatePasswordStrength(password); err != nil {
		return "", err
	}

	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, argon2Time, argon2Memory, argon2Threads, argon2KeyLen)

	encoded := fmt.Sprintf("$argon2id$v=19$m=%d,t=%d,p=%d$%x$%x",
		argon2Memory, argon2Time, argon2Threads, salt, hash)

	return encoded, nil
}

// VerifyPassword verifies a password against its hash using constant-time comparison
func VerifyPassword(password, encodedHash string) error {
	// Parse the encoded hash
	hashInfo, err := parsePasswordHash(encodedHash)
	if err != nil {
		return err
	}

	// Generate hash with same parameters
	computedHash := argon2.IDKey([]byte(password), hashInfo.Salt,
		hashInfo.Time, hashInfo.Memory, hashInfo.Threads, hashInfo.KeyLen)

	// Use constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare(hashInfo.Hash, computedHash) == 1 {
		return nil
	}

	return errors.New("invalid password")
}

// ValidatePasswordStrength ensures password meets security requirements
func ValidatePasswordStrength(password string) error {
	if len(password) < 12 {
		return ErrPasswordTooShort
	}

	var (
		hasUpper  = false
		hasLower  = false
		hasNumber = false
		hasSymbol = false
	)

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSymbol = true
		}
	}

	if !hasUpper || !hasLower || !hasNumber || !hasSymbol {
		return ErrPasswordTooWeak
	}

	return nil
}

// parsePasswordHash parses an encoded Argon2id hash string
func parsePasswordHash(encodedHash string) (*PasswordHash, error) {
	// Expected format: $argon2id$v=19$m=65536,t=3,p=4$salt$hash
	re := regexp.MustCompile(`^\$argon2id\$v=19\$m=(\d+),t=(\d+),p=(\d+)\$([a-f0-9]+)\$([a-f0-9]+)$`)
	matches := re.FindStringSubmatch(encodedHash)

	if len(matches) != 6 {
		return nil, ErrInvalidPasswordHash
	}

	var memory, time, threads uint32
	if _, err := fmt.Sscanf(matches[1], "%d", &memory); err != nil {
		return nil, ErrInvalidPasswordHash
	}
	if _, err := fmt.Sscanf(matches[2], "%d", &time); err != nil {
		return nil, ErrInvalidPasswordHash
	}
	if _, err := fmt.Sscanf(matches[3], "%d", &threads); err != nil {
		return nil, ErrInvalidPasswordHash
	}

	salt := make([]byte, len(matches[4])/2)
	if _, err := fmt.Sscanf(matches[4], "%x", &salt); err != nil {
		return nil, ErrInvalidPasswordHash
	}

	hash := make([]byte, len(matches[5])/2)
	if _, err := fmt.Sscanf(matches[5], "%x", &hash); err != nil {
		return nil, ErrInvalidPasswordHash
	}

	return &PasswordHash{
		Hash:    hash,
		Salt:    salt,
		Time:    time,
		Memory:  memory,
		Threads: uint8(threads),
		KeyLen:  uint32(len(hash)),
	}, nil
}
