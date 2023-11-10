package service

import (
	"expenses-service/internal/models"
	"expenses-service/internal/repositories"
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
