package middleware

import (
	"log/slog"
	"os"
	"shared/observability"
	"time"

	"github.com/gin-gonic/gin"
)

func Observe() gin.HandlerFunc {
	instance := os.Getenv("INSTANCE_ID")
	return func(ctx *gin.Context) {
		started := time.Now()
		ctx.Next()
		elapsed := time.Since(started)
		route := ctx.FullPath()
		observability.ObserveHTTP(ctx.Request.Method, route, ctx.Writer.Status(), elapsed)
		if route == "" {
			route = "unmatched"
		}
		slog.Info("http request", "instance", instance, "method", ctx.Request.Method,
			"route", route, "status", ctx.Writer.Status(), "duration", elapsed)
	}
}
