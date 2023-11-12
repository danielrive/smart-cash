package service

import (
	"expenses-service/internal/models"
	"expenses-service/internal/repositories"
	"log"
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

// Calculate total expenes per category
func (exps *ExpensesService) CalculateTotalPerCategory(userId string, category string) int32 {
	var total int32 = 0

	expenses, err := exps.expensesRepository.GetExpensesByUserIdAndCategory(userId, category)

	if err != nil {
		log.Print("error", err)
		return -1
	}

	// for loop to calculate total
	for _, expense := range expenses {
		total = total + expense.Amount
	}

	return total

}
