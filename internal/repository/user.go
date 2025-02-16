package repository

import (
	"context"
	"errors"

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

func (r *userRepositoryImpl) Create(ctx context.Context, username, password string) (*domain.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	query := `INSERT INTO users (username, password, balance) VALUES ($1, $2, 500) RETURNING id, username, balance`
	row := r.database.Connection.QueryRow(ctx, query, username, string(hashedPassword))

	var user domain.User
	err = row.Scan(&user.ID, &user.Username, &user.Balance)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepositoryImpl) GetByUsernamePassword(ctx context.Context, username, password string) (*domain.User, error) {
	row := r.database.Connection.QueryRow(
		ctx,
		`SELECT id, username, password, balance FROM users WHERE username = $1`,
		username,
	)

	var user domain.User
	err := row.Scan(&user.ID, &user.Username, &user.HashedPassword, &user.Balance)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password)); err != nil {
		return nil, errors.New("invalid username or password")
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
