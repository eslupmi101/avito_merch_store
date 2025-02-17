package controller

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/eslupmi101/avito_merch_store/internal/config"
	domainAPI "github.com/eslupmi101/avito_merch_store/internal/domain/api"
	"github.com/eslupmi101/avito_merch_store/internal/utility"
)

type Auth struct {
	AuthUsecase domainAPI.AuthUsecase
	Cfg         *config.Config
}

func (auth *Auth) Authentication(w http.ResponseWriter, r *http.Request) {
	var request domainAPI.AuthRequest

	slog.Info("Received authentication request")

	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		slog.Warn("Failed to decode request body", slog.String("error", err.Error()))
		http.Error(w, utility.JsonError("Invalid body/json"), http.StatusBadRequest)
		return
	}

	if err := request.ValidateUsername(); err != nil {
		slog.Warn("Invalid username format", slog.String("username", request.Username))
		http.Error(w, utility.JsonError("Invalid username"), http.StatusBadRequest)
		return
	}

	if err := request.ValidatePassword(); err != nil {
		slog.Warn("Invalid password format", slog.String("password", request.Username))
		http.Error(w, utility.JsonError("Invalid password"), http.StatusBadRequest)
		return
	}

	slog.Info(
		"Valid format of json username and password ",
		slog.String("username", request.Username),
	)

	user, err := auth.AuthUsecase.GetOrCreateByUsernamePassword(r.Context(), request.Username, request.Password)
	if err != nil {
		slog.Error("User not authorized or lost connection", slog.String("error", err.Error()))
		http.Error(w, utility.JsonError("User not authorized"), http.StatusUnauthorized)
		return
	}

	token, err := auth.AuthUsecase.CreateToken(user.ID, auth.Cfg.SecretKey)
	if err != nil {
		slog.Error("Failed to create token", slog.Int("userID", user.ID), slog.String("error", err.Error()))
		http.Error(w, utility.JsonError("Internal server error"), http.StatusInternalServerError)
		return
	}
	slog.Info("Token generated successfully", slog.Int("userID", user.ID))

	authResponseData := domainAPI.AuthResponse{
		Token: token,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(authResponseData)

	slog.Info("Authentication successful", slog.Int("userID", user.ID))
}
