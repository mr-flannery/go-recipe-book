package auth

import (
	"strings"
	"testing"
)

func TestHashPassword_Success(t *testing.T) {
	password := "ValidPassword123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if hash == "" {
		t.Fatal("expected hash, got empty string")
	}
	if !strings.HasPrefix(hash, "$argon2id$v=19$") {
		t.Errorf("expected hash to start with '$argon2id$v=19$', got %s", hash[:20])
	}
}

func TestHashPassword_WeakPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{"too short", "Short1!", ErrPasswordTooShort},
		{"no uppercase", "alllowercase123!", ErrPasswordTooWeak},
		{"no lowercase", "ALLUPPERCASE123!", ErrPasswordTooWeak},
		{"no numbers", "NoNumbersHere!@#", ErrPasswordTooWeak},
		{"no symbols", "NoSymbols12345Ab", ErrPasswordTooWeak},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := HashPassword(tt.password)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if err != tt.wantErr {
				t.Errorf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestHashPassword_DifferentSalts(t *testing.T) {
	password := "SamePassword123!"

	hash1, _ := HashPassword(password)
	hash2, _ := HashPassword(password)

	if hash1 == hash2 {
		t.Error("expected different hashes due to random salt")
	}
}

func TestVerifyPassword_Success(t *testing.T) {
	password := "MySecurePassword123!"
	hash, _ := HashPassword(password)

	err := VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestVerifyPassword_WrongPassword(t *testing.T) {
	password := "CorrectPassword123!"
	hash, _ := HashPassword(password)

	err := VerifyPassword("WrongPassword123!", hash)
	if err == nil {
		t.Fatal("expected error for wrong password, got nil")
	}
}

func TestVerifyPassword_InvalidHash(t *testing.T) {
	tests := []struct {
		name string
		hash string
	}{
		{"empty hash", ""},
		{"random string", "not-a-valid-hash"},
		{"wrong prefix", "$bcrypt$v=19$m=65536,t=3,p=4$abc123$def456"},
		{"missing parts", "$argon2id$v=19$"},
		{"invalid hex", "$argon2id$v=19$m=65536,t=3,p=4$ZZZZ$YYYY"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := VerifyPassword("AnyPassword123!", tt.hash)
			if err == nil {
				t.Fatal("expected error for invalid hash, got nil")
			}
		})
	}
}

func TestValidatePasswordStrength(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{"valid password", "ValidPassword123!", nil},
		{"too short", "Short1!", ErrPasswordTooShort},
		{"exactly 12 chars valid", "ValidPass12!", nil},
		{"11 chars invalid", "ValidPas12!", ErrPasswordTooShort},
		{"no uppercase", "alllowercase123!", ErrPasswordTooWeak},
		{"no lowercase", "ALLUPPERCASE123!", ErrPasswordTooWeak},
		{"no numbers", "NoNumbersHere!@#", ErrPasswordTooWeak},
		{"no symbols", "NoSymbols12345Ab", ErrPasswordTooWeak},
		{"unicode symbols work", "ValidPassword123@", nil},
		{"with spaces", "Valid Pass 123!", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePasswordStrength(tt.password)
			if err != tt.wantErr {
				t.Errorf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestParsePasswordHash_Valid(t *testing.T) {
	password := "TestPassword123!"
	hash, _ := HashPassword(password)

	parsed, err := parsePasswordHash(hash)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if parsed == nil {
		t.Fatal("expected parsed hash, got nil")
	}
	if parsed.Memory != argon2Memory {
		t.Errorf("expected memory %d, got %d", argon2Memory, parsed.Memory)
	}
	if parsed.Time != argon2Time {
		t.Errorf("expected time %d, got %d", argon2Time, parsed.Time)
	}
	if parsed.Threads != argon2Threads {
		t.Errorf("expected threads %d, got %d", argon2Threads, parsed.Threads)
	}
	if len(parsed.Salt) != saltLength {
		t.Errorf("expected salt length %d, got %d", saltLength, len(parsed.Salt))
	}
	if len(parsed.Hash) != argon2KeyLen {
		t.Errorf("expected hash length %d, got %d", argon2KeyLen, len(parsed.Hash))
	}
}

func TestParsePasswordHash_Invalid(t *testing.T) {
	tests := []struct {
		name string
		hash string
	}{
		{"empty", ""},
		{"random", "not-a-hash"},
		{"wrong algorithm", "$bcrypt$v=19$m=65536,t=3,p=4$abc123$def456"},
		{"incomplete", "$argon2id$v=19$"},
		{"bad params format", "$argon2id$v=19$invalid$abc$def"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parsePasswordHash(tt.hash)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func BenchmarkHashPassword(b *testing.B) {
	password := "BenchmarkPassword123!"
	for i := 0; i < b.N; i++ {
		HashPassword(password)
	}
}

func BenchmarkVerifyPassword(b *testing.B) {
	password := "BenchmarkPassword123!"
	hash, _ := HashPassword(password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		VerifyPassword(password, hash)
	}
}
