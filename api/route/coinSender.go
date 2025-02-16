package route

import (
	"time"

	"github.com/eslupmi101/avito_merch_store/api/controller"
	"github.com/eslupmi101/avito_merch_store/internal/config"
	"github.com/eslupmi101/avito_merch_store/internal/repository"
	"github.com/eslupmi101/avito_merch_store/internal/usecase"
	"github.com/go-chi/chi/v5"
)

func NewCoinSender(cfg *config.Config, timeout time.Duration, db *config.PostgresDb, router chi.Router) {
	tr := repository.NewTransactionRepository(db)
	scc := &controller.CoinSender{
		CoinSenderUsecase: usecase.NewCoinSender(tr, timeout),
		Cfg:               cfg,
	}
	router.Post("/api/sendCoin", scc.CoinSender)
}
