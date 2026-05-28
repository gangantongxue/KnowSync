package auth

import (
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func testKeys(t *testing.T) (*rsa.PrivateKey, *rsa.PublicKey) {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey() error = %v", err)
	}
	return key, &key.PublicKey
}

func TestGenerateAccessToken(t *testing.T) {
	privKey, _ := testKeys(t)
	ttl := 15 * time.Minute

	t.Run("generates valid token", func(t *testing.T) {
		userID := "user123"
		token, err := GenerateAccessToken(userID, privKey, ttl)
		if err != nil {
			t.Fatalf("GenerateAccessToken() error = %v", err)
		}
		if token == "" {
			t.Fatal("GenerateAccessToken() returned empty token")
		}
	})

	t.Run("different user IDs produce different tokens", func(t *testing.T) {
		t1, _ := GenerateAccessToken("user1", privKey, ttl)
		t2, _ := GenerateAccessToken("user2", privKey, ttl)
		if t1 == t2 {
			t.Error("GenerateAccessToken() should produce different tokens for different user IDs")
		}
	})
}

func TestValidateAccessToken(t *testing.T) {
	privKey, pubKey := testKeys(t)
	ttl := 15 * time.Minute

	t.Run("valid token", func(t *testing.T) {
		userID := "user123"
		token, err := GenerateAccessToken(userID, privKey, ttl)
		if err != nil {
			t.Fatalf("GenerateAccessToken() error = %v", err)
		}

		gotUserID, err := ValidateAccessToken(token, pubKey)
		if err != nil {
			t.Fatalf("ValidateAccessToken() error = %v", err)
		}
		if gotUserID != userID {
			t.Errorf("ValidateAccessToken() userID = %s, want %s", gotUserID, userID)
		}
	})

	t.Run("wrong key", func(t *testing.T) {
		userID := "user123"
		token, err := GenerateAccessToken(userID, privKey, ttl)
		if err != nil {
			t.Fatalf("GenerateAccessToken() error = %v", err)
		}

		wrongKey, _ := rsa.GenerateKey(rand.Reader, 2048)
		_, err = ValidateAccessToken(token, &wrongKey.PublicKey)
		if err == nil {
			t.Error("ValidateAccessToken() expected error for wrong key, got nil")
		}
	})

	t.Run("expired token", func(t *testing.T) {
		userID := "user123"
		token, err := GenerateAccessToken(userID, privKey, -time.Minute)
		if err != nil {
			t.Fatalf("GenerateAccessToken() error = %v", err)
		}

		_, err = ValidateAccessToken(token, pubKey)
		if err == nil {
			t.Error("ValidateAccessToken() expected error for expired token, got nil")
		}
	})

	t.Run("malformed token string", func(t *testing.T) {
		_, err := ValidateAccessToken("not-a-valid-jwt", pubKey)
		if err == nil {
			t.Error("ValidateAccessToken() expected error for malformed token, got nil")
		}
	})

	t.Run("none signing algorithm", func(t *testing.T) {
		claims := AccessTokenClaims{
			UserID: "user123",
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				Issuer:    "knowsync-user-server",
			},
		}
		token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
		tokenString, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
		if err != nil {
			t.Fatalf("failed to sign token with none algorithm: %v", err)
		}

		_, err = ValidateAccessToken(tokenString, pubKey)
		if err == nil {
			t.Error("ValidateAccessToken() expected error for none algorithm, got nil")
		}
	})

	t.Run("empty token", func(t *testing.T) {
		_, err := ValidateAccessToken("", pubKey)
		if err == nil {
			t.Error("ValidateAccessToken() expected error for empty token, got nil")
		}
	})
}
