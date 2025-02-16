package usecase

import (
	"context"
	"time"

	"github.com/eslupmi101/avito_merch_store/internal/domain"
	domainAPI "github.com/eslupmi101/avito_merch_store/internal/domain/api"
	"github.com/eslupmi101/avito_merch_store/internal/utility"
)

type auth struct {
	userRepository domain.UserRepository
	contextTimeout time.Duration
}

func NewAuth(userRepository domain.UserRepository, timeout time.Duration) domainAPI.AuthUsecase {
	return &auth{
		userRepository: userRepository,
		contextTimeout: timeout,
	}
}

func (au *auth) GetByUsernamePassword(ctx context.Context, username, password string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, au.contextTimeout)
	defer cancel()
	return au.userRepository.GetByUsernamePassword(ctx, username, password)
}

func (au *auth) CreateUser(ctx context.Context, username, password string) (*domain.User, error) {
	ctx, cancel := context.WithTimeout(ctx, au.contextTimeout)
	defer cancel()
	return au.userRepository.Create(ctx, username, password)
}

func (auth *auth) CreateToken(userID int, secretKey string) (string, error) {
	return utility.CreateToken(userID, secretKey)
}
