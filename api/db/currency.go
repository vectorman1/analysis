package db

import (
	"analysis-api/entity"
	"gorm.io/gorm"
)

type Currency interface {
	GetByCode(code string) (entity.Currency, error)
}

type CurrencyRepository struct {
	db *gorm.DB
}

func NewCurrencyRepository(db *gorm.DB) *CurrencyRepository {
	return &CurrencyRepository{
		db: db,
	}
}

func (r *CurrencyRepository) GetByCode(code string) (entity.Currency, error) {
	var res entity.Currency
	err := r.db.
		First(&res, "code = ?", code).
		Error
	if err != nil {
		return entity.Currency{}, err
	}

	return res, nil
}

func (r *CurrencyRepository) Create(curr *entity.Currency) error {
	err := r.db.
		Create(&curr).
		Error

	return err
}
