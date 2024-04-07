package service

import (
	"payment-service/internal/models"
	"payment-service/internal/repositories"
)

// Define service interface

type PaymentService struct {
	paymentRepository *repositories.DynamoDBPaymentRepository
}

// Create a new payment service
func NewPaymentService(paymentRepository *repositories.DynamoDBPaymentRepository) *PaymentService {
	return &PaymentService{paymentRepository: paymentRepository}
}

// create order
func (pay *PaymentService) CreateOrder(order models.Order) error {

	err := pay.paymentRepository.CreateOrder(order)

	if err != nil {
		return err
	}

	return nil
}

// Get order
func (exps *PaymentService) GetOrderById(id string) (models.Order, error) {
	order, err := exps.paymentRepository.GetOrderById(id)

	if err != nil {
		return order, err
	}

	return order, nil
}
