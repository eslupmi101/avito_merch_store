package domain

import "context"

type Order struct {
	ID     int    `json:"id"`
	Owner  User   `json:"owner"`
	Merch  Merch  `json:"merch"`
	Status string `json:"status"`
}

type MerchAmount struct {
	Name   string
	Amount int
}

type OrderRepository interface {
	BuyMerch(ctx context.Context, userID int, merchName string) error
	GetUserMerchAmount(ctx context.Context, userID int) ([]MerchAmount, error)
}
