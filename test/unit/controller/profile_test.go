package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/eslupmi101/avito_merch_store/api/controller"
	authTokenMiddleware "github.com/eslupmi101/avito_merch_store/api/middleware"
	"github.com/eslupmi101/avito_merch_store/internal/config"
	domainAPI "github.com/eslupmi101/avito_merch_store/internal/domain/api"
	"github.com/eslupmi101/avito_merch_store/internal/repository"
	"github.com/eslupmi101/avito_merch_store/internal/usecase"
	"github.com/eslupmi101/avito_merch_store/internal/utility"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

var MERCH_DATA = []struct {
	name  string
	price int
}{
	{"t-shirt", 80},
	{"cup", 20},
	{"book", 50},
	{"pen", 10},
	{"powerbank", 200},
	{"hoody", 300},
	{"umbrella", 200},
	{"socks", 10},
	{"wallet", 50},
	{"pink-hoody", 500},
}

func insertMerchData(t *testing.T) {
	for _, merch := range MERCH_DATA {
		var merchID int
		err := Db.Connection.QueryRow(context.Background(),
			"INSERT INTO merch (name, price) VALUES ($1, $2) RETURNING id",
			merch.name, merch.price).Scan(&merchID)
		if err != nil {
			t.Fatalf("Failed to insert merch %s: %v", merch.name, err)
		}
	}
}

func insertMerchOrder(t *testing.T, owner int, merchName string) {
	var merchID int
	err := Db.Connection.QueryRow(context.Background(),
		"SELECT id FROM merch WHERE name = $1", merchName).Scan(&merchID)
	if err != nil {
		t.Fatalf("Failed to get merch id for %s: %v", merchName, err)
	}
	_, err = Db.Connection.Exec(context.Background(),
		"INSERT INTO merch_orders (owner, merch) VALUES ($1, $2)", owner, merchID)
	if err != nil {
		t.Fatalf("Failed to insert merch order: %v", err)
	}
}

func insertTransaction(t *testing.T, sender, recipient, amount int) {
	_, err := Db.Connection.Exec(context.Background(),
		"INSERT INTO transactions (sender, recipient, amount) VALUES ($1, $2, $3)",
		sender, recipient, amount)
	if err != nil {
		t.Fatalf("Failed to insert transaction: %v", err)
	}
}

func setupProfileController(userID int) (*controller.Profile, *chi.Mux, string) {
	or := repository.NewOrderRepository(Db)
	mr := repository.NewMerchRepository(Db)
	tr := repository.NewTransactionRepository(Db)
	ur := repository.NewUserRepository(Db)
	profileUsecase := usecase.NewProfile(or, mr, tr, ur, 2*time.Second)
	cfg := &config.Config{SecretKey: "testsecret"}
	prfController := &controller.Profile{
		ProfileUsecase: profileUsecase,
		Cfg:            cfg,
	}
	router := chi.NewRouter()
	router.Use(authTokenMiddleware.Authorization(cfg.SecretKey))
	router.Get("/api/info", prfController.Profile)
	token, _ := utility.CreateToken(userID, cfg.SecretKey)
	return prfController, router, token
}

