package db

import (
	"analysis-api/entity"
	"gorm.io/gorm"
)

type Symbol interface {
	CreateBatch(syms *[]entity.Symbol) error
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
