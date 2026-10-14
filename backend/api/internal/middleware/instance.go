package middleware

import "github.com/gin-gonic/gin"

func Instance(id string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if id != "" {
			ctx.Header("X-Backend-Instance", id)
		}
		ctx.Next()
	}
}
