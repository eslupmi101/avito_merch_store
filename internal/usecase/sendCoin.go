package usecase

import (
	"context"
	"time"

	"github.com/eslupmi101/avito_merch_store/internal/domain"
	domainAPI "github.com/eslupmi101/avito_merch_store/internal/domain/api"
)

type coinSender struct {
	transactionRepository domain.TransactionRepository
	contextTimeout        time.Duration
}

func NewCoinSender(transactionRepository domain.TransactionRepository, timeout time.Duration) domainAPI.CoinSenderUsecase {
	return &coinSender{
		transactionRepository: transactionRepository,
		contextTimeout:        timeout,
	}
}

func (cs *coinSender) SendCoinToUser(ctx context.Context, userID int, ToUser string, amount int) error {
	ctx, cancel := context.WithTimeout(ctx, cs.contextTimeout)
	defer cancel()

	return cs.transactionRepository.SendCoinToUser(ctx, userID, ToUser, amount)
}
