package logging

import (
	"log/slog"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
)

// HTTPMiddleware creates HTTP request logging middleware
func HTTPMiddleware(logger *slog.Logger, excludeEndpoints []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		// Skip logging for health checks and metrics
		if slices.Contains(excludeEndpoints, c.Request.URL.Path) {
			return
		}

		duration := time.Since(start)

		attrs := []slog.Attr{
			slog.String("method", c.Request.Method),
			slog.String("path", c.Request.URL.Path),
			slog.Int("status", c.Writer.Status()),
			slog.Duration("duration_ms", duration),
			slog.String("client_ip", c.ClientIP()),
		}

		args := make([]any, len(attrs))
		for i, attr := range attrs {
			args[i] = attr
		}

		logger.Info("request processed", args...)
	}
}
