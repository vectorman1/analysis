package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx"

	"github.com/jackc/pgx/pgtype"
	"github.com/vectorman1/analysis/analysis-api/db"
	"github.com/vectorman1/analysis/analysis-api/generated/proto_models"
	"github.com/vectorman1/analysis/analysis-api/generated/symbol_service"
	"github.com/vectorman1/analysis/analysis-api/generated/trading212_service"
	"github.com/vectorman1/analysis/analysis-api/model"
)

type symbolsService interface {
	// repo methods
	GetPaged(ctx context.Context, req *symbol_service.ReadPagedSymbolRequest) (*[]*proto_models.Symbol, error)
	GetByISINAndName(isin, name string) (*proto_models.Symbol, error)
	InsertBulk(timeoutContext context.Context, symbols []*proto_models.Symbol) (bool, error)
	// service methods
	ProcessRecalculationResponse(
		input *chan *trading212_service.RecalculateSymbolsResponse,
		responseChan *chan *symbol_service.RecalculateSymbolResponse,
		errChan *chan error)
	symbolDataToEntity(in *[]*proto_models.Symbol) ([]*model.Symbol, error)
}

type SymbolsService struct {
	symbolsService
	currencyRepository *db.CurrencyRepository
	symbolsRepository  *db.SymbolRepository
}

func NewSymbolsService(symbolsRepository *db.SymbolRepository, currencyRepository *db.CurrencyRepository) *SymbolsService {
	return &SymbolsService{symbolsRepository: symbolsRepository, currencyRepository: currencyRepository}
}

func (s *SymbolsService) GetPaged(ctx context.Context, req *symbol_service.ReadPagedSymbolRequest) (*[]*proto_models.Symbol, error) {
	var res []*proto_models.Symbol
	syms, err := s.symbolsRepository.GetPaged(ctx, req)
	if err != nil {
		return nil, err
	}
	for _, sym := range *syms {
		res = append(res, sym.ToProtoObject())
	}

	return &res, nil
}

func (s *SymbolsService) InsertBulk(timeoutContext context.Context, symbols []*proto_models.Symbol) (bool, error) {
	entities, err := s.symbolDataToEntity(&symbols)
	if err != nil {
		return false, err
	}

	// begin a transaction for the whole of the insert
	tx, err := s.symbolsRepository.BeginTx(&timeoutContext, &pgx.TxOptions{})
	if err != nil {
		return false, err
	}

	// insert items in batches
	_, err = s.symbolsRepository.InsertBulk(tx, &timeoutContext, entities)
	if err != nil {
		err2 := tx.RollbackEx(timeoutContext)
		if err2 != nil {
			return false, fmt.Errorf("%v %v", err, err2)
		}
		return false, err
	}

	// commit insertion
	err = tx.CommitEx(timeoutContext)
	if err != nil {
		return false, err
	}

	return true, nil
}

func (s *SymbolsService) ProcessRecalculationResponse(
	input []*trading212_service.RecalculateSymbolsResponse,
	ctx *context.Context) (*symbol_service.RecalculateSymbolResponse, error) {

	var createSymbols []*proto_models.Symbol
	var updateSymbols []*proto_models.Symbol
	var deleteSymbols []*proto_models.Symbol

	itemsIgnored := int64(0)
	for _, res := range input {
		switch res.Type {
		case trading212_service.RecalculateSymbolsResponse_CREATE:
			createSymbols = append(createSymbols, res.Symbol)
		case trading212_service.RecalculateSymbolsResponse_UPDATE:
			updateSymbols = append(updateSymbols, res.Symbol)
		case trading212_service.RecalculateSymbolsResponse_DELETE:
			deleteSymbols = append(deleteSymbols, res.Symbol)
		case trading212_service.RecalculateSymbolsResponse_IGNORE:
			itemsIgnored++
		}
	}

	tctx, c := context.WithTimeout(*ctx, 5*time.Second)
	defer c()

	tx, err := s.symbolsRepository.BeginTx(&tctx, &pgx.TxOptions{})
	if err != nil {
		return nil, err
	}

	// create new symbols
	createEntities, err := s.symbolDataToEntity(&createSymbols)
	if err != nil {
		tx.RollbackEx(tctx)
		return nil, err
	}
	temp := make(map[string][]*model.Symbol)
	for _, sym := range createEntities {
		var u string
		_ = sym.Uuid.AssignTo(&u)
		temp[u] = append(temp[u], sym)
	}
	_, err = s.symbolsRepository.InsertBulk(tx, &tctx, createEntities)
	if err != nil {
		tx.RollbackEx(tctx)
		return nil, err
	}

	// delete entities
	deleteEntities, err := s.symbolDataToEntity(&deleteSymbols)
	if err != nil {
		tx.RollbackEx(tctx)
		return nil, err
	}
	_, err = s.symbolsRepository.DeleteBulk(tx, &tctx, deleteEntities)
	if err != nil {
		tx.RollbackEx(tctx)
		return nil, err
	}

	// update entities
	updateEntities, err := s.symbolDataToEntity(&updateSymbols)
	if err != nil {
		tx.RollbackEx(tctx)
		return nil, err
	}
	_, err = s.symbolsRepository.UpdateBulk(tx, &tctx, updateEntities)
	if err != nil {
		tx.RollbackEx(tctx)
		return nil, err
	}

	err = tx.CommitEx(tctx)
	if err != nil {
		return nil, err
	}

	return &symbol_service.RecalculateSymbolResponse{
		ItemsCreated: int64(len(createSymbols)),
		ItemsUpdated: int64(len(updateSymbols)),
		ItemsDeleted: int64(len(deleteSymbols)),
		ItemsIgnored: itemsIgnored,
		TotalItems:   int64(len(input)),
	}, nil
}

func (s *SymbolsService) symbolDataToEntity(in *[]*proto_models.Symbol) ([]*model.Symbol, error) {
	var result []*model.Symbol
	var wg sync.WaitGroup

	wg.Add(len(*in))
	for _, sym := range *in {
		e := model.Symbol{}

		// get or create the currency of the symbol
		curr, err := s.currencyRepository.GetOrCreate(sym.Currency.Code)
		if err != nil {
			return nil, err
		}

		e.Uuid = pgtype.UUID{}
		err = e.Uuid.Set(sym.Uuid)
		if err != nil {
			return nil, err
		}

		e.CurrencyID = curr.ID
		e.Isin = sym.Isin
		e.Identifier = sym.Identifier
		e.Name = sym.Name
		var moq pgtype.Float4
		_ = moq.Set(sym.MinimumOrderQuantity)
		e.MinimumOrderQuantity = moq
		e.MarketName = sym.MarketName
		e.MarketHoursGmt = sym.MarketHoursGmt

		e.CreatedAt = pgtype.Timestamptz{Time: time.Now(), Status: pgtype.Present}
		e.UpdatedAt = pgtype.Timestamptz{Time: time.Now(), Status: pgtype.Present}
		e.DeletedAt = pgtype.Timestamptz{Status: pgtype.Null}

		result = append(result, &e)
	}

	return result, nil
}