func TestProfileInvalidToken(t *testing.T) {
	Setup()
	defer TearDown()
	cfg := &config.Config{SecretKey: "testsecret"}
	prfController := &controller.Profile{
		ProfileUsecase: nil,
		Cfg:            cfg,
	}
	router := chi.NewRouter()
	router.Use(authTokenMiddleware.Authorization(cfg.SecretKey))
	router.Get("/api/info", prfController.Profile)
	req, _ := http.NewRequest(http.MethodGet, "/api/info", nil)
	req.Header.Set("Authorization", "Bearer invalidtoken")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestProfileNoToken(t *testing.T) {
	Setup()
	defer TearDown()
	cfg := &config.Config{SecretKey: "testsecret"}
	prfController := &controller.Profile{
		ProfileUsecase: nil,
		Cfg:            cfg,
	}
	router := chi.NewRouter()
	router.Use(authTokenMiddleware.Authorization(cfg.SecretKey))
	router.Get("/api/info", prfController.Profile)
	req, _ := http.NewRequest(http.MethodGet, "/api/info", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	assert.Equal(t, http.StatusUnauthorized, rr.Code)
}

func TestProfileCleanUser(t *testing.T) {
	Setup()
	defer TearDown()
	userID := InsertUser(t, "clean_user", "password", 0)
	_, router, token := setupProfileController(userID)
	req, _ := http.NewRequest(http.MethodGet, "/api/info", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	var res domainAPI.ProfileResponse
	err := json.NewDecoder(rr.Body).Decode(&res)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, 0, res.Coins)
	assert.Empty(t, res.Inventory)
	assert.Empty(t, res.CoinHistory.Received)
	assert.Empty(t, res.CoinHistory.Sent)
}

func TestProfileUserWithBalance(t *testing.T) {
	Setup()
	defer TearDown()
	userID := InsertUser(t, "rich_user", "password", 1000)
	_, router, token := setupProfileController(userID)
	req, _ := http.NewRequest(http.MethodGet, "/api/info", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	var res domainAPI.ProfileResponse
	err := json.NewDecoder(rr.Body).Decode(&res)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, 1000, res.Coins)
}

func TestProfileUserWithInventory(t *testing.T) {
	Setup()
	defer TearDown()
	insertMerchData(t)
	userID := InsertUser(t, "inventory_user", "password", 500)
	insertMerchOrder(t, userID, "hoody")
	_, router, token := setupProfileController(userID)
	req, _ := http.NewRequest(http.MethodGet, "/api/info", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	var res domainAPI.ProfileResponse
	err := json.NewDecoder(rr.Body).Decode(&res)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, 500, res.Coins)
	var found bool
	for _, item := range res.Inventory {
		if item.Type == "hoody" && item.Quantity >= 1 {
			found = true
		}
	}
	assert.True(t, found)
}

func TestProfileUserWithSentTransactions(t *testing.T) {
	Setup()
	defer TearDown()
	userID := InsertUser(t, "sender_user", "password", 800)
	recipientID := InsertUser(t, "recipient_user", "password", 200)
	insertTransaction(t, userID, recipientID, 150)
	_, router, token := setupProfileController(userID)
	req, _ := http.NewRequest(http.MethodGet, "/api/info", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	var res domainAPI.ProfileResponse
	err := json.NewDecoder(rr.Body).Decode(&res)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Len(t, res.CoinHistory.Sent, 1)
	assert.Equal(t, 150, res.CoinHistory.Sent[0].Amount)
}

func TestProfileUserWithReceivedTransactions(t *testing.T) {
	Setup()
	defer TearDown()
	senderID := InsertUser(t, "sender_user", "password", 800)
	userID := InsertUser(t, "receiver_user", "password", 300)
	insertTransaction(t, senderID, userID, 200)
	_, router, token := setupProfileController(userID)
	req, _ := http.NewRequest(http.MethodGet, "/api/info", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	var res domainAPI.ProfileResponse
	err := json.NewDecoder(rr.Body).Decode(&res)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Len(t, res.CoinHistory.Received, 1)
	assert.Equal(t, 200, res.CoinHistory.Received[0].Amount)
}

func TestProfileUserFullData(t *testing.T) {
	Setup()
	defer TearDown()
	insertMerchData(t)
	userID := InsertUser(t, "full_user", "password", 1200)
	insertMerchOrder(t, userID, "pen")
	otherUserID := InsertUser(t, "other_user", "password", 500)
	insertTransaction(t, userID, otherUserID, 100)
	insertTransaction(t, otherUserID, userID, 50)
	_, router, token := setupProfileController(userID)
	req, _ := http.NewRequest(http.MethodGet, "/api/info", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)
	var res domainAPI.ProfileResponse
	err := json.NewDecoder(rr.Body).Decode(&res)
	if err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, 1200, res.Coins)
	assert.NotEmpty(t, res.Inventory)
	assert.Len(t, res.CoinHistory.Sent, 1)
	assert.Len(t, res.CoinHistory.Received, 1)
}
