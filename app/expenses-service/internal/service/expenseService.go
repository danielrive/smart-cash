package service

import (
	"bytes"
	"encoding/json"
	"log"
	"smart-cash/expenses-service/internal/common"
	"smart-cash/expenses-service/internal/models"
	"smart-cash/expenses-service/internal/repositories"

	"fmt"
	"net/http"
	"net/url"
)

var domain_name string = "rootdr.info"

// Define service interface

type ExpensesService struct {
	expensesRepository *repositories.DynamoDBExpensesRepository
}

// Create a new expenses service
func NewExpensesService(expensesRepository *repositories.DynamoDBExpensesRepository) *ExpensesService {
	return &ExpensesService{expensesRepository: expensesRepository}
}

func (exps *ExpensesService) CreateExpense(expense models.Expense) error {

	_, err := exps.expensesRepository.CreateExpense(expense)

	if err != nil {
		return err
	}
	// validate if the expense has automatic pay
	if expense.AutomaticPay {
		// Call the internal function to validate the user token
		createOrder(expense)
	}

	return nil
}

// Function to get expenses by tag

func (exps *ExpensesService) GetExpensesByCategory(tag string, userId string) ([]models.Expense, error) {
	if validateUserToken(userId) == 200 {
		expenses, err := exps.expensesRepository.GetExpensesByCategory(tag, userId)

		if err != nil {
			return nil, err
		}

		return expenses, nil
	} else {
		return nil, common.ErrWrongCredentials
	}
}

// Get expenses by userId

func (exps *ExpensesService) GetExpensesByUserId(userId string) ([]models.Expense, error) {

	expenses, err := exps.expensesRepository.GetExpensesByUserId(userId)

	if err != nil {
		return nil, err
	}

	return expenses, nil
}

// Function to calculate the cost of expenses by tag

func (exps *ExpensesService) CalculateCostByTag(tag string, userId string) (float64, error) {

	expenses, err := exps.expensesRepository.GetExpensesByCategory(tag, userId)

	if err != nil {
		return 0, err
	}

	var cost float64
	for _, expense := range expenses {
		cost += expense.Amount
	}

	return cost, nil
}

// Internal function to validate user token
func validateUserToken(userId string) int {
	// Define the base URL of the service
	baseURL := "http://user/login"

	// Create a map to hold query parameters
	queryParams := map[string]string{
		"id": userId,
	}

	// Encode the query parameters and append them to the base URL
	u, err := url.Parse(baseURL)
	if err != nil {
		fmt.Printf("Error parsing URL: %v\n", err)
		//	return err
	}
	q := u.Query()
	for key, value := range queryParams {
		q.Set(key, value)
	}
	u.RawQuery = q.Encode()
	// Make a GET request with the constructed URL
	resp, err := http.Get(u.String())
	if err != nil {
		fmt.Printf("Error making HTTP request: %v\n", err)
		//return err
	}
	defer resp.Body.Close()
	// Check the response status code
	fmt.Println(resp.StatusCode)

	return resp.StatusCode
}

// Create order to automatic pay an expense

func createOrder(expense models.Expense) error {
	// create order format input
	baseURL := "http://payment"
	//baseURL := "http://payment:8383"

	// Create a map to hold query parameters
	data := map[string]interface{}{
		"expensesId": expense.ExpenseId,
		"amount":     expense.Amount,
		"userId":     expense.UserId,
		"currency":   expense.Currency,
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshalling data to JSON:", err)
		return err
	}

	// Create a new HTTP request object
	req, err := http.NewRequest("POST", baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return err
	}

	// Set the Content-Type header to indicate JSON data (optional, depends on API requirement)
	req.Header.Set("Content-Type", "application/json")

	// Send the request and get the response
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return err
	}

	// Close the response body after reading
	defer resp.Body.Close()

	// Call the internal function to validate the user token
	fmt.Println("scheduled to pay ", resp.Body)

	return nil
}

func (us *ExpensesService) ConnectOtherSVC(svc_name string) error {
	baseURL := "http://" + svc_name + "." + domain_name + "/health"

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
