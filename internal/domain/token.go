package domain

import "github.com/golang-jwt/jwt/v5"

type TokenClaims struct {
	UserID int `json:"id"`
	jwt.RegisteredClaims
}
