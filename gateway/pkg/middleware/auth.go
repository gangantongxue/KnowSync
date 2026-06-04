// Package middleware provides HTTP middleware for authentication and authorization.
package middleware

import (
	"context"
	"crypto/rsa"
	"strings"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/gangantongxue/knowsync/gateway/pkg/errcode"
	"github.com/gangantongxue/knowsync/gateway/pkg/response"
	"github.com/golang-jwt/jwt/v5"
)

// AccessTokenClaims represents the JWT claims for an access token.
type AccessTokenClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// Auth returns a middleware that verifies JWT access tokens using the given RSA public key.
func Auth(publicKey *rsa.PublicKey) app.HandlerFunc {
	return func(c context.Context, ctx *app.RequestContext) {
		authHeader := string(ctx.GetHeader("Authorization"))
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			response.Error(c, ctx, 401, errcode.ErrUnauth, "未授权")
			ctx.Abort()
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.ParseWithClaims(tokenString, &AccessTokenClaims{}, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return publicKey, nil
		})
		if err != nil {
			response.Error(c, ctx, 401, errcode.ErrUnauth, "令牌无效或已过期")
			ctx.Abort()
			return
		}

		claims, ok := token.Claims.(*AccessTokenClaims)
		if !ok || !token.Valid {
			response.Error(c, ctx, 401, errcode.ErrUnauth, "令牌无效或已过期")
			ctx.Abort()
			return
		}

		ctx.Set("user_id", claims.UserID)
		ctx.Next(c)
	}
}
