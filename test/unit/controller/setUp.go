package controller

import (
	"context"
	"log"
	"testing"

	"github.com/eslupmi101/avito_merch_store/internal/config"
	"golang.org/x/crypto/bcrypt"
)

var Db *config.PostgresDb

// Инициализация соединения с базой данных перед каждым тестом
func Setup() {
	connStr := "postgres://test_user:test_password@localhost:5433/test_db"
	Db = config.NewPostgresDb(context.Background(), connStr)
	if err := ClearTables(Db); err != nil {
		log.Fatalf("Не удалось очистить таблицы: %v", err)
	}
}

func ClearTables(db *config.PostgresDb) error {
	tables := []string{"users", "merch", "merch_orders", "transactions"}

	for _, table := range tables {
		_, err := db.Connection.Exec(context.Background(), "TRUNCATE TABLE "+table+" RESTART IDENTITY CASCADE")
		if err != nil {
			return err
		}
	}
	return nil
}

func TearDown() {
	if err := ClearTables(Db); err != nil {
		log.Fatalf("Не удалось очистить таблицы после теста: %v", err)
	}
}

func InsertUser(t *testing.T, username, password string, balance int) int {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}
	var userID int
	err = Db.Connection.QueryRow(context.Background(),
		"INSERT INTO users (username, password, balance) VALUES ($1, $2, $3) RETURNING id",
		username, string(hashedPassword), balance).Scan(&userID)
	if err != nil {
		t.Fatalf("Failed to insert user: %v", err)
	}
	return userID
}
