package handler

import (
	"net/http"

	"user-service/internal/common"
	"user-service/internal/models"
	"user-service/internal/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

func (h *UserHandler) GetUser(c *gin.Context) {

	uri := c.Request.URL.Query()

	if _, isMapContainsKey := uri["userId"]; isMapContainsKey {
		user, err := h.userService.GetUserById(uri["userId"][0])
		if err != nil {
			if err == common.ErrUserNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": err})
				return
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
				return
			}
		}
		c.JSON(http.StatusOK, user)
	} else if _, isMapContainsKey := uri["email"]; isMapContainsKey {
		user, err := h.userService.GetUserByEmail(uri["email"][0])
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusOK, user)
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
}

// Handler for creating new user

func (h *UserHandler) CreateUser(c *gin.Context) {
	user := models.User{}
	// bind the JSON data to the user struct
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	// create the user
	if err := h.userService.CreateUser(user); err != nil {
		c.JSON(http.StatusNotImplemented, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, "ok")
}

// Handler for login

func (h *UserHandler) Login(c *gin.Context) {
	// extract the email and user from request

	user := models.User{}
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	}
	if _, token, err := h.userService.Login(user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err})
		return
	} else {
		c.JSON(http.StatusOK, gin.H{"token": token})
		return
	}
}

/// Health check

func (h *UserHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, "ok")
}
