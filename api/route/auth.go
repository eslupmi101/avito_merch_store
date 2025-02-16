package route

import (
	"time"

	"github.com/eslupmi101/avito_merch_store/api/controller"
	"github.com/eslupmi101/avito_merch_store/internal/config"
	"github.com/eslupmi101/avito_merch_store/internal/repository"
	"github.com/eslupmi101/avito_merch_store/internal/usecase"
	"github.com/go-chi/chi/v5"
)

func NewAuth(cfg *config.Config, timeout time.Duration, db *config.PostgresDb, router chi.Router) {
	ur := repository.NewUserRepository(db)
	ac := &controller.Auth{
		AuthUsecase: usecase.NewAuth(ur, timeout),
		Cfg:         cfg,
	}
	router.Post("/api/auth", ac.Authentication)
}
