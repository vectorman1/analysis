package service

import (
	"github.com/dystopia-systems/alaskalog"
	"github.com/jackc/pgx/pgtype"
	"github.com/vectorman1/analysis/analysis-api/db"
	"github.com/vectorman1/analysis/analysis-api/entity"
	"github.com/vectorman1/analysis/analysis-api/infrastructure/proto"
	"sync"
	"time"
)

type symbolService interface {
	Get(pageSize int, pageNumber int, orderBy string, asc bool) (*[]entity.Symbol, error)
	GetBulk() (*proto.Symbol, error)
	UpdateBulk(in <-chan *proto.Symbol) (error, int64)
}

type SymbolService struct {
	symbolService
	symbolRepository   *db.SymbolRepository
	currencyRepository *db.CurrencyRepository
}

func NewSymbolService(symbolRepository *db.SymbolRepository, currencyRepository *db.CurrencyRepository) *SymbolService {
	return &SymbolService{
		symbolRepository:   symbolRepository,
		currencyRepository: currencyRepository,
	}
}

func (s *SymbolService) Get(pageSize int, pageNumber int, orderBy string, asc bool) (*[]entity.Symbol, error) {
	return s.symbolRepository.GetPaged(pageSize, pageNumber, orderBy, asc)
}

func (s *SymbolService) GetBulk() *proto.Symbols {
	symbolEntities := s.symbolRepository.GetBulkAsSymbolData()

	syms := proto.Symbols{Symbols: []*proto.Symbol{}}
	var wg sync.WaitGroup
	for eSym := range symbolEntities {
		wg.Add(1)
		go func(eSym *entity.Symbol, wg *sync.WaitGroup) {
			defer wg.Done()
			protoSym := proto.Symbol{
				ID:         uint64(eSym.ID),
				CurrencyID: uint64(eSym.CurrencyID),
				Currency: &proto.Currency{
					ID:       uint64(eSym.Currency.ID),
					Code:     eSym.Currency.Code,
					LongName: eSym.Currency.LongName,
				},
				ISIN:                 eSym.ISIN,
				Identifier:           eSym.Identifier,
				Name:                 eSym.Name,
				MinimumOrderQuantity: eSym.MinimumOrderQuantity.Float,
				MarketName:           eSym.MarketName,
				MarketHoursGMT:       eSym.MarketHoursGMT,
				CreatedAt:            eSym.CreatedAt.Time.Unix(),
				UpdatedAt:            eSym.UpdatedAt.Time.Unix(),
			}
			if eSym.DeletedAt.Status != pgtype.Null {
				protoSym.DeletedAt = eSym.DeletedAt.Time.Unix()
			} else {
				protoSym.DeletedAt = 0
			}

			syms.Symbols = append(syms.Symbols, &protoSym)
		}(eSym, &wg)
	}
	wg.Wait()

	return &syms
}

func (s *SymbolService) InsertBulk(in *[]*proto.Symbol) (bool, error) {
	symbols := s.symbolDataToEntity(in)
	ok, err := s.symbolRepository.InsertBulk(symbols)
	return ok, err
}

func (s *SymbolService) symbolDataToEntity(in *[]*proto.Symbol) []*entity.Symbol {
	var result []*entity.Symbol
	var wg sync.WaitGroup

	queue := make(chan *entity.Symbol, 1)

	wg.Add(len(*in))
	for _, sym := range *in {
		go func(protoSym *proto.Symbol) {
			e := entity.Symbol{}
			if protoSym.CurrencyID == 0 {
				curr, err := s.currencyRepository.GetByCode(protoSym.Currency.Code)
				if err != nil {
					c := &entity.Currency{}
					c.Code = protoSym.Currency.Code
					c.LongName = "temp name"

					id, createErr := s.currencyRepository.Create(c)
					if createErr != nil {
						alaskalog.Logger.Warnf("failed creating currency: %v", err)
						return
					}
					c.ID = id
					e.CurrencyID = c.ID
				} else {
					e.CurrencyID = curr.ID
				}
			} else {
				e.CurrencyID = uint(protoSym.CurrencyID)
			}
			if protoSym.ID != 0 {
				e.ID = uint(protoSym.ID)
			}

			e.ISIN = protoSym.ISIN
			e.Identifier = protoSym.Identifier
			e.Name = protoSym.Name
			var moq pgtype.Float4
			_ = moq.Set(protoSym.MinimumOrderQuantity)
			e.MinimumOrderQuantity = moq
			e.MarketName = protoSym.MarketName
			e.MarketHoursGMT = protoSym.MarketHoursGMT

			if protoSym.CreatedAt == 0 {
				e.CreatedAt = pgtype.Timestamptz{Time: time.Now(), Status: pgtype.Present}
			}
			if protoSym.UpdatedAt != 0 {
				e.UpdatedAt = pgtype.Timestamptz{Time: time.Unix(0, protoSym.UpdatedAt), Status: pgtype.Present}
			} else {
				e.UpdatedAt = pgtype.Timestamptz{Time: time.Now(), Status: pgtype.Present}
			}
			if protoSym.DeletedAt != 0 {
				e.DeletedAt = pgtype.Timestamptz{Time: time.Unix(0, protoSym.DeletedAt), Status: pgtype.Present}
			} else {
				e.DeletedAt = pgtype.Timestamptz{Status: pgtype.Null}
			}

			queue <- &e
		}(sym)
	}

	go func() {
		for sym := range queue {
			result = append(result, sym)
			wg.Done()
		}
	}()

	wg.Wait()

	return result
}
