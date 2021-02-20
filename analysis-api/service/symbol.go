package service

import (
	"context"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/dystopia-systems/alaskalog"
	"github.com/jackc/pgx/pgtype"
	"github.com/vectorman1/analysis/analysis-api/model"

	"github.com/vectorman1/analysis/analysis-api/generated/proto_models"

	"github.com/vectorman1/analysis/analysis-api/generated/symbol_service"
	"github.com/vectorman1/analysis/analysis-api/generated/trading212_service"

	"github.com/vectorman1/analysis/analysis-api/common"

	"github.com/vectorman1/analysis/analysis-api/db"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type SymbolsServiceServer struct {
	rpcClient          *common.Rpc
	symbolsRepository  *db.SymbolRepository
	currencyRepository *db.CurrencyRepository
	symbol_service.UnimplementedSymbolServiceServer
}

func NewSymbolsServiceServer(rpcClient *common.Rpc, symbolsRepository *db.SymbolRepository, currencyRepository *db.CurrencyRepository) *SymbolsServiceServer {
	return &SymbolsServiceServer{
		rpcClient:          rpcClient,
		symbolsRepository:  symbolsRepository,
		currencyRepository: currencyRepository,
	}
}

func (s *SymbolsServiceServer) ReadPaged(ctx context.Context, req *symbol_service.ReadPagedSymbolRequest) (*symbol_service.ReadPagedSymbolResponse, error) {
	if req.Filter == nil {
		return nil, status.Errorf(http.StatusBadRequest, "bad request")
	}

	entityResult, err := s.symbolsRepository.GetPaged(ctx, req)
	if err != nil {
		return nil, status.Error(codes.Unavailable, err.Error())
	}

	protoSyms := &proto_models.Symbols{Symbols: []*proto_models.Symbol{}}
	for _, e := range *entityResult {
		protoSyms.Symbols = append(protoSyms.Symbols, e.ToProtoObject())
	}

	resp := &symbol_service.ReadPagedSymbolResponse{Symbols: protoSyms}
	return resp, nil
}

func (s *SymbolsServiceServer) Read(ctx context.Context, req *symbol_service.ReadSymbolRequest) (*symbol_service.ReadSymbolResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Read not implemented")
}

func (s *SymbolsServiceServer) Create(ctx context.Context, req *symbol_service.CreateSymbolRequest) (*symbol_service.CreateSymbolResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Create not implemented")
}

func (s *SymbolsServiceServer) Update(ctx context.Context, req *symbol_service.UpdateSymbolRequest) (*symbol_service.UpdateSymbolResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Update not implemented")
}

func (s *SymbolsServiceServer) Delete(ctx context.Context, req *symbol_service.DeleteSymbolRequest) (*symbol_service.DeleteSymbolResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method Delete not implemented")
}

func (s *SymbolsServiceServer) Populate(ctx context.Context, req *symbol_service.PopulateSymbolRequest) (*symbol_service.PopulateSymbolResponse, error) {
	timeoutContext, c := context.WithTimeout(ctx, 10*time.Second)
	defer c()

	client := trading212_service.NewTrading212ServiceClient(s.rpcClient.Connection)
	stream, err := client.GetSymbols(timeoutContext, &trading212_service.GetRequest{})

	var protoSyms []*proto_models.Symbol
	for {
		res, err := stream.Recv()
		if err != nil {
			if err != io.EOF {
				return nil, status.Errorf(status.Code(err), "error while reading stream: %v", err)
			}
			break
		}
		protoSyms = append(protoSyms, res)
	}

	syms := s.symbolDataToEntity(&protoSyms)

	_, err = s.symbolsRepository.InsertBulk(timeoutContext, syms)
	if err != nil {
		return nil, status.Errorf(codes.Canceled, "error while inserting symbols %v", err)
	}

	return &symbol_service.PopulateSymbolResponse{}, nil
}

func (s *SymbolsServiceServer) symbolDataToEntity(in *[]*proto_models.Symbol) []*model.Symbol {
	var result []*model.Symbol
	var wg sync.WaitGroup

	queue := make(chan *model.Symbol, 1)

	wg.Add(len(*in))
	for _, sym := range *in {
		go func(protoSym *proto_models.Symbol) {
			e := model.Symbol{}
			curr, err := s.currencyRepository.GetByCode(protoSym.Currency.Code)
			if err != nil {
				c := &model.Currency{}
				c.Code = protoSym.Currency.Code
				c.LongName = "temp name"

				id, createErr := s.currencyRepository.Create(c)
				if createErr != nil {
					alaskalog.Logger.Warnf("failed creating currency: %v", err)
					return
				}

				e.CurrencyID = id
			} else {
				e.CurrencyID = curr.ID
			}

			e.Isin = protoSym.Isin
			e.Identifier = protoSym.Identifier
			e.Name = protoSym.Name
			var moq pgtype.Float4
			_ = moq.Set(protoSym.MinimumOrderQuantity)
			e.MinimumOrderQuantity = moq
			e.MarketName = protoSym.MarketName
			e.MarketHoursGmt = protoSym.MarketHoursGmt

			e.CreatedAt = pgtype.Timestamptz{Time: time.Now(), Status: pgtype.Present}
			e.UpdatedAt = pgtype.Timestamptz{Time: time.Now(), Status: pgtype.Present}
			e.DeletedAt = pgtype.Timestamptz{Status: pgtype.Null}

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
