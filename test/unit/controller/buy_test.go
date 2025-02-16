package controller

import (
	"context"
	"net/http"
	"net/http/httptest"
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

func insertMerch(t *testing.T, name string, price int) int {
	var merchID int
	err := Db.Connection.QueryRow(context.Background(),
		"INSERT INTO merch (name, price) VALUES ($1, $2) RETURNING id",
		name, price).Scan(&merchID)
	if err != nil {
		t.Fatalf("Failed to insert merch: %v", err)
	}
	return merchID
}

func TestBuySuccess(t *testing.T) {
	Setup()
	defer TearDown()

	userID := InsertUser(t, "buyer", "password", 500)
	merchID := insertMerch(t, "t-shirt", 80)

	or := repository.NewOrderRepository(Db)
	buyUsecase := usecase.NewOrder(or, 2*time.Second)
	cfg := &config.Config{SecretKey: "testsecret"}
	buyController := &controller.Buy{
		BuyUsecase: buyUsecase,
		Cfg:        cfg,
	}

	router := chi.NewRouter()
	router.Use(authTokenMiddleware.Authorization(cfg.SecretKey))
	router.Get("/api/buy/{merchName}", buyController.Buy)

	token, err := utility.CreateToken(userID, cfg.SecretKey)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	req, err := http.NewRequest(http.MethodGet, "/api/buy/t-shirt", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Expected status 200 OK on successful purchase")

	var balance int
	err = Db.Connection.QueryRow(context.Background(), "SELECT balance FROM users WHERE id = $1", userID).Scan(&balance)
	if err != nil {
		t.Fatalf("Failed to query user balance: %v", err)
	}
	assert.Equal(t, 420, balance, "User balance should be deducted correctly")

	var count int
	err = Db.Connection.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM merch_orders WHERE owner = $1 AND merch = $2", userID, merchID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query merch_orders: %v", err)
	}
	assert.Equal(t, 1, count, "There should be one merch order record")
}

func TestBuyFailed(t *testing.T) {
	Setup()
	defer TearDown()

	userID := InsertUser(t, "buyer", "password", 0)
	_ = insertMerch(t, "cup", 20)

	or := repository.NewOrderRepository(Db)
	buyUsecase := usecase.NewOrder(or, 2*time.Second)
	cfg := &config.Config{SecretKey: "testsecret"}
	buyController := &controller.Buy{
		BuyUsecase: buyUsecase,
		Cfg:        cfg,
	}

	router := chi.NewRouter()
	router.Use(authTokenMiddleware.Authorization(cfg.SecretKey))
	router.Get("/api/buy/{merchName}", buyController.Buy)

	token, err := utility.CreateToken(userID, cfg.SecretKey)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	req, err := http.NewRequest(http.MethodGet, "/api/buy/cup", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code, "Expected status 400 for insufficient funds")

	var balance int
	err = Db.Connection.QueryRow(context.Background(), "SELECT balance FROM users WHERE id = $1", userID).Scan(&balance)
	if err != nil {
		t.Fatalf("Failed to query user balance: %v", err)
	}
	assert.Equal(t, 0, balance, "User balance should remain unchanged")

	var count int
	err = Db.Connection.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM merch_orders WHERE owner = $1", userID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query merch_orders: %v", err)
	}
	assert.Equal(t, 0, count, "No merch order should be created")
}

