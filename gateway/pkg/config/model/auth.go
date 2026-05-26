package model

// AuthCfg 认证配置
type AuthCfg struct {
	JWTSecret string `yaml:"jwt_secret"` // JWT 签名密钥，用于验证 access token
}
