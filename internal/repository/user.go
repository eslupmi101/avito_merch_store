package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/eslupmi101/avito_merch_store/internal/config"
	"github.com/eslupmi101/avito_merch_store/internal/domain"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type userRepositoryImpl struct {
	database *config.PostgresDb
}

func NewUserRepository(db *config.PostgresDb) domain.UserRepository {
	return &userRepositoryImpl{database: db}
}

func (r *userRepositoryImpl) GetOrCreateByUsernamePassword(ctx context.Context, username, password string) (*domain.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		return nil, fmt.Errorf("failed to generate hashed password: %w", err)
	}

	insertQuery := `
		INSERT INTO users (username, password, balance)
		VALUES ($1, $2, 10000000)
		ON CONFLICT (username) DO NOTHING
		RETURNING id, username, password, balance
	`
	var user domain.User
	err = r.database.Connection.QueryRow(ctx, insertQuery, username, string(hashedPassword)).
		Scan(&user.ID, &user.Username, &user.HashedPassword, &user.Balance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			selectQuery := `SELECT id, username, password, balance FROM users WHERE username = $1`
			err = r.database.Connection.QueryRow(ctx, selectQuery, username).
				Scan(&user.ID, &user.Username, &user.HashedPassword, &user.Balance)
			if err != nil {
				return nil, fmt.Errorf("failed to select existing user: %w", err)
			}

			if bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password)) != nil {
				return nil, errors.New("invalid username or password")
			}
			return &user, nil
		}
		return nil, fmt.Errorf("failed to insert user: %w", err)
	}

	return &user, nil
}

func (r userRepositoryImpl) GetByID(ctx context.Context, id int) (*domain.User, error) {
	row := r.database.Connection.QueryRow(
		ctx,
		`SELECT id, username, password, balance FROM users WHERE id = $1`,
		id,
	)

	var user domain.User
	var hashedPassword string
	err := row.Scan(&user.ID, &user.Username, &hashedPassword, &user.Balance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &user, nil
}
