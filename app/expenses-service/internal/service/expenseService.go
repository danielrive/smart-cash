package service

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"smart-cash/expenses-service/internal/common"
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
func (exps *ExpensesService) PayExpenses(expensesId models.ExpensesPay) (models.Expense, error) {
	// get the expense from DB
	expense, err := exps.expensesRepository.GetExpenseById(expensesId.ExpenseId)
	if err != nil {
		log.Printf("Error getting expense from DB %v:", err)
		return models.Expense{}, common.ErrInternalError
	}
	// send the expenses to payment services sync proccess
	// create payment request per expenses
	baseURL := "http://bank/bank/pay"
	paymentRequest := models.PaymentRequest{
		ExpenseId: expense.ExpenseId,
		Date:      "11-11-2024", // HARDCODED FOR TESTING
		UserId:    expense.UserId,
		Amount:    expense.Amount,
		Status:    expense.Status,
	}
	jsonData, err := json.Marshal(paymentRequest)
	if err != nil {
		log.Printf("Error marshalling data to JSON %v:", err)
		return models.Expense{}, common.ErrInternalError

	}
	// Prepare the request for bank service
	req, err := http.NewRequest(http.MethodPost, baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("Error creating request %v:", err)
		return models.Expense{}, common.ErrInternalError
	}
	// set headers
	req.Header.Set("Content-Type", "application/json")
	// send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error sending request %v:", err)
		return models.Expense{}, common.ErrInternalError
	}

	// validate response code
	if resp.StatusCode != http.StatusCreated {
		log.Printf("expense  %v not paid %v:", expense.ExpenseId, resp.StatusCode)
		return models.Expense{}, common.ErrExpenseNotPaid
	} else {
		resBody, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("client: could not read response body: %v", err) //// HOW to manage this kind of errors when the request already was procesed by another service but
			return models.Expense{}, common.ErrInternalError            // for some situation like server error faild in the service that caalled
		}
		err = json.Unmarshal(resBody, &paymentRequest)
		if err != nil {
			log.Printf("client: could not parse response body: %v", err)
			return models.Expense{}, common.ErrInternalError
		}
		expense.Status = paymentRequest.Status
	}
	// Process response
	exps.expensesRepository.UpdateExpenseStatus(expense)

	return expense, nil
}
