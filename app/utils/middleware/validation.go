package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Tag     string `json:"tag"`
	Value   string `json:"value,omitempty"`
}

// ValidationErrorResponse represents the full error response
type ValidationErrorResponse struct {
	Error   string            `json:"error"`
	Details []ValidationError `json:"details"`
}

// Global validator instance (reusable)
var validate *validator.Validate

func init() {
	validate = validator.New()
}

func ValidateBody[T any]() gin.HandlerFunc {
	return func(c *gin.Context) {
		var body T

		// Bind JSON to struct
		if err := c.ShouldBindJSON(&body); err != nil {
			recordValidationError(c, http.StatusBadRequest, "Invalid JSON format", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":   "Invalid JSON format",
				"details": err.Error(),
			})
			return
		}

		// Validate struct with validate tags
		if err := validate.Struct(body); err != nil {
			validationErrors := err.(validator.ValidationErrors)
			errorDetails := formatValidationErrors(validationErrors)

			recordValidationError(c, http.StatusBadRequest, "Validation failed", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, ValidationErrorResponse{
				Error:   "Validation failed",
				Details: errorDetails,
			})
			return
		}

		// Store validated body in context for handler to use
		c.Set("validatedBody", body)
		c.Next()
	}
}

func ValidatePathParam(paramName string, validationType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		value := c.Param(paramName)
		// Check if param exists and is not empty
		if value == "" {
			err := fmt.Errorf("missing required path parameter: %s", paramName)
			recordValidationError(c, http.StatusBadRequest, fmt.Sprintf("Missing required path parameter: %s", paramName), err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":   fmt.Sprintf("Missing required path parameter: %s", paramName),
				"details": fmt.Sprintf("Path parameter '%s' is required", paramName),
			})
			return
		}
		var err error
		switch validationType {
		case "uuid":
			err = validate.Var(value, "uuid")
		case "required":
		default:
			err = validate.Var(value, validationType)
		}

		if err != nil {
			recordValidationError(c, http.StatusBadRequest, fmt.Sprintf("Invalid path parameter: %s", paramName), err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":   fmt.Sprintf("Invalid path parameter: %s", paramName),
				"details": fmt.Sprintf("'%s' must be a valid %s", paramName, validationType),
				"value":   value,
			})
			return
		}

		c.Next()
	}
}

func ValidateQueryParams(rules map[string]string) gin.HandlerFunc {
	return func(c *gin.Context) {
		errors := []ValidationError{}

		for paramName, validationType := range rules {
			value := c.Query(paramName)

			// Skip validation if param is optional (empty)
			if value == "" && !strings.Contains(validationType, "required") {
				continue
			}

			// Validate the query param
			if err := validate.Var(value, validationType); err != nil {
				errors = append(errors, ValidationError{
					Field:   paramName,
					Message: fmt.Sprintf("'%s' must be a valid %s", paramName, validationType),
					Tag:     validationType,
					Value:   value,
				})
			}
		}

		if len(errors) > 0 {
			err := fmt.Errorf("query parameter validation failed: %d errors", len(errors))
			recordValidationError(c, http.StatusBadRequest, "Query parameter validation failed", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, ValidationErrorResponse{
				Error:   "Query parameter validation failed",
				Details: errors,
			})
			return
		}

		c.Next()
	}
}

// RequireOneOfQueryParams ensures at least one of the specified query params is present
// Usage: router.GET("/user", middleware.RequireOneOfQueryParams([]string{"email", "username"}), handler.GetUser)
func RequireOneOfQueryParams(params []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		found := false
		for _, param := range params {
			if c.Query(param) != "" {
				found = true
				break
			}
		}

		if !found {
			err := fmt.Errorf("missing required query parameter: at least one of %s is required", strings.Join(params, ", "))
			recordValidationError(c, http.StatusBadRequest, "Missing required query parameter", err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":   "Missing required query parameter",
				"details": fmt.Sprintf("At least one of the following query parameters is required: %s", strings.Join(params, ", ")),
			})
			return
		}

		c.Next()
	}
}

func ValidateHeader(headerName string, validationType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		value := c.GetHeader(headerName)

		if value == "" {
			err := fmt.Errorf("missing required header: %s", headerName)
			recordValidationError(c, http.StatusBadRequest, fmt.Sprintf("Missing required header: %s", headerName), err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":   fmt.Sprintf("Missing required header: %s", headerName),
				"details": fmt.Sprintf("Header '%s' is required", headerName),
			})
			return
		}

		if err := validate.Var(value, validationType); err != nil {
			recordValidationError(c, http.StatusBadRequest, fmt.Sprintf("Invalid header value: %s", headerName), err)
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":   fmt.Sprintf("Invalid header value: %s", headerName),
				"details": fmt.Sprintf("'%s' must be a valid %s", headerName, validationType),
				"value":   value,
			})
			return
		}

		c.Next()
	}
}

func formatValidationErrors(errs validator.ValidationErrors) []ValidationError {
	var errors []ValidationError

	for _, err := range errs {
		errors = append(errors, ValidationError{
			Field:   err.Field(),
			Message: getErrorMessage(err),
			Tag:     err.Tag(),
			Value:   fmt.Sprintf("%v", err.Value()),
		})
	}

	return errors
}

// getErrorMessage returns a human-readable error message based on the validation tag
func getErrorMessage(err validator.FieldError) string {
	field := err.Field()
	tag := err.Tag()
	param := err.Param()

	switch tag {
	case "required":
		return fmt.Sprintf("'%s' is required", field)
	case "email":
		return fmt.Sprintf("'%s' must be a valid email address", field)
	case "min":
		return fmt.Sprintf("'%s' must be at least %s characters", field, param)
	case "max":
		return fmt.Sprintf("'%s' must be at most %s characters", field, param)
	case "gt":
		return fmt.Sprintf("'%s' must be greater than %s", field, param)
	case "gte":
		return fmt.Sprintf("'%s' must be greater than or equal to %s", field, param)
	case "lt":
		return fmt.Sprintf("'%s' must be less than %s", field, param)
	case "lte":
		return fmt.Sprintf("'%s' must be less than or equal to %s", field, param)
	case "uuid":
		return fmt.Sprintf("'%s' must be a valid UUID", field)
	case "alphanum":
		return fmt.Sprintf("'%s' must contain only letters and numbers", field)
	case "datetime":
		return fmt.Sprintf("'%s' must be a valid datetime in format %s", field, param)
	case "oneof":
		return fmt.Sprintf("'%s' must be one of: %s", field, param)
	default:
		return fmt.Sprintf("'%s' failed validation: %s", field, tag)
	}
}

func recordValidationError(c *gin.Context, statusCode int, message string, err error) {
	span := trace.SpanFromContext(c.Request.Context())
	if !span.IsRecording() {
		return
	}

	span.SetAttributes(
		semconv.HTTPStatusCodeKey.Int(statusCode),
		semconv.HTTPMethodKey.String(c.Request.Method),
		semconv.HTTPRouteKey.String(c.FullPath()),
		attribute.String("validation.error", message),
		attribute.String("validation.type", "request_validation"),
	)

	span.RecordError(err)

	if statusCode >= 500 {
		span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d: %s", statusCode, message))
	}
}
