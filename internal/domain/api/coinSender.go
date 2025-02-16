package domainAPI

import (
	"context"
	"errors"
	"log/slog"

	"github.com/eslupmi101/avito_merch_store/internal/utility"
)

type CoinSenderRequest struct {
	ToUser string `json:"toUser" form:"toUser" binding:"required"`
	Amount int    `json:"amount" form:"amount" binding:"required"`
}

type CoinSenderUsecase interface {
	SendCoinToUser(ctx context.Context, userID int, ToUser string, amount int) error
}

func (cs *CoinSenderRequest) ValidateToUser() error {
	if err := utility.ValidateUsername(cs.ToUser); err != nil {
		slog.Info("Invalid ToUser format", slog.String("ToUser", cs.ToUser))
		return errors.New("invalid ToUser format")
	}
	return nil
}

func (cs *CoinSenderRequest) ValidateAmount() error {
	if cs.Amount <= 0 {
		slog.Info("Invalid amount", slog.Int("amount", cs.Amount))
		return errors.New("invalid amount")
	}

	return nil
}
