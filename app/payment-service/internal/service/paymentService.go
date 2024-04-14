package service

import (
	"fmt"
	"log"
	"net/http"
	"smart-cash/payment-service/internal/models"
	"smart-cash/payment-service/internal/repositories"
)

// Define service interface

var domain_name string = "rootdr.info"

type PaymentService struct {
	paymentRepository *repositories.DynamoDBPaymentRepository
}

// Create a new payment service
func NewPaymentService(paymentRepository *repositories.DynamoDBPaymentRepository) *PaymentService {
	return &PaymentService{paymentRepository: paymentRepository}
}

// create order
func (pay *PaymentService) CreateOrder(order models.Order) error {
	data := models.Order{
		OrderId:    "12",
		Date:       "20-12-2024",
		Paid:       false,
		ExpensesId: order.ExpensesId,
		UserId:     order.UserId,
		Amount:     order.Amount,
		Currency:   order.Currency,
	}
	err := pay.paymentRepository.CreateOrder(data)

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

func (us *PaymentService) ConnectOtherSVC(svc_name string) error {
	baseURL := "http://" + svc_name + "." + domain_name + "/health"
	//baseURL := "http://payment:8383"

	resp, err := http.Get(baseURL)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return err
	}

	// Close the response body after reading
	defer resp.Body.Close()

	// Call the internal function to validate the user token
	log.Println("response from http call ", resp)
	return nil

}
