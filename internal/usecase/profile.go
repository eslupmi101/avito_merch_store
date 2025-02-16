package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/eslupmi101/avito_merch_store/internal/domain"
	domainAPI "github.com/eslupmi101/avito_merch_store/internal/domain/api"
)

type profile struct {
	orderRepository       domain.OrderRepository
	transactionRepository domain.TransactionRepository
	userRepository        domain.UserRepository
	contextTimeout        time.Duration
}

func NewProfile(
	orderRepository domain.OrderRepository,
	merchRepository domain.MerchRepository,
	transactionRepository domain.TransactionRepository,
	userRepository domain.UserRepository,
	timeout time.Duration,
) domainAPI.ProfileUsecase {
	return &profile{
		orderRepository:       orderRepository,
		transactionRepository: transactionRepository,
		userRepository:        userRepository,
		contextTimeout:        timeout,
	}
}

func (prf profile) GetProfile(ctx context.Context, userID int) (*domainAPI.ProfileResponse, error) {
	user, err := prf.userRepository.GetByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	transactions, err := prf.transactionRepository.GetUserTransactions(ctx, userID)
	if err != nil {
		return nil, err
	}

	merchAmount, err := prf.orderRepository.GetUserMerchAmount(ctx, userID)
	if err != nil {
		return nil, err
	}

	var received []domainAPI.Transaction
	var sent []domainAPI.Transaction

	for _, t := range transactions {
		if t.Recipient == userID {
			received = append(received, domainAPI.Transaction{
				FromUser: fmt.Sprintf("%d", t.Sender),
				Amount:   t.Amount,
			})
		} else if t.Sender == userID {
			sent = append(sent, domainAPI.Transaction{
				ToUser: fmt.Sprintf("%d", t.Recipient),
				Amount: t.Amount,
			})
		}
	}

	var inventory []domainAPI.InventoryItem
	for _, item := range merchAmount {
		inventory = append(inventory, domainAPI.InventoryItem{
			Type:     item.Name,
			Quantity: item.Amount,
		})
	}

	profileResponse := &domainAPI.ProfileResponse{
		Coins:     user.Balance,
		Inventory: inventory,
		CoinHistory: domainAPI.CoinHistory{
			Received: received,
			Sent:     sent,
		},
	}

	return profileResponse, nil
}
