package service

import (
	"smart-cash/bank-service/internal/models"
	"smart-cash/bank-service/internal/repositories"
)

// Define service interface

type BankService struct {
	transactionRepository *repositories.DynamoDBTransactionRepository
}

// Create a new CreateTransaction service
func NewBankService(transactionRepository *repositories.DynamoDBTransactionRepository) *BankService {
	return &BankService{transactionRepository: transactionRepository}
}

func (trans *BankService) CreateTransaction(transaction models.Transaction) error {

	err := trans.transactionRepository.CreateTransaction(transaction)

	if err != nil {
		return err
	}

	return nil
}

// Get transactions by Id
func (trans *BankService) GetTransactions(id string) (models.Transaction, error) {
	transactions, err := trans.transactionRepository.GetTransactionById(id)

	if err != nil {
		return models.Transaction{}, err
	}

	return transactions, nil
}
