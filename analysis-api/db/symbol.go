package db

import (
	"analysis-api/common"
	"analysis-api/entity"
	"context"
	"gorm.io/gorm"
	"time"
)

type Symbol interface {
	CreateBatch(syms *[]entity.Symbol) error
	Get(pageSize int, pageNumber int, order string, asc bool)
}

type SymbolRepository struct {
	db *gorm.DB
}

func NewSymbolRepository(db *gorm.DB) *SymbolRepository {
	return &SymbolRepository{
		db: db,
	}
}

func (r *SymbolRepository) CreateBatch(syms *[]entity.Symbol) error {
	chunkedSyms := chunkSymbols(*syms, 500)

	r.db.Begin()
	for _, s := range chunkedSyms {
		err := r.db.
			CreateInBatches(&s, 100).
			Error
		if err != nil {
			r.db.Rollback()
			return err
		}
	}
	r.db.Commit()

	return nil
}

func (r *SymbolRepository) Get(pageSize int, pageNumber int, orderBy string, asc bool) ([]entity.Symbol, error) {
	var e []entity.Symbol
	timeoutContext, c := context.WithTimeout(context.Background(), time.Second)
	defer c()

	order := common.FormatOrderQuery(orderBy, asc)
	err := r.db.
		WithContext(timeoutContext).
		Preload("Currency").
		Order(order).
		Offset((pageNumber - 1) * pageSize).
		Limit(pageSize).
		Find(&e).
		Error
	if err != nil {
		return nil, err
	}

	return e, nil
}

func chunkSymbols(slice []entity.Symbol, chunkSize int) [][]entity.Symbol {
	var res [][]entity.Symbol
	for i := 0; i < len(slice); i += chunkSize {
		end := i + chunkSize

		if end > len(slice) {
			end = len(slice)
		}

		res = append(res, slice[i:end])
	}

	return res
}
