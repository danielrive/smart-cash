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
