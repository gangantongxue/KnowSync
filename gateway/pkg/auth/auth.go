// Package auth provides RSA public key loading for JWT token verification.
package auth

import (
	"crypto/rsa"
	"fmt"
	"os"

	"github.com/golang-jwt/jwt/v5"
)

// LoadPublicKeyFromFile reads an RSA public key from the specified file path.
func LoadPublicKeyFromFile(path string) (*rsa.PublicKey, error) {
	keyData, err := os.ReadFile(path) //nolint:gosec // path is from config, not user input
	if err != nil {
		return nil, fmt.Errorf("读取 RSA 公钥文件失败: %w", err)
	}
	return jwt.ParseRSAPublicKeyFromPEM(keyData)
}
