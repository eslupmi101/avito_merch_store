package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/eslupmi101/avito_merch_store/api/controller"
	authTokenMiddleware "github.com/eslupmi101/avito_merch_store/api/middleware"
	"github.com/eslupmi101/avito_merch_store/internal/config"
	"github.com/eslupmi101/avito_merch_store/internal/repository"
	"github.com/eslupmi101/avito_merch_store/internal/usecase"
	"github.com/eslupmi101/avito_merch_store/internal/utility"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestSendCoinSuccess(t *testing.T) {
	Setup()
	defer TearDown()

	// Setup users and balances
	senderID := InsertUser(t, "sender", "password", 500)
	receiverID := InsertUser(t, "receiver", "password", 100)

	cs := repository.NewTransactionRepository(Db)
	csUsecase := usecase.NewCoinSender(cs, 2*time.Second)
	cfg := &config.Config{SecretKey: "testsecret"}
	csController := &controller.CoinSender{
		CoinSenderUsecase: csUsecase,
		Cfg:               cfg,
	}

	router := chi.NewRouter()
	router.Use(authTokenMiddleware.Authorization(cfg.SecretKey))
	router.Post("/api/sendCoin", csController.CoinSender)

	token, err := utility.CreateToken(senderID, cfg.SecretKey)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Prepare request
	reqBody := `{"toUser":"receiver", "amount":100}`
	req, err := http.NewRequest(http.MethodPost, "/api/sendCoin", strings.NewReader(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Assert success
	assert.Equal(t, http.StatusOK, rr.Code, "Expected status 200 OK on successful coin transfer")

	// Check balances
	var senderBalance int
	err = Db.Connection.QueryRow(context.Background(), "SELECT balance FROM users WHERE id = $1", senderID).Scan(&senderBalance)
	if err != nil {
		t.Fatalf("Failed to query sender balance: %v", err)
	}
	assert.Equal(t, 400, senderBalance, "Sender balance should be deducted correctly")

	var receiverBalance int
	err = Db.Connection.QueryRow(context.Background(), "SELECT balance FROM users WHERE id = $1", receiverID).Scan(&receiverBalance)
	if err != nil {
		t.Fatalf("Failed to query receiver balance: %v", err)
	}
	assert.Equal(t, 200, receiverBalance, "Receiver balance should be credited correctly")
}

func TestSendCoinInsufficientBalance(t *testing.T) {
	Setup()
	defer TearDown()

	// Setup users
	senderID := InsertUser(t, "sender", "password", 50)
	receiverID := InsertUser(t, "receiver", "password", 100)

	cs := repository.NewTransactionRepository(Db)
	csUsecase := usecase.NewCoinSender(cs, 2*time.Second)
	cfg := &config.Config{SecretKey: "testsecret"}
	csController := &controller.CoinSender{
		CoinSenderUsecase: csUsecase,
		Cfg:               cfg,
	}

	router := chi.NewRouter()
	router.Use(authTokenMiddleware.Authorization(cfg.SecretKey))
	router.Post("/api/sendCoin", csController.CoinSender)

	token, err := utility.CreateToken(senderID, cfg.SecretKey)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Prepare request with amount greater than balance
	reqBody := `{"toUser":"receiver", "amount":100}`
	req, err := http.NewRequest(http.MethodPost, "/api/sendCoin", strings.NewReader(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Assert failure due to insufficient balance
	assert.Equal(t, http.StatusBadRequest, rr.Code, "Expected status 400 for insufficient funds")

	// Check balances remain unchanged
	var senderBalance int
	err = Db.Connection.QueryRow(context.Background(), "SELECT balance FROM users WHERE id = $1", senderID).Scan(&senderBalance)
	if err != nil {
		t.Fatalf("Failed to query sender balance: %v", err)
	}
	assert.Equal(t, 50, senderBalance, "Sender balance should remain unchanged")

	var receiverBalance int
	err = Db.Connection.QueryRow(context.Background(), "SELECT balance FROM users WHERE id = $1", receiverID).Scan(&receiverBalance)
	if err != nil {
		t.Fatalf("Failed to query receiver balance: %v", err)
	}
	assert.Equal(t, 100, receiverBalance, "Receiver balance should remain unchanged")
}

func TestSendCoinUserNotFound(t *testing.T) {
	Setup()
	defer TearDown()

	// Setup user
	senderID := InsertUser(t, "sender", "password", 500)

	cs := repository.NewTransactionRepository(Db)
	csUsecase := usecase.NewCoinSender(cs, 2*time.Second)
	cfg := &config.Config{SecretKey: "testsecret"}
	csController := &controller.CoinSender{
		CoinSenderUsecase: csUsecase,
		Cfg:               cfg,
	}

	router := chi.NewRouter()
	router.Use(authTokenMiddleware.Authorization(cfg.SecretKey))
	router.Post("/api/sendCoin", csController.CoinSender)

	token, err := utility.CreateToken(senderID, cfg.SecretKey)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Prepare request with non-existent user
	reqBody := `{"toUser":"nonexistentuser", "amount":100}`
	req, err := http.NewRequest(http.MethodPost, "/api/sendCoin", strings.NewReader(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Assert failure due to user not found
	assert.Equal(t, http.StatusBadRequest, rr.Code, "Expected status 400 for non-existent user")
}

func TestSendCoinInvalidToken(t *testing.T) {
	Setup()
	defer TearDown()

	InsertUser(t, "sender", "password", 500)

	cs := repository.NewTransactionRepository(Db)
	csUsecase := usecase.NewCoinSender(cs, 2*time.Second)
	cfg := &config.Config{SecretKey: "testsecret"}
	csController := &controller.CoinSender{
		CoinSenderUsecase: csUsecase,
		Cfg:               cfg,
	}

	router := chi.NewRouter()
	router.Use(authTokenMiddleware.Authorization(cfg.SecretKey))
	router.Post("/api/sendCoin", csController.CoinSender)

	// Prepare request with invalid token
	reqBody := `{"toUser":"receiver", "amount":100}`
	req, err := http.NewRequest(http.MethodPost, "/api/sendCoin", strings.NewReader(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer invalid_token")

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Assert failure due to invalid token
	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Expected status 401 for invalid token")
}

func TestSendCoinMissingToken(t *testing.T) {
	Setup()
	defer TearDown()

	InsertUser(t, "sender", "password", 500)

	cs := repository.NewTransactionRepository(Db)
	csUsecase := usecase.NewCoinSender(cs, 2*time.Second)
	cfg := &config.Config{SecretKey: "testsecret"}
	csController := &controller.CoinSender{
		CoinSenderUsecase: csUsecase,
		Cfg:               cfg,
	}

	router := chi.NewRouter()
	router.Use(authTokenMiddleware.Authorization(cfg.SecretKey))
	router.Post("/api/sendCoin", csController.CoinSender)

	// Prepare request without token
	reqBody := `{"toUser":"receiver", "amount":100}`
	req, err := http.NewRequest(http.MethodPost, "/api/sendCoin", strings.NewReader(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Assert failure due to missing token
	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Expected status 401 for missing token")
}

func TestSendCoinZeroAmount(t *testing.T) {
	Setup()
	defer TearDown()

	// Setup user
	senderID := InsertUser(t, "sender", "password", 500)

	cs := repository.NewTransactionRepository(Db)
	csUsecase := usecase.NewCoinSender(cs, 2*time.Second)
	cfg := &config.Config{SecretKey: "testsecret"}
	csController := &controller.CoinSender{
		CoinSenderUsecase: csUsecase,
		Cfg:               cfg,
	}

	router := chi.NewRouter()
	router.Use(authTokenMiddleware.Authorization(cfg.SecretKey))
	router.Post("/api/sendCoin", csController.CoinSender)

	token, err := utility.CreateToken(senderID, cfg.SecretKey)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	// Prepare request with zero amount
	reqBody := `{"toUser":"receiver", "amount":0}`
	req, err := http.NewRequest(http.MethodPost, "/api/sendCoin", strings.NewReader(reqBody))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	// Assert failure due to zero amount
	assert.Equal(t, http.StatusBadRequest, rr.Code, "Expected status 400 for zero amount")
}
