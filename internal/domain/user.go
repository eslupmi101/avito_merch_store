package domain

import (
	"context"
)

type User struct {
	ID             int    `json:"id"`
	Username       string `json:"username"`
	HashedPassword string `json:"hashedPassword"`
	Balance        int    `json:"balance"`
}

type UserRepository interface {
	GetOrCreateByUsernamePassword(ctx context.Context, username, password string) (*User, error)
	GetByID(ctx context.Context, id int) (*User, error)
}
