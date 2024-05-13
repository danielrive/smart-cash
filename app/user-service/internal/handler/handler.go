package handler

import (
	"net/http"

	"smart-cash/user-service/internal/common"
	"smart-cash/user-service/internal/models"
	"smart-cash/user-service/internal/service"

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
				c.JSON(http.StatusNotFound, gin.H{"message": common.ErrUserNotFound})
				return
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
				return
			}
		}
		c.JSON(http.StatusOK, user)
	} else if _, isMapContainsKey := uri["email"]; isMapContainsKey {
		user, err := h.userService.GetUserByEmailorUsername("email", uri["email"][0])
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": common.ErrUserNotFound})
			return
		}
		c.JSON(http.StatusOK, user)
	} else if _, isMapContainsKey := uri["username"]; isMapContainsKey {
		user, err := h.userService.GetUserByEmailorUsername("username", uri["email"][0])
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": common.ErrUserNotFound})
			return
		}
		c.JSON(http.StatusOK, user)
	} else {
		c.JSON(http.StatusNotFound, gin.H{"message": common.ErrUserNotFound})
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
	c.JSON(http.StatusCreated, "ok")
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

// Handler to connect to other svc (just test)

func (h *UserHandler) ConnectToOtherSvc(c *gin.Context) {

	uri := c.Request.URL.Query()

	err := h.userService.ConnectOtherSVC(uri["svcName"][0], uri["port"][0])

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, "ok")

}
