package controller

import (
	"log/slog"
	"net/http"

	"github.com/eslupmi101/avito_merch_store/internal/config"
	domainAPI "github.com/eslupmi101/avito_merch_store/internal/domain/api"
	"github.com/go-chi/chi/v5"
)

type Buy struct {
	BuyUsecase domainAPI.BuyUsecase
	Cfg        *config.Config
}

func (buy *Buy) Buy(w http.ResponseWriter, r *http.Request) {
	merchName := chi.URLParam(r, "merchName")
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
		slog.Error("User ID not found")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err := buy.BuyUsecase.BuyMerch(ctx, userID, merchName)
	if err != nil {
		switch err.Error() {
		case "insufficient funds":
			slog.Info("Insufficient funds for user.", slog.Int("userID", userID))
			http.Error(w, "Insufficient funds", http.StatusBadRequest)

		case "merch does not exist":
			slog.Info("Merch does not exists.", slog.String("merchName", merchName))
			http.Error(w, "Merch does not exists", http.StatusBadRequest)

		default:
			slog.Error("Failed to buy merch.", slog.Int("userID", userID), slog.String("merchName", merchName), slog.String("error", err.Error()))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	slog.Info("Buying successful", slog.Int("userID", userID), slog.String("merchName", merchName))
}