func TestBuyFakeMerchFailed(t *testing.T) {
	Setup()
	defer TearDown()

	userID := InsertUser(t, "buyer", "password", 500)
	fakeMerchName := "asdads123"

	or := repository.NewOrderRepository(Db)
	buyUsecase := usecase.NewOrder(or, 2*time.Second)
	cfg := &config.Config{SecretKey: "testsecret"}
	buyController := &controller.Buy{
		BuyUsecase: buyUsecase,
		Cfg:        cfg,
	}

	router := chi.NewRouter()
	router.Use(authTokenMiddleware.Authorization(cfg.SecretKey))
	router.Get("/api/buy/{merchName}", buyController.Buy)

	token, err := utility.CreateToken(userID, cfg.SecretKey)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}

	req, err := http.NewRequest(http.MethodGet, "/api/buy/"+fakeMerchName, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusBadRequest, rr.Code, "Expected status 400 for non-existent merch")

	var balance int
	err = Db.Connection.QueryRow(context.Background(), "SELECT balance FROM users WHERE id = $1", userID).Scan(&balance)
	if err != nil {
		t.Fatalf("Failed to query user balance: %v", err)
	}
	assert.Equal(t, 500, balance, "User balance should remain unchanged")

	var count int
	err = Db.Connection.QueryRow(context.Background(),
		"SELECT COUNT(*) FROM merch_orders WHERE owner = $1", userID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query merch_orders: %v", err)
	}
	assert.Equal(t, 0, count, "No merch order should be created")
}

func TestBuyBadTokenFailed(t *testing.T) {
	Setup()
	defer TearDown()

	userID := InsertUser(t, "buyer", "password", 500)
	_ = insertMerch(t, "book", 50)

	or := repository.NewOrderRepository(Db)
	buyUsecase := usecase.NewOrder(or, 2*time.Second)
	cfg := &config.Config{SecretKey: "testsecret"}
	buyController := &controller.Buy{
		BuyUsecase: buyUsecase,
		Cfg:        cfg,
	}

	router := chi.NewRouter()
	router.Use(authTokenMiddleware.Authorization(cfg.SecretKey))
	router.Get("/api/buy/{merchName}", buyController.Buy)

	token, err := utility.CreateToken(userID, cfg.SecretKey)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	badToken := token + "invalid"

	req, err := http.NewRequest(http.MethodGet, "/api/buy/book", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+badToken)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Expected status 401 for invalid token")

	var balance int
	err = Db.Connection.QueryRow(context.Background(), "SELECT balance FROM users WHERE id = $1", userID).Scan(&balance)
	if err != nil {
		t.Fatalf("Failed to query user balance: %v", err)
	}
	assert.Equal(t, 500, balance, "User balance should remain unchanged")

	var count int
	err = Db.Connection.QueryRow(context.Background(), "SELECT COUNT(*) FROM merch_orders WHERE owner = $1", userID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query merch_orders: %v", err)
	}
	assert.Equal(t, 0, count, "No merch order should be created")
}

func TestBuyWithoutTokenFailed(t *testing.T) {
	Setup()
	defer TearDown()

	userID := InsertUser(t, "buyer", "password", 500)
	_ = insertMerch(t, "pen", 10)

	or := repository.NewOrderRepository(Db)
	buyUsecase := usecase.NewOrder(or, 2*time.Second)
	cfg := &config.Config{SecretKey: "testsecret"}
	buyController := &controller.Buy{
		BuyUsecase: buyUsecase,
		Cfg:        cfg,
	}

	router := chi.NewRouter()
	router.Use(authTokenMiddleware.Authorization(cfg.SecretKey))
	router.Get("/api/buy/{merchName}", buyController.Buy)

	req, err := http.NewRequest(http.MethodGet, "/api/buy/pen", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Expected status 401 for missing token")

	var balance int
	err = Db.Connection.QueryRow(context.Background(), "SELECT balance FROM users WHERE id = $1", userID).Scan(&balance)
	if err != nil {
		t.Fatalf("Failed to query user balance: %v", err)
	}
	assert.Equal(t, 500, balance, "User balance should remain unchanged")

	var count int
	err = Db.Connection.QueryRow(context.Background(), "SELECT COUNT(*) FROM merch_orders WHERE owner = $1", userID).Scan(&count)
	if err != nil {
		t.Fatalf("Failed to query merch_orders: %v", err)
	}
	assert.Equal(t, 0, count, "No merch order should be created")
}
