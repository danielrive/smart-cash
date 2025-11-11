package handler

import (
	"log/slog"
	"net/http"

	"smart-cash/user-service/internal/common"
	"smart-cash/user-service/internal/handler/dto"
	"smart-cash/user-service/internal/service"
	"smart-cash/user-service/models"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
)

type UserHandler struct {
	userService *service.UserService
	logger      *slog.Logger
}

func NewUserHandler(userService *service.UserService, logger *slog.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		logger:      logger,
	}
}

func (h *UserHandler) GetUserById(c *gin.Context) {
	tr := otel.Tracer(common.ServiceName)
	trContext, childSpan := tr.Start(c.Request.Context(), "HandlerGetUserById")
	defer childSpan.End()
	userId := c.Param("userId")
	user, err := h.userService.GetUserById(trContext, userId)
	if err != nil {
		if err == common.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"message": common.ErrUserNotFound})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}
	response := models.UserResponse{
		UserId:   user.UserId,
		Username: user.Username,
		Email:    user.Email,
		Active:   user.Active,
	}

	c.JSON(http.StatusOK, response)
}

// Handler for Get user by email or username

func (h *UserHandler) GetUserByQuery(c *gin.Context) {
	tr := otel.Tracer(common.ServiceName)
	trContext, childSpan := tr.Start(c.Request.Context(), "HandlerGetUserByQuery")
	defer childSpan.End()

	query := c.Request.URL.Query()
	var key, value string
	// Check and store the query in the request
	if email, ok := query["email"]; ok {
		key, value = "email", email[0]
	} else if username, ok := query["username"]; ok {
		key, value = "username", username[0]
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}
	// Get user info by the query
	user, err := h.userService.GetUserByEmailorUsername(trContext, key, value)
	if err != nil {
		if err == common.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"message": common.ErrUserNotFound.Error()})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalError.Error()})
			return
		}
	}

	response := models.UserResponse{
		UserId:   user.UserId,
		Username: user.Username,
		Email:    user.Email,
		Active:   user.Active,
	}

	c.JSON(http.StatusOK, response)

}

// Handler for creating new user

func (h *UserHandler) CreateUser(c *gin.Context) {
	tr := otel.Tracer(common.ServiceName)
	trContext, childSpan := tr.Start(c.Request.Context(), "HandlerCreateUser")
	defer childSpan.End()
	
	// Get validated body from middleware
	validatedBody, exists := c.Get("validatedBody")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed"})
		h.logger.Error("validatedBody not found in context")
		return
	}
	
	// Type assert to our DTO (this is safe because ValidateBody uses generics)
	userDTO, ok := validatedBody.(dto.CreateUserRequest)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		h.logger.Error("failed to cast validatedBody to CreateUserRequest")
		return
	}
	
	// Map DTO to domain model
	user := models.User{
		FirstName: userDTO.FirstName,
		LastName:  userDTO.LastName,
		Username:  userDTO.Username,
		Email:     userDTO.Email,
		Password:  userDTO.Password,
	}
	
	// create the user
	response, err := h.userService.CreateUser(trContext, user)
	if err != nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": err.Error()})
		return
	}
	c.Header("Location", "/user/"+response.UserId)
	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully", "user": response})
}

// login

func (h *UserHandler) Login(c *gin.Context) {
	tr := otel.Tracer(common.ServiceName)
	trContext, childSpan := tr.Start(c.Request.Context(), "HandlerLogin")
	defer childSpan.End()
	
	// Get validated body from middleware
	validatedBody, exists := c.Get("validatedBody")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed"})
		h.logger.Error("validatedBody not found in context")
		return
	}
	
	// Type assert to our DTO
	loginData, ok := validatedBody.(dto.LoginRequest)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		h.logger.Error("failed to cast validatedBody to LoginRequest")
		return
	}

	token, err := h.userService.Login(trContext, loginData.Username, loginData.Password)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"token": token})
}

/// Health check

func (h *UserHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, "ok")
}
