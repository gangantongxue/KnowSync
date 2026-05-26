package service

import (
	"strings"
	"testing"
)

func TestHashPassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"normal password", "password123", false},
		{"short password", "abc123", false},
		{"password with spaces", "my password", false},
		{"special characters", "p@ssw0rd!#$", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hash, err := HashPassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("HashPassword() error = %v, wantErr = %v", err, tt.wantErr)
				return
			}
			if hash == "" {
				t.Error("HashPassword() returned empty hash")
			}
			if !strings.HasPrefix(hash, "$2a$") {
				t.Errorf("HashPassword() = %s, want bcrypt hash starting with $2a$", hash)
			}
		})
	}
}

func TestHashPassword_SamePasswordDifferentHash(t *testing.T) {
	password := "testpassword"

	h1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	h2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	if h1 == h2 {
		t.Error("HashPassword() with same password should produce different hashes (random salt)")
	}
}

func TestCheckPassword(t *testing.T) {
	password := "securePassword123!"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword() error = %v", err)
	}

	t.Run("correct password", func(t *testing.T) {
		if err := CheckPassword(password, hash); err != nil {
			t.Errorf("CheckPassword() error = %v, want nil", err)
		}
	})

	t.Run("incorrect password", func(t *testing.T) {
		if err := CheckPassword("wrongpassword", hash); err == nil {
			t.Error("CheckPassword() expected error, got nil")
		}
	})

	t.Run("empty password", func(t *testing.T) {
		if err := CheckPassword("", hash); err == nil {
			t.Error("CheckPassword() with empty password expected error, got nil")
		}
	})

	t.Run("empty hash", func(t *testing.T) {
		if err := CheckPassword(password, ""); err == nil {
			t.Error("CheckPassword() with empty hash expected error, got nil")
		}
	})

	t.Run("invalid hash format", func(t *testing.T) {
		if err := CheckPassword(password, "invalidhash"); err == nil {
			t.Error("CheckPassword() with invalid hash expected error, got nil")
		}
	})
}
