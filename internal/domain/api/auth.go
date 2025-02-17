package domainAPI

import (
	"context"
	"errors"

	"github.com/eslupmi101/avito_merch_store/internal/domain"
	"github.com/eslupmi101/avito_merch_store/internal/utility"
)

type AuthRequest struct {
	Username string `json:"username" form:"username" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
}

type AuthResponse struct {
	Token string `json:"token"`
}

type AuthUsecase interface {
	GetOrCreateByUsernamePassword(ctx context.Context, username, password string) (*domain.User, error)
	CreateToken(userID int, secretKey string) (string, error)
}

func (ar *AuthRequest) ValidateUsername() error {
	if err := utility.ValidateUsername(ar.Username); err != nil {
		return errors.New("invalid username format")
	}
	return nil
}

func (ar *AuthRequest) ValidatePassword() error {
	if err := utility.ValidatePassword(ar.Password); err != nil {
		return errors.New("invalid password format")
	}
	return nil
}
