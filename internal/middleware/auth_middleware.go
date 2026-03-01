package middleware

import (
	"airline-tracker/internal/service"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		header := ctx.GetHeader("Authorization")
		if header == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, "unauthorized")
			return
		}

		tokenString := strings.TrimPrefix(header, "Bearer ")

		claims := &service.JWTClaims{}

		token, err := jwt.ParseWithClaims(
			tokenString, claims,
			func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, fmt.Errorf("invalid signing method")
				}
				return service.GetJWTKey(), nil
			})

		if err != nil || !token.Valid {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, "unauthorized")
			return
		}

		if claims.Issuer != "auth-service" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, "unauthorized")
			return
		}

		ctx.Set("user_id", claims.Subject)
		ctx.Set("role", claims.Role)

		ctx.Next()
	}
}
