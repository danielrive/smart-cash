package service

import (
	"log"
	"smart-cash/bank-service/internal/common"
	"smart-cash/bank-service/internal/models"
	"smart-cash/bank-service/internal/repositories"
)

// Define service interface

type BankService struct {
	bankRepository *repositories.DynamoDBBankRepository
}

// Create a new bank service
func NewBankService(bankRepository *repositories.DynamoDBBankRepository) *BankService {
	return &BankService{bankRepository: bankRepository}
}

func (bank *BankService) ProcessPayment(transaction models.PaymentRequest) (models.PaymentRequest, error) {
	// proccess expenses
	// validate user in bank
	user, err := bank.bankRepository.GetUser(transaction.UserId)
	if err != nil {
		log.Printf("user not registered in the bank: %v", err)
		transaction.Status = "NotPaid"
		return transaction, err
	}
	log.Printf("Processing the expense %v for user %v", transaction.ExpenseId, user.UserId)
	newSaldo, err := processPayment(transaction.Amount, user.Savings)
	if err != nil {
		log.Printf("Transaction failed: %v", err)
		transaction.Status = "NotPaid"
		return transaction, err
	}
	// update saving in user account
	user.Savings = newSaldo
	err = bank.bankRepository.UpdateSavingsUser(user)
	if err != nil {
		log.Printf("Transaction failed: %v", err)
		transaction.Status = "NotPaid"
		return transaction, err
	}
	transaction.Status = "Paid"
	log.Printf("Transaction processed for expense %v", transaction.ExpenseId)

	return transaction, nil

}

// Function to get bank by Id

func (bank *BankService) GetUser(userId string) (models.BankUser, error) {
	user, err := bank.bankRepository.GetUser(userId)
	if err != nil {
		log.Printf("error: %v", err)
		return models.BankUser{}, err
	}

	return user, nil

}

func processPayment(amount, savings float64) (float64, error) {
	newSavings := savings - amount
	// Check if the new savings is negative
	if newSavings < 0 {
		return 0, common.ErrInsufficientFundsError
	}
	return newSavings, nil
}
