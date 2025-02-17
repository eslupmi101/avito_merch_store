package repository

import (
	"context"
	"errors"
	"fmt"

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

func (tr transactionRepositoryImpl) SendCoinToUser(ctx context.Context, userID int, toUser string, amount int) (err error) {
	tx, err := tr.database.Connection.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.ReadCommitted})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	result, err := tx.Exec(ctx, `
        UPDATE users
        SET balance = balance - $1
        WHERE id = $2 AND balance >= $1
    `, amount, userID)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return errors.New("insufficient funds")
	}

	var recipientID int
	err = tx.QueryRow(ctx, `
        UPDATE users
        SET balance = balance + $1
        WHERE username = $2
        RETURNING id
    `, amount, toUser).Scan(&recipientID)
	if err != nil {
		return errors.New("toUser does not exist")
	}

	return tx.Commit(ctx)
}

func (r transactionRepositoryImpl) GetUserTransactions(ctx context.Context, userID int) ([]domain.Transaction, error) {
	var transactions []domain.Transaction

	rows, err := r.database.Connection.Query(
		ctx,
		`SELECT id, sender, recipient, amount
		FROM transactions
		WHERE sender = $1 OR recipient = $1
		ORDER BY id DESC`,
		userID,
	)
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
