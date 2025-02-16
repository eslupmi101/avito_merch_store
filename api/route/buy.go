package route

import (
	"time"

	"github.com/eslupmi101/avito_merch_store/api/controller"
	"github.com/eslupmi101/avito_merch_store/internal/config"
	"github.com/eslupmi101/avito_merch_store/internal/repository"
	"github.com/eslupmi101/avito_merch_store/internal/usecase"
	"github.com/go-chi/chi/v5"
)

func NewBuy(cfg *config.Config, timeout time.Duration, db *config.PostgresDb, router chi.Router) {
	or := repository.NewOrderRepository(db)
	bc := &controller.Buy{
		BuyUsecase: usecase.NewOrder(or, timeout),
		Cfg:        cfg,
	}
	router.Get("/api/buy/{merchName}", bc.Buy)
}
