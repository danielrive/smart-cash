package common

import "errors"

var (
	ErrOrchestrationFailed   = errors.New("payment orchestration failed")
	ErrExpenseUpdateFailed   = errors.New("failed to update expense status")
	ErrBankTransactionFailed = errors.New("bank transaction failed")
	ErrPaymentUpdateFailed   = errors.New("failed to update payment status")
	ErrInternalError         = errors.New("internal error")
)
