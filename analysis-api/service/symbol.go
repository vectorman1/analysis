package service

import (
	"analysis-api/db"
	"analysis-api/entity"
	"analysis-api/infrastructure"
	"strconv"
)

type symbol interface {
	CreateOrUpdateBatchFromExternal(syms []infrastructure.ExternalSymbol) error
}

type SymbolService struct {
	symbolRepository   *db.SymbolRepository
	currencyRepository *db.CurrencyRepository
}

func NewSymbolService(symbolRepository *db.SymbolRepository, currencyRepository *db.CurrencyRepository) *SymbolService {
	return &SymbolService{
		symbolRepository:   symbolRepository,
		currencyRepository: currencyRepository,
	}
}

func (s *SymbolService) CreateBatchFromExternal(syms []infrastructure.ExternalSymbol) error {
	var symbols []entity.Symbol
	for _, es := range syms {
		curr, err := s.currencyRepository.GetByCode(es.CurrencyCode)
		if err != nil {
			curr = entity.Currency{
				Code: es.CurrencyCode,
				Name: "temp currency name",
			}
			err = s.currencyRepository.Create(&curr)
			if err != nil {
				return err
			}
		}

		minQuantity, _ := strconv.ParseFloat(es.MinTradedQuantity, 32)
		sym := entity.Symbol{
			Name:                 es.Instrument,
			CompanyName:          es.Company,
			CurrencyID:           curr.ID,
			ISIN:                 es.ISIN,
			MinimumOrderQuantity: float32(minQuantity),
			MarketName:           es.MarketName,
			MarketHoursGMT:       es.MarketHoursGMT,
		}
		symbols = append(symbols, sym)
	}

	return s.symbolRepository.CreateBatch(&symbols)
}
