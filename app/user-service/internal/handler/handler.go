package handler

import (
	"log/slog"
	"net/http"

	"smart-cash/user-service/internal/common"
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
	tr := otel.Tracer("user")
	trContext, childSpan := tr.Start(c.Request.Context(), "handler-get-by-id")
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

	c.JSON(http.StatusOK, user)
}

// Handler for Get user by email or username

func (h *UserHandler) GetUserByQuery(c *gin.Context) {

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
	user, err := h.userService.GetUserByEmailorUsername(key, value)
	if err != nil {
		if err == common.ErrUserNotFound {
			c.JSON(http.StatusNotFound, gin.H{"message": common.ErrUserNotFound.Error()})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalError.Error()})
			return
		}
	}
	c.JSON(http.StatusOK, user)

}

// Handler for creating new user

func (h *UserHandler) CreateUser(c *gin.Context) {
	user := models.User{}
	// bind the JSON data to the user struct
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// create the user
	response, err := h.userService.CreateUser(user)
	if err != nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": err.Error()})
		return
	}
	c.Header("Location", "/user/"+response.UserId)
	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully", "user": response})
}

// login

func (h *UserHandler) Login(c *gin.Context) {
	loginData := models.LoginRequest{}

	if err := c.ShouldBindJSON(&loginData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := h.userService.Login(loginData.Username, loginData.Password)

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
