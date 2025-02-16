package repository

import (
	"github.com/eslupmi101/avito_merch_store/internal/config"
	"github.com/eslupmi101/avito_merch_store/internal/domain"
)

type merchRepositoryImpl struct {
	database *config.PostgresDb
}

func NewMerchRepository(db *config.PostgresDb) domain.MerchRepository {
	return &merchRepositoryImpl{database: db}
}
