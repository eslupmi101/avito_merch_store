package route

import (
	"time"

	"github.com/eslupmi101/avito_merch_store/api/controller"
	"github.com/eslupmi101/avito_merch_store/internal/config"
	"github.com/eslupmi101/avito_merch_store/internal/repository"
	"github.com/eslupmi101/avito_merch_store/internal/usecase"
	"github.com/go-chi/chi/v5"
)

func NewInfo(cfg *config.Config, timeout time.Duration, db *config.PostgresDb, router chi.Router) {
	or := repository.NewOrderRepository(db)
	mr := repository.NewMerchRepository(db)
	tr := repository.NewTransactionRepository(db)
	ur := repository.NewUserRepository(db)
	pc := &controller.Profile{
		ProfileUsecase: usecase.NewProfile(
			or,
			mr,
			tr,
			ur,
			timeout,
		),
		Cfg: cfg,
	}
	router.Get("/api/info", pc.Profile)
}
