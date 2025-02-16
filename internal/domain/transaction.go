package domain

import "context"

type Transaction struct {
	ID        int    `json:"id"`
	Sender    int    `json:"sender"`
	Recipient int    `json:"recipient"`
	Amount    int    `json:"amount"`
}

type TransactionRepository interface {
	SendCoinToUser(ctx context.Context, userID int, ToUser string, amount int) error
	GetUserTransactions(ctx context.Context, userID int) ([]Transaction, error)
}
