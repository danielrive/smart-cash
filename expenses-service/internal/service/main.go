package service

import (
	"expenses-service/internal/common"
	"expenses-service/internal/models"
	"expenses-service/internal/repositories"
	"fmt"
	"net/http"
	"net/url"
)

// Define service interface

type ExpensesService struct {
	expensesRepository *repositories.DynamoDBExpensesRepository
}

// Create a new expenses service
func NewExpensesService(expensesRepository *repositories.DynamoDBExpensesRepository) *ExpensesService {
	return &ExpensesService{expensesRepository: expensesRepository}
}

func (exps *ExpensesService) CreateExpense(expense models.Expense) error {
	// search user by email

	_, err := exps.expensesRepository.CreateExpense(expense)

	if err != nil {
		return err
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
	baseURL := "http://user-service:8181"

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
