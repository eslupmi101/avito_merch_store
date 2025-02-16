package route

import (
	"time"

	"github.com/eslupmi101/avito_merch_store/internal/config"
	"github.com/go-chi/chi/v5"
)

func Setup(cfg *config.Config, timeout time.Duration, db *config.PostgresDb, r *chi.Mux) {
	r.Group(func(r chi.Router) {
		NewAuth(cfg, timeout, db, r)
		NewBuy(cfg, timeout, db, r)
		NewCoinSender(cfg, timeout, db, r)
		NewInfo(cfg, timeout, db, r)
	})
}
