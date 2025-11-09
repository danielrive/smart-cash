package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"slices"
	"smart-cash/expenses-service/internal/common"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// ValidateRequest validates the request body against the given struct
func ValidateRequest(obj interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := c.ShouldBindJSON(obj); err != nil {
			c.JSON(http.StatusBadRequest, common.NewAppError(
				common.ErrInvalidInput,
				"Invalid request body",
				err.Error(),
			))
			c.Abort()
			return
		}

		if err := validate.Struct(obj); err != nil {
			validationErrors := err.(validator.ValidationErrors)
			c.JSON(http.StatusBadRequest, common.NewAppError(
				common.ErrInvalidInput,
				"Validation failed",
				validationErrors,
			))
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequestLogger logs incoming requests and their responses
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()

		// Process request
		c.Next()

		// Log only if not a health check endpoint
		if !slices.Contains(notToLogEndpoints, c.Request.URL.Path) {
			duration := time.Since(start)
			logger.Info("request processed",
				slog.String("method", c.Request.Method),
				slog.String("path", c.Request.URL.Path),
				slog.Int("status", c.Writer.Status()),
				slog.Duration("duration", duration),
				slog.String("client_ip", c.ClientIP()),
			)
		}
	}
}

// Recovery middleware recovers from any panics and writes a 500 if there was one
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("panic recovered",
					slog.Any("error", err),
					slog.String("stack", string(debug.Stack())),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, common.NewAppError(
					common.ErrInternal,
					"Internal server error",
					nil,
				))
			}
		}()
		c.Next()
	}
}
