package middleware

import (
	"context"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/gangantongxue/knowsync/gateway/pkg/response"
	"github.com/golang-jwt/jwt/v5"
)

// AccessTokenClaims JWT access token 声明，与 user-server 保持一致
type AccessTokenClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// Auth JWT Bearer 认证中间件
// 从 Authorization 头中提取并校验 access token，通过后将 user_id 注入上下文
func Auth(jwtSecret string) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		authHeader := string(ctx.GetHeader("Authorization"))
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			response.Error(c, ctx, 401, 40100, "未授权")
			ctx.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(jwtSecret), nil
		})
		if err != nil {
			response.Error(c, ctx, 401, 40100, "令牌无效或已过期")
			ctx.Abort()
			return
		}

		claims, ok := token.Claims.(*AccessTokenClaims)
		if !ok || !token.Valid {
			response.Error(c, ctx, 401, 40100, "令牌无效或已过期")
			ctx.Abort()
			return
		}

		ctx.Set("user_id", claims.UserID)
		ctx.Next(c)
	}
}
