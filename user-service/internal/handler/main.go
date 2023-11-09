package handler

import (
	"net/http"

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
	userID := c.Param("id")
	user, err := h.userService.GetUser(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
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
	if err := h.userService.CreateUser(user); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not created"})
		return
	}
	c.JSON(http.StatusOK, "ok")
}

func (h *UserHandler) FindUser(c *gin.Context) {
	//user := models.User{}
	uri := c.Request.URL.Query()

	if _, isMapContainsKey := uri["email"]; isMapContainsKey {
		user, err := h.userService.FindUserByEmail(uri["email"][0])
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusOK, user)
	} else if _, isMapContainsKey := uri["username"]; isMapContainsKey {
		user, err := h.userService.FindUserByUsername(uri["username"][0])
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
