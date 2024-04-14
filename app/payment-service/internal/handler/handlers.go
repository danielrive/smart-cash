package handler

import (
	"net/http"

	"smart-cash/payment-service/internal/common"
	"smart-cash/payment-service/internal/models"
	"smart-cash/payment-service/internal/service"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentService *service.PaymentService
}

func NewPaymentHandler(paymentService *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{paymentService: paymentService}
}

// Handler for creating new order

func (h *PaymentHandler) CreateOrder(c *gin.Context) {
	order := models.Order{}
	// bind the JSON data to the payment struct
	if err := c.ShouldBindJSON(&order); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// create the expense
	if err := h.paymentService.CreateOrder(order); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not created"})
		return
	}
	c.JSON(http.StatusOK, "ok")

}

// Handler for Get method

func (h *PaymentHandler) GetOrder(c *gin.Context) {
	// validate the query in the url to see with what attribute filter
	uri := c.Request.URL.Query()

	// if orderId is not present, then we need to return an error
	// if tag is present, then we need to get payment by category, otherwise get all payment by orderId

	if _, isMapContainsKey := uri["orderId"]; !isMapContainsKey {
		c.JSON(http.StatusBadRequest, gin.H{"error": "orderId is required"})
		return
	} else if _, isMapContainsKey := uri["orderId"]; isMapContainsKey {
		order, err := h.paymentService.GetOrderById(uri["orderId"][0])
		if err != nil {
			if err == common.ErrOrderNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": common.ErrOrderNotFound})
				return
			} else if err == common.ErrWrongCredentials {
				c.JSON(http.StatusUnauthorized, gin.H{"error": common.ErrWrongCredentials})
				return
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
				return
			}
		}
		c.JSON(http.StatusOK, order)
		return
	}
}

/// Health check

func (h *PaymentHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, "ok")
}

func (h *PaymentHandler) ConnectToOtherSvc(c *gin.Context) {

	uri := c.Request.URL.Query()

	err := h.paymentService.ConnectOtherSVC(uri["svcName"][0])

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, "ok")

}
