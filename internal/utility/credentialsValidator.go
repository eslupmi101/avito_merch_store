package utility

import (
	"errors"
	"strings"

	"github.com/go-passwd/validator"
)

func visibleASCIIChars() string {
	var sb strings.Builder
	for i := 33; i <= 126; i++ {
		sb.WriteByte(byte(i))
	}
	return sb.String()
}

func ValidateUsername(username string) error {
	allowedChars := visibleASCIIChars()

	usernameValidator := validator.New(
		validator.MinLength(1, errors.New("username must be at least 1 characters")),
		validator.MaxLength(100, errors.New("username must not exceed 100 characters")),
		validator.ContainsOnly(allowedChars, errors.New("username must contain only visible ASCII characters")),
	)

	return usernameValidator.Validate(username)
}

func ValidatePassword(password string) error {
	allowedChars := visibleASCIIChars()

	passwordValidator := validator.New(
		validator.MinLength(1, errors.New("password must be at least 1 characters")),
		validator.MaxLength(100, errors.New("password must not exceed 100 characters")),
		validator.ContainsOnly(allowedChars, errors.New("password must contain only visible ASCII characters")),
	)

	return passwordValidator.Validate(password)
}

func ValidateMerchName(merchName string) error {
	allowedChars := visibleASCIIChars()

	merchNameValidator := validator.New(
		validator.MinLength(1, errors.New("merchName must be at least 1 characters")),
		validator.MaxLength(100, errors.New("merchName must not exceed 100 characters")),
		validator.ContainsOnly(allowedChars, errors.New("merchName must contain only visible ASCII characters")),
	)

	return merchNameValidator.Validate(merchName)

}
