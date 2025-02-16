package controller

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/eslupmi101/avito_merch_store/internal/config"
	domainAPI "github.com/eslupmi101/avito_merch_store/internal/domain/api"
)

type Profile struct {
	ProfileUsecase domainAPI.ProfileUsecase
	Cfg            *config.Config
}

func (prf *Profile) Profile(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	values, ok := ctx.Value(config.AuthMiddlewareValuesKey).(map[string]interface{})
	if !ok {
		slog.Error("Cannot retrieve middleware values from context")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	isAuthorizedAny, _ := values["isAuthorized"].(bool)
	if !isAuthorizedAny {
		slog.Error("User not authorized", slog.Bool("bool", values["isAuthorized"].(bool)))
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	userID, ok := values["userID"].(int)
	if !ok {
		slog.Error("UserID not found")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	profileResponseData, err := prf.ProfileUsecase.GetProfile(ctx, userID)
	if err != nil {
		slog.Error("Failed to get profile", slog.Int("userID", userID), slog.String("error", err.Error()))
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(profileResponseData)

	slog.Info("Authentication successful", slog.Int("userID", userID))
}
