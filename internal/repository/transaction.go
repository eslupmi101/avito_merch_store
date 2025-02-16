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

type transactionRepositoryImpl struct {
	database *config.PostgresDb
}

func NewTransactionRepository(db *config.PostgresDb) domain.TransactionRepository {
	return &transactionRepositoryImpl{database: db}
}

func (tr transactionRepositoryImpl) SendCoinToUser(ctx context.Context, userID int, toUser string, amount int) error {
	tx, err := tr.database.Connection.BeginTx(
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

	var recipientID int
	err = tx.QueryRow(ctx, "SELECT id FROM users WHERE username = $1", toUser).Scan(&recipientID)
	if err != nil {
		return errors.New("toUser does not exists")
	}

	_, err = tx.Exec(ctx, "UPDATE users SET balance = balance - $1 WHERE id = $2", amount, userID)
	if err != nil {
		if err.Error() == `ERROR: new row for relation "users" violates check constraint "users_balance_check" (SQLSTATE 23514)` {
			slog.Info("insufficient funds", slog.String("SQL error:", err.Error()))
			return errors.New("insufficient funds")
		} else {
			slog.Error(
				"Unexpected error while updating user balance (send coin)",
				slog.String("SQL error:", err.Error()),
				slog.Int("userId", userID),
				slog.Int("amount", amount),
			)
			return fmt.Errorf("failed to update sender balance: %w", err)
		}
	}

	_, err = tx.Exec(ctx, "UPDATE users SET balance = balance + $1 WHERE id = $2", amount, recipientID)
	if err != nil {
		return fmt.Errorf("failed to update recipient balance: %w", err)
	}

	_, err = tx.Exec(ctx, "INSERT INTO transactions (sender, recipient, amount) VALUES ($1, $2, $3)", userID, recipientID, amount)
	if err != nil {
		return fmt.Errorf("failed to insert transaction record: %w", err)
	}

	// Коммитим транзакцию
	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r transactionRepositoryImpl) GetUserTransactions(ctx context.Context, userID int) ([]domain.Transaction, error) {
	var transactions []domain.Transaction

	rows, err := r.database.Connection.Query(
		ctx, `
			SELECT id, sender, recipient, amount
			FROM transactions
			WHERE sender = $1 OR recipient = $1
			ORDER BY id DESC
		`, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve transactions: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var transaction domain.Transaction
		err := rows.Scan(&transaction.ID, &transaction.Sender, &transaction.Recipient, &transaction.Amount)
		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction row: %w", err)
		}
		transactions = append(transactions, transaction)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate over transactions: %w", err)
	}

	return transactions, nil
}
