package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/eslupmi101/avito_merch_store/api/controller"
	"github.com/eslupmi101/avito_merch_store/internal/config"
	domainAPI "github.com/eslupmi101/avito_merch_store/internal/domain/api"
	"github.com/eslupmi101/avito_merch_store/internal/repository"
	"github.com/eslupmi101/avito_merch_store/internal/usecase"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// ------------------ Тесты Auth ------------------

func TestAuthSuccess(t *testing.T) {
	Setup()
	defer TearDown()

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("testpassword"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Не удалось захешировать пароль: %v", err)
	}

	_, err = Db.Connection.Exec(context.Background(), `INSERT INTO users (username, password, balance) VALUES ('testuser', $1, 100)`, hashedPassword)
	if err != nil {
		t.Fatalf("Не удалось создать пользователя: %v", err)
	}

	userRepo := repository.NewUserRepository(Db)
	authUsecase := usecase.NewAuth(userRepo, 2*time.Second)
	cfg := &config.Config{SecretKey: "testsecret"}
	authController := &controller.Auth{
		AuthUsecase: authUsecase,
		Cfg:         cfg,
	}

	authRequest := domainAPI.AuthRequest{
		Username: "testuser",
		Password: "testpassword",
	}
	requestBody, _ := json.Marshal(authRequest)
	req, err := http.NewRequest(http.MethodPost, "/api/auth", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatalf("Не удалось создать HTTP-запрос: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Создаем ResponseRecorder для получения ответа
	rr := httptest.NewRecorder()

	// Создаем роутер и регистрируем обработчик
	router := chi.NewRouter()
	router.Post("/api/auth", authController.Authentication)

	router.ServeHTTP(rr, req)

	t.Log(rr.Code)
	t.Log(rr.Body.String())

	assert.Equal(t, http.StatusOK, rr.Code, "Ожидался статус-код 200 OK")

	// Проверяем тело ответа
	var authResponse domainAPI.AuthResponse
	err = json.Unmarshal(rr.Body.Bytes(), &authResponse)
	if err != nil {
		t.Fatalf("Не удалось распарсить тело ответа: %v", err)
	}
	assert.NotEmpty(t, authResponse.Token, "Токен не должен быть пустым")
}

func TestAuthWithRegistrationSuccess(t *testing.T) {
	slog.SetDefault(
		slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelError}),
		),
	)

	Setup()
	defer TearDown()

	// Создаем необходимые зависимости
	userRepo := repository.NewUserRepository(Db)
	authUsecase := usecase.NewAuth(userRepo, 2*time.Second)
	cfg := &config.Config{SecretKey: "testsecret"}
	authController := &controller.Auth{
		AuthUsecase: authUsecase,
		Cfg:         cfg,
	}

	// Создаем HTTP-запрос для несуществующего пользователя
	authRequest := domainAPI.AuthRequest{
		Username: "newuser",
		Password: "newpassword",
	}
	requestBody, _ := json.Marshal(authRequest)
	req, err := http.NewRequest(http.MethodPost, "/api/auth", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatalf("Не удалось создать HTTP-запрос: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()

	router := chi.NewRouter()
	router.Post("/api/auth", authController.Authentication)

	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Ожидался статус-код 200 OK")

	var authResponse domainAPI.AuthResponse
	err = json.Unmarshal(rr.Body.Bytes(), &authResponse)
	if err != nil {
		t.Fatalf("Не удалось распарсить тело ответа: %v", err)
	}
	assert.NotEmpty(t, authResponse.Token, "Токен не должен быть пустым")

	var userCount int
	err = Db.Connection.QueryRow(context.Background(), "SELECT COUNT(*) FROM users WHERE username = 'newuser'").Scan(&userCount)
	if err != nil {
		t.Fatalf("Ошибка при проверке наличия пользователя: %v", err)
	}
	assert.Equal(t, 1, userCount, "Пользователь не был создан в базе данных")
}

func TestAuthWithAuthFailure(t *testing.T) {
	// Настройка подключения к базе данных
	Setup()
	defer TearDown()

	// Генерация хешированного пароля для корректного пароля "testpassword"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("testpassword"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("Не удалось захешировать пароль: %v", err)
	}

	// Вставляем пользователя с корректным хешем пароля
	_, err = Db.Connection.Exec(context.Background(),
		`INSERT INTO users (username, password, balance) VALUES ('testuser', $1, 100)`, hashedPassword)
	if err != nil {
		t.Fatalf("Не удалось создать пользователя: %v", err)
	}

	// Создаем зависимости для контроллера
	userRepo := repository.NewUserRepository(Db)
	authUsecase := usecase.NewAuth(userRepo, 2*time.Second)
	cfg := &config.Config{SecretKey: "testsecret"}
	authController := &controller.Auth{
		AuthUsecase: authUsecase,
		Cfg:         cfg,
	}

	// Формируем запрос с неверным паролем
	authRequest := domainAPI.AuthRequest{
		Username: "testuser",
		Password: "wrongpassword", // неверный пароль
	}
	requestBody, _ := json.Marshal(authRequest)
	req, err := http.NewRequest(http.MethodPost, "/api/auth", bytes.NewBuffer(requestBody))
	if err != nil {
		t.Fatalf("Не удалось создать HTTP-запрос: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// Создаем ResponseRecorder для получения ответа
	rr := httptest.NewRecorder()

	// Регистрируем обработчик на роутере
	router := chi.NewRouter()
	router.Post("/api/auth", authController.Authentication)

	// Выполняем запрос
	router.ServeHTTP(rr, req)

	t.Log(rr.Code)
	t.Log(rr.Body.String())

	// Ожидаем, что статус ответа 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, rr.Code, "Ожидался статус-код 401 Unauthorized")

	// Проверяем, что в ответе отсутствует токен (или он пустой)
	var authResponse domainAPI.AuthResponse
	err = json.Unmarshal(rr.Body.Bytes(), &authResponse)
	if err != nil {
		t.Fatalf("Не удалось распарсить тело ответа: %v", err)
	}
	assert.Empty(t, authResponse.Token, "Токен должен быть пустым при ошибке аутентификации")
}

func TestAuthInvalidJsonFailure(t *testing.T) {
	// Создаем контроллер и используем репозитории/кейсы
	userRepo := repository.NewUserRepository(Db)
	authUsecase := usecase.NewAuth(userRepo, 2*time.Second)
	cfg := &config.Config{SecretKey: "testsecret"}
	authController := &controller.Auth{
		AuthUsecase: authUsecase,
		Cfg:         cfg,
	}

	invalidJsonTests := []struct {
		name          string
		requestBody   string
		expectedCode  int
		expectedError string
	}{
		{
			name:          "No username",
			requestBody:   `{"password": "validpassword"}`,
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid username",
		},
		{
			name:          "No password",
			requestBody:   `{"username": "validuser"}`,
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid password",
		},
		{
			name:          "No username and password",
			requestBody:   `{}`,
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid username",
		},
		{
			name:          "Empty username and valid password",
			requestBody:   `{"username": "", "password": "validpassword"}`,
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid username",
		},
		{
			name:          "Valid username and empty password",
			requestBody:   `{"username": "validuser", "password": ""}`,
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid password",
		},
		{
			name:          "Empty username and empty password",
			requestBody:   `{"username": "", "password": ""}`,
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid username",
		},
		{
			name:          "Username with invalid characters (Cyrillic)",
			requestBody:   `{"username": "testuserРусский", "password": "validpassword"}`,
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid username",
		},
		{
			name:          "Password with invalid characters (Cyrillic)",
			requestBody:   `{"username": "validuser", "password": "validpasswordРусский"}`,
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid password",
		},
		{
			name:          "Username and password with invalid characters",
			requestBody:   `{"username": "testuserРусский", "password": "validpasswordРусский"}`,
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid username",
		},
		{
			name:          "Empty JSON",
			requestBody:   `{}`,
			expectedCode:  http.StatusBadRequest,
			expectedError: "Invalid username",
		},
	}

	for _, tt := range invalidJsonTests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, "/api/auth", bytes.NewBufferString(tt.requestBody))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			router := chi.NewRouter()
			// Создаем обработчик с контроллером
			router.Post("/api/auth", authController.Authentication)

			// Выполняем запрос
			router.ServeHTTP(rr, req)

			// Проверка HTTP кода
			if rr.Code != tt.expectedCode {
				t.Errorf("Expected status %v, got %v", tt.expectedCode, rr.Code)
			}

			// Проверка сообщения об ошибке
			var response map[string]string
			err = json.Unmarshal(rr.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("Failed to parse response body: %v", err)
			}

			if response["error"] != tt.expectedError {
				t.Errorf("Expected error message %v, got %v", tt.expectedError, response["error"])
			}
		})
	}
}
