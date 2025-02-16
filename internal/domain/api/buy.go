package domainAPI

import "context"

type BuyUsecase interface {
	BuyMerch(ctx context.Context, userID int, merchName string) error
}
