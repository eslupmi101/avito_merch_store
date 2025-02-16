package usecase

import (
	"context"
	"time"

	"github.com/eslupmi101/avito_merch_store/internal/domain"
	domainAPI "github.com/eslupmi101/avito_merch_store/internal/domain/api"
)

type buy struct {
	orderRepository domain.OrderRepository
	contextTimeout  time.Duration
}

func NewOrder(orderRepository domain.OrderRepository, timeout time.Duration) domainAPI.BuyUsecase {
	return &buy{
		orderRepository: orderRepository,
		contextTimeout:  timeout,
	}
}

func (o *buy) BuyMerch(ctx context.Context, userID int, merchName string) error {
	ctx, cancel := context.WithTimeout(ctx, o.contextTimeout)
	defer cancel()

	return o.orderRepository.BuyMerch(ctx, userID, merchName)
}
