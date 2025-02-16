package utility

import (
	"fmt"
	"time"

	"github.com/eslupmi101/avito_merch_store/internal/domain"
	"github.com/golang-jwt/jwt/v5"
)

func CreateToken(userID int, secretKey string) (string, error) {
	claims := &domain.TokenClaims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}

func parseToken(requestToken string, secret string) (*jwt.Token, *domain.TokenClaims, error) {
	claims := &domain.TokenClaims{}

	token, err := jwt.ParseWithClaims(requestToken, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(secret), nil
	})

	if err != nil {
		return nil, nil, fmt.Errorf("error parsing token: %v", err)
	}

	return token, claims, nil
}

func IsAuthorized(requestToken string, secret string) (bool, error) {
	token, _, err := parseToken(requestToken, secret)
	if err != nil {
		return false, err
	}

	if !token.Valid {
		return false, fmt.Errorf("invalid token")
	}

	return true, nil
}

func ExtractIDFromToken(requestToken string, secret string) (int, error) {
	token, claims, err := parseToken(requestToken, secret)
	if err != nil {
		return 0, err
	}

	if !token.Valid {
		return 0, fmt.Errorf("invalid token")
	}

	return claims.UserID, nil
}
