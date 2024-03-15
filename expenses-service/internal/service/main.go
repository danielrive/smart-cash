package service

import (
	"expenses-service/internal/models"
	"expenses-service/internal/repositories"
	"log"
)

type totalExpensesPerCategory struct {
	Category string  `json:"category"`
	UserId   string  `json:"userId"`
	Total    float32 `json:"total"`
}

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

<<<<<<< HEAD
// Calculate total expenes per category
func (exps *ExpensesService) CalculateTotalPerCategory(userId string, category string) (totalExpensesPerCategory, error) {
	var total float32 = 0.0

	expenses, err := exps.expensesRepository.GetExpensesByUserIdAndCategory(userId, category)

	if err != nil {
		log.Print("error", err)
		return totalExpensesPerCategory{}, err
	}

	// for loop to calculate total
	for _, expense := range expenses {
		total = total + float32(expense.Amount)
	}

	// create json response for totalExpensesPerCategory
	totalExpensesPerCategory := totalExpensesPerCategory{
		Category: category,
		UserId:   userId,
		Total:    total,
	}

	return totalExpensesPerCategory, nil

}
=======
// Function to get expenses by tag

func (exps *ExpensesService) GetExpensesByTag(tag string, userId string) ([]models.Expense, error) {

	expenses, err := exps.expensesRepository.GetExpensesByTag(tag, userId)

	if err != nil {
		return nil, err
	}

	return expenses, nil
}


// Function to calculate the cost of expenses by tag

func (exps *ExpensesService) CalculateCostByTag(tag string, userId string) (float64, error) {
	
	expenses, err := exps.expensesRepository.GetExpensesByTag(tag, userId)

	if err != nil {
		return 0, err
	}

	var cost float64
	for _, expense := range expenses {
		cost += expense.amount
	}

	return cost, nil
}
>>>>>>> 2826218 (update k8 version to 1.29)
