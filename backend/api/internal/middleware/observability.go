package middleware

import (
	"log/slog"
	"net/http"
	"regexp"
	"shared/observability"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var requestIDPattern = regexp.MustCompile(`^[A-Za-z0-9._-]{1,128}$`)

func validRequestID(value string) string {
	if requestIDPattern.MatchString(value) {
		return value
	}
	return uuid.NewString()
}

func Correlate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		id := validRequestID(request.Header.Get("X-Request-ID"))
		request.Header.Set("X-Request-ID", id)
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, request)
	})
}

func Observe() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		started := time.Now()
		requestID := ctx.GetHeader("X-Request-ID")
		parent := observability.Extract(ctx.Request.Context(), map[string]string{
			"traceparent": ctx.GetHeader("traceparent"),
			"tracestate":  ctx.GetHeader("tracestate"),
		})
		parent = observability.WithCorrelationID(parent, requestID)
		route := ctx.FullPath()
		if route == "" {
			route = "unmatched"
		}
		requestContext, span := observability.Start(
			parent,
			ctx.Request.Method+" "+route,
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()
		ctx.Request = ctx.Request.WithContext(requestContext)
		ctx.Next()
		elapsed := time.Since(started)
		status := ctx.Writer.Status()
		observability.ObserveHTTP(ctx.Request.Method, route, status, elapsed)
		span.SetAttributes(
			attribute.String("http.request.method", ctx.Request.Method),
			attribute.String("http.route", route),
			attribute.Int("http.response.status_code", status),
		)
		if status >= http.StatusInternalServerError {
			span.SetStatus(codes.Error, http.StatusText(status))
		}
		if err := requestContext.Err(); err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		slog.InfoContext(requestContext, "http request",
			"method", ctx.Request.Method,
			"route", route,
			"status", status,
			"duration", elapsed,
		)
	}
}
