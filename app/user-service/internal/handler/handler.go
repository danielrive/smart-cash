package handler

import (
	"log/slog"
	"net/http"

	"smart-cash/user-service/internal/common"
	"smart-cash/user-service/internal/handler/dto"
	"smart-cash/user-service/internal/service"
	"smart-cash/user-service/models"
	"smart-cash/utils"

	"github.com/gin-gonic/gin"
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
	ctx, endSpan := utils.StartSpanWithComponent(c.Request.Context(), common.ServiceName, "HandlerGetUserById", "handler")
	defer endSpan()

	userId := c.Param("userId")

	h.logger.Info("getting user by id",
		slog.String("user_id", userId),
		slog.String("component", "handler"),
	)

	user, err := h.userService.GetUserById(ctx, userId)
	if err != nil {
		if err == common.ErrUserNotFound {
			h.logger.Warn("user not found",
				slog.String("user_id", userId),
				slog.String("component", "handler"),
			)
			c.JSON(http.StatusNotFound, gin.H{"error": common.ErrUserNotFound.Error()})
			return
		} else {
			h.logger.Error("error getting user",
				slog.String("user_id", userId),
				slog.String("error", err.Error()),
				slog.String("component", "handler"),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalError.Error()})
			return
		}
	}

	h.logger.Info("user retrieved successfully",
		slog.String("user_id", userId),
		slog.String("username", user.Username),
		slog.String("component", "handler"),
	)

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
	ctx, endSpan := utils.StartSpanWithComponent(c.Request.Context(), common.ServiceName, "HandlerGetUserByQuery", "handler")
	defer endSpan()

	query := c.Request.URL.Query()
	var key, value string
	// Check and store the query in the request
	if email, ok := query["email"]; ok {
		key, value = "email", email[0]
	} else if username, ok := query["username"]; ok {
		key, value = "username", username[0]
	} else {
		h.logger.Warn("invalid query parameters",
			slog.String("component", "handler"),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}

	h.logger.Info("getting user by query",
		slog.String("key", key),
		slog.String("value", value),
		slog.String("component", "handler"),
	)

	// Get user info by the query
	user, err := h.userService.GetUserByEmailorUsername(ctx, key, value)
	if err != nil {
		if err == common.ErrUserNotFound {
			h.logger.Warn("user not found by query",
				slog.String("key", key),
				slog.String("value", value),
				slog.String("component", "handler"),
			)
			c.JSON(http.StatusNotFound, gin.H{"error": common.ErrUserNotFound.Error()})
			return
		} else {
			h.logger.Error("error getting user by query",
				slog.String("key", key),
				slog.String("value", value),
				slog.String("error", err.Error()),
				slog.String("component", "handler"),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalError.Error()})
			return
		}
	}

	h.logger.Info("user retrieved by query successfully",
		slog.String("user_id", user.UserId),
		slog.String("key", key),
		slog.String("component", "handler"),
	)

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
	ctx, endSpan := utils.StartSpanWithComponent(c.Request.Context(), common.ServiceName, "HandlerCreateUser", "handler")
	defer endSpan()

	// Get validated body from middleware
	validatedBody, exists := c.Get("validatedBody")
	if !exists {
		h.logger.Error("validatedBody not found in context",
			slog.String("component", "handler"),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed"})
		return
	}

	// Type assert to our DTO (this is safe because ValidateBody uses generics)
	userDTO, ok := validatedBody.(dto.CreateUserRequest)
	if !ok {
		h.logger.Error("failed to cast validatedBody to CreateUserRequest",
			slog.String("component", "handler"),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	h.logger.Info("creating new user",
		slog.String("username", userDTO.Username),
		slog.String("email", userDTO.Email),
		slog.String("component", "handler"),
	)

	// Map DTO to domain model
	user := models.User{
		FirstName: userDTO.FirstName,
		LastName:  userDTO.LastName,
		Username:  userDTO.Username,
		Email:     userDTO.Email,
		Password:  userDTO.Password,
	}

	// create the user
	response, err := h.userService.CreateUser(ctx, user)
	if err != nil {
		if err == common.ErrUserAlreadyExists {
			h.logger.Warn("user creation failed - already exists",
				slog.String("username", userDTO.Username),
				slog.String("email", userDTO.Email),
				slog.String("component", "handler"),
			)
			c.JSON(http.StatusConflict, gin.H{"error": common.ErrUserAlreadyExists.Error()})
		} else {
			h.logger.Error("error creating user",
				slog.String("username", userDTO.Username),
				slog.String("email", userDTO.Email),
				slog.String("error", err.Error()),
				slog.String("component", "handler"),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalError.Error()})
		}
		return
	}

	h.logger.Info("user created successfully",
		slog.String("user_id", response.UserId),
		slog.String("username", response.Username),
		slog.String("component", "handler"),
	)

	c.Header("Location", "/user/"+response.UserId)
	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully", "user": response})
}

// login

func (h *UserHandler) Login(c *gin.Context) {
	ctx, endSpan := utils.StartSpanWithComponent(c.Request.Context(), common.ServiceName, "HandlerLogin", "handler")
	defer endSpan()

	// Get validated body from middleware
	validatedBody, exists := c.Get("validatedBody")
	if !exists {
		h.logger.Error("validatedBody not found in context",
			slog.String("component", "handler"),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed"})
		return
	}

	// Type assert to our DTO
	loginData, ok := validatedBody.(dto.LoginRequest)
	if !ok {
		h.logger.Error("failed to cast validatedBody to LoginRequest",
			slog.String("component", "handler"),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	h.logger.Info("login attempt",
		slog.String("username", loginData.Username),
		slog.String("component", "handler"),
	)

	token, err := h.userService.Login(ctx, loginData.Username, loginData.Password)

	if err != nil {
		if err == common.ErrWrongCredentials {
			h.logger.Warn("login failed - wrong credentials",
				slog.String("username", loginData.Username),
				slog.String("component", "handler"),
			)
			c.JSON(http.StatusUnauthorized, gin.H{"error": common.ErrWrongCredentials.Error()})
		} else if err == common.ErrUserInactive {
			h.logger.Warn("login failed - user account is inactive",
				slog.String("username", loginData.Username),
				slog.String("component", "handler"),
			)
			c.JSON(http.StatusForbidden, gin.H{"error": common.ErrUserInactive.Error()})
		} else {
			h.logger.Error("login failed - internal error",
				slog.String("username", loginData.Username),
				slog.String("error", err.Error()),
				slog.String("component", "handler"),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalError.Error()})
		}
		return
	}

	h.logger.Info("login successful",
		slog.String("username", loginData.Username),
		slog.String("component", "handler"),
	)

	c.JSON(http.StatusOK, gin.H{"token": token})
}

/// Health check

func (h *UserHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, "ok")
}
