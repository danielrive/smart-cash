package service

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"smart-cash/expenses-service/internal/models"
	"smart-cash/expenses-service/internal/repositories"
)

// Define service interface

type ExpensesService struct {
	expensesRepository *repositories.DynamoDBExpensesRepository
}

// Create a new expenses service
func NewExpensesService(expensesRepository *repositories.DynamoDBExpensesRepository) *ExpensesService {
	return &ExpensesService{expensesRepository: expensesRepository}
}

func (exps *ExpensesService) CreateExpense(expense models.Expense) (models.ExpensesReturn, error) {

	response, err := exps.expensesRepository.CreateExpense(expense)

	if err != nil {
		log.Println("error", err)
		return models.ExpensesReturn{}, err
	}
	return response, nil
}

// Function to get expenses by Id

func (exps *ExpensesService) GetExpenseById(expenseId string) (models.Expense, error) {
	expense, err := exps.expensesRepository.GetExpenseById(expenseId)
	if err != nil {
		log.Println("error", err)
		return models.Expense{}, err
	}

	return expense, nil

}

// Function to get expenses by userId or category

func (exps *ExpensesService) GetExpByUserIdorCat(key string, value string) ([]models.Expense, error) {
	expenses, err := exps.expensesRepository.GetExpByUserIdorCat(key, value)

	if err != nil {
		log.Println("error", err)
		return expenses, err
	}

	return expenses, nil
}

// Function to process expenses

func (exps *ExpensesService) PayExpenses(expensesId models.ExpensesPay) []models.Expense {
	// get the expense from DB
	expense, err := exps.expensesRepository.GetExpenseById(expensesId.ExpenseId)
	if err != nil {
		log.Printf("Error getting expense from DB %v:", err)
		return []models.Expense{}
	}
	expenses := []models.Expense{expense}
	// send the expenses to payment services sync proccess
	// create payment request per expenses
	baseURL := "http://bank/bank/pay"
	for _, exp := range expenses {
		paymentRequest := models.PaymentRequest{
			ExpenseId: exp.ExpenseId,
			Date:      "11-11-2024", // HARDCODED FOR TESTING
			UserId:    exp.UserId,
			Amount:    exp.Amount,
			Status:    exp.Status,
		}
		jsonData, err := json.Marshal(paymentRequest)
		if err != nil {
			log.Printf("Error marshalling data to JSON %v:", err)
			continue
		}
		// Prepare the request for bank service
		req, err := http.NewRequest(http.MethodPost, baseURL, bytes.NewBuffer(jsonData))
		if err != nil {
			log.Printf("Error creating request %v:", err)
			continue
		}
		// set headers
		req.Header.Set("Content-Type", "application/json")
		// send the request
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Printf("Error sending request %v:", err)
			continue
		}
		// update the state in the expense

		resBody, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("client: could not read response body: %v", err)
			continue
		}
		err = json.Unmarshal(resBody, &paymentRequest)
		if err != nil {
			log.Printf("client: could not parse response body: %v", err)
			continue
		}
		exp.Status = paymentRequest.Status

		exps.expensesRepository.UpdateExpenseStatus(exp)
	}
	return expenses
}
