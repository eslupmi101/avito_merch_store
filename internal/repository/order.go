package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/eslupmi101/avito_merch_store/internal/config"
	"github.com/eslupmi101/avito_merch_store/internal/domain"
	"github.com/jackc/pgx/v5"
)

type orderRepositoryImpl struct {
	database *config.PostgresDb
}

func NewOrderRepository(db *config.PostgresDb) domain.OrderRepository {
	return &orderRepositoryImpl{database: db}
}

func (r orderRepositoryImpl) BuyMerch(ctx context.Context, userID int, merchName string) error {
	tx, err := r.database.Connection.BeginTx(
		ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead},
	)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	var merch domain.Merch
	err = tx.QueryRow(ctx, "SELECT id, name, price FROM merch WHERE name = $1", merchName).Scan(&merch.ID, &merch.Name, &merch.Price)
	if err != nil {
		return errors.New("merch does not exists")
	}

	_, err = tx.Exec(ctx, "UPDATE users SET balance = balance - $1 WHERE id = $2", merch.Price, userID)
	if err != nil {
		if err.Error() == `ERROR: new row for relation "users" violates check constraint "users_balance_check" (SQLSTATE 23514)` {
			slog.Info("insufficient funds", slog.String("SQL error:", err.Error()))
			return errors.New("insufficient funds")
		} else {
			slog.Error(
				"Unexpected error while updating user balance (buy merch)",
				slog.String("SQL error:", err.Error()),
				slog.Int("userId", userID),
				slog.Int("merchPrice", merch.Price),
			)
			return errors.New("failed to update user balance")
		}
	}

	var orderID int
	err = tx.QueryRow(ctx, "INSERT INTO merch_orders (owner, merch) VALUES ($1, $2) RETURNING id", userID, merch.ID).Scan(&orderID)
	if err != nil {
		slog.Error("failed to create merch order", slog.String("SQL error:", err.Error()))
		return fmt.Errorf("failed to create merch order: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		slog.Error("failed to commit transaction", slog.String("SQL error:", err.Error()))
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r orderRepositoryImpl) GetUserMerchAmount(ctx context.Context, userID int) ([]domain.MerchAmount, error) {
	var merchAmounts []domain.MerchAmount

	rows, err := r.database.Connection.Query(
		ctx, `
            SELECT m.name, COUNT(*) AS amount
            FROM merch_orders mo
            JOIN merch m ON mo.merch = m.id
            WHERE mo.owner = $1
            GROUP BY m.name
        `, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve merch amounts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var merchAmount domain.MerchAmount
		err := rows.Scan(&merchAmount.Name, &merchAmount.Amount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan merch amount row: %w", err)
		}
		merchAmounts = append(merchAmounts, merchAmount)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over merch amounts: %w", err)
	}

	return merchAmounts, nil
}
