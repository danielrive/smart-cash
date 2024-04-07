package service

import (
	"bank-service/internal/models"
	"bank-service/internal/repositories"
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
