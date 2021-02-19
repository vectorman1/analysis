package infrastructure

import (
	"context"
	"github.com/dystopia-systems/alaskalog"
	"github.com/vectorman1/analysis/analysis-api/infrastructure/proto"
	"github.com/vectorman1/analysis/analysis-api/service"
	"io"
	"time"
)

type UpdateResult struct {
	Success bool
}

type trading212Service interface {
	GetUpdatedSymbols() (chan bool, context.Context)
	GetSymbols() error
}

type Trading212Service struct {
	trading212Service
	instrumentsLink          string
	showMoreSelector         string
	instrumentsTableSelector string
	symbolsService           *service.SymbolService
	rpcClient                *Rpc
}

func (s Trading212Service) New(
	instrumentsLink string,
	showMoreSelector string,
	instrumentsTableSelector string,
	symbolsService *service.SymbolService,
	rpcClient *Rpc) *Trading212Service {
	return &Trading212Service{
		instrumentsLink:          instrumentsLink,
		showMoreSelector:         showMoreSelector,
		instrumentsTableSelector: instrumentsTableSelector,
		symbolsService:           symbolsService,
		rpcClient:                rpcClient,
	}
}

func (s *Trading212Service) GetUpdatedSymbols() (bool, context.Context) {
	return false, nil
}

func (s *Trading212Service) GetSymbols() (bool, error) {
	timeoutContext, c := context.WithTimeout(context.Background(), 15*time.Second)
	defer c()

	client := proto.NewTrading212ServiceClient(s.rpcClient.connection)
	stream, err := client.GetSymbols(timeoutContext, &proto.GetRequest{})
	if err != nil {
		return false, err
	}

	var protoSyms []*proto.Symbol
	for {
		resp, e := stream.Recv()
		if e != nil {
			if e != io.EOF {
				alaskalog.Logger.Warnf("error while reading stream: %v", e.Error())
			}
			break
		}
		protoSyms = append(protoSyms, resp)
		alaskalog.Logger.Infoln("received ", resp.Identifier)
	}

	if err := stream.CloseSend(); err != nil {
		alaskalog.Logger.Warnf("error closing stream")
	}

	_, err = s.symbolsService.InsertBulk(&protoSyms)
	if err != nil {
		return false, err
	}
	return true, nil
}
