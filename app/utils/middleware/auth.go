package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

type Claims struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Active   bool   `json:"active"`
	jwt.RegisteredClaims
}

func AuthMiddleware(jwtSecret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			err := fmt.Errorf("missing authorization header")
			recordAuthError(c, http.StatusUnauthorized, "missing authorization header", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing authorization header",
			})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			err := fmt.Errorf("invalid authorization header format")
			recordAuthError(c, http.StatusUnauthorized, "invalid authorization header format, expected 'Bearer <token>'", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid authorization header format, expected 'Bearer <token>'",
			})
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtSecret, nil
		})

		if err != nil {
			recordAuthError(c, http.StatusUnauthorized, "invalid or expired token", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid or expired token",
			})
			return
		}

		if !token.Valid {
			err := fmt.Errorf("token is not valid")
			recordAuthError(c, http.StatusUnauthorized, "invalid token", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid token",
			})
			return
		}

		// Check if user is active
		if !claims.Active {
			err := fmt.Errorf("user account is inactive")
			recordAuthError(c, http.StatusForbidden, "user account is inactive", err)
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "user account is inactive",
			})
			return
		}

		c.Set("userId", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)
		c.Set("active", claims.Active)
		c.Next()
	}
}

// recordAuthError records authentication errors in the OpenTelemetry span
// According to OpenTelemetry best practices:
// - 4xx (client errors) are expected and should NOT be marked as codes.Error
// - 5xx (server errors) are unexpected and SHOULD be marked as codes.Error
func recordAuthError(c *gin.Context, statusCode int, message string, err error) {
	span := trace.SpanFromContext(c.Request.Context())
	if !span.IsRecording() {
		return
	}

	// Add HTTP status code and semantic conventions
	span.SetAttributes(
		semconv.HTTPStatusCodeKey.Int(statusCode),
		semconv.HTTPMethodKey.String(c.Request.Method),
		semconv.HTTPRouteKey.String(c.FullPath()),
		attribute.String("auth.error", message),
		attribute.String("auth.type", "authentication_failure"),
	)

	span.RecordError(err)

	if statusCode >= 500 {
		span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d: %s", statusCode, message))
	}
}

func OptionalAuthMiddleware(jwtSecret []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.Next()
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return jwtSecret, nil
		})
		if err == nil && token.Valid {
			c.Set("userId", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("email", claims.Email)
		}

		c.Next()
	}
}
