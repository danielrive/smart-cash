package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

type Claims struct {
	UserID      string `json:"user_id,omitempty"`      // For user tokens
	Username    string `json:"username,omitempty"`     // For user tokens
	Email       string `json:"email,omitempty"`        // For user tokens
	Active      bool   `json:"active,omitempty"`       // For user tokens
	Type        string `json:"type,omitempty"`         // "user" or "service"
	ServiceName string `json:"service_name,omitempty"` // For service tokens
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

		// Handle service tokens vs user tokens
		if claims.Type == "service" {
			// Service token: skip Active check (services are always active)
			if claims.ServiceName == "" {
				err := fmt.Errorf("service token missing service name")
				recordAuthError(c, http.StatusUnauthorized, "invalid service token", err)
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
					"error": "invalid service token: missing service name",
				})
				return
			}
			c.Set("serviceName", claims.ServiceName)
			c.Set("type", "service")
		} else {
			// User token: check if user is active
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
			c.Set("type", "user")
		}

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

// GenerateServiceJWT generates a JWT token for service-to-service authentication
// This allows services to authenticate with other services without requiring user credentials
func GenerateServiceJWT(serviceName string, jwtSecret []byte) (string, error) {
	expirationTime := time.Now().Add(1 * time.Hour)
	claims := &Claims{
		Type:        "service",
		ServiceName: serviceName,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", fmt.Errorf("failed to sign service JWT token: %w", err)
	}

	return tokenString, nil
}
