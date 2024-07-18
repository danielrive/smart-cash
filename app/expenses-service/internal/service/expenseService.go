package service

import (
	"log"
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

/*
// Create order to automatic pay an expense

func createOrder(expense models.Expense) error {
	// create order format input
	baseURL := "http://payment:8383"
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
		log.Println("Error creating request:", err)
		return err
	}

	// Set the Content-Type header to indicate JSON data (optional, depends on API requirement)
	req.Header.Set("Content-Type", "application/json")

	// Send the request and get the response
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending request:", err)
		return err
	}

	// Close the response body after reading
	defer resp.Body.Close()

	// Call the internal function to validate the user token
	log.Println("Scheduled to pay ", resp.Body)

	return nil
}

func (us *ExpensesService) ConnectOtherSVC(svc_name string, port string) error {
	baseURL := "http://" + svc_name + ":" + port + "/health"
	log.Println(baseURL)
	resp, err := http.Get(baseURL)
	log.Println("response from http call ", resp)
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
*/
