package controller

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/eslupmi101/avito_merch_store/internal/config"
	domainAPI "github.com/eslupmi101/avito_merch_store/internal/domain/api"
	"github.com/eslupmi101/avito_merch_store/internal/utility"
)

type CoinSender struct {
	CoinSenderUsecase domainAPI.CoinSenderUsecase
	Cfg               *config.Config
}

func (cs *CoinSender) CoinSender(w http.ResponseWriter, r *http.Request) {
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

	var request domainAPI.CoinSenderRequest
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		slog.Warn("Failed to decode request body", slog.String("error", err.Error()))
		http.Error(w, utility.JsonError("Invalid body/json"), http.StatusBadRequest)
		return
	}

	if err := request.ValidateToUser(); err != nil {
		slog.Info("Validation toUser failed", slog.String("error", err.Error()))
		http.Error(w, "Invalid ToUser", http.StatusBadRequest)
		return
	}

	if err := request.ValidateAmount(); err != nil {
		slog.Info("Validation amount failed", slog.String("error", err.Error()))
		http.Error(w, "Invalid amount", http.StatusBadRequest)
		return
	}

	slog.Info(
		"Valid format of json toUser and amount ",
		slog.String("username", request.ToUser),
	)

	if err := cs.CoinSenderUsecase.SendCoinToUser(ctx, userID, request.ToUser, request.Amount); err != nil {
		switch err.Error() {
		case "insufficient funds":
			slog.Info(".", slog.Int("userID", userID))
			http.Error(w, "Insufficient funds", http.StatusBadRequest)

		case "toUser does not exist":
			slog.Info("toUser does not exists.", slog.String("ToUser", request.ToUser))
			http.Error(w, "toUser does not exists", http.StatusBadRequest)

		default:
			slog.Error("Failed to send coin.", slog.Int("userID", userID), slog.String("ToUser", request.ToUser), slog.String("error", err.Error()))
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
	slog.Info("Coin sent successfully", slog.Int("userID", userID), slog.String("ToUser", request.ToUser))
}
