package service

import (
	"context"
	"errors"
	"github.com/chromedp/chromedp"
	"github.com/dystopia-systems/alaskalog"
	"github.com/vectorman1/analysis/analysis-worker/common"
	"github.com/vectorman1/analysis/analysis-worker/generated/proto_models"
	"github.com/vectorman1/analysis/analysis-worker/generated/trading212_service"
	"golang.org/x/net/html"
	"google.golang.org/protobuf/types/known/timestamppb"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Trading212Service struct {
	instrumentsLink          string
	showMoreSelector         string
	instrumentsTableSelector string
	trading212_service.UnimplementedTrading212ServiceServer
}

func (s Trading212Service) New(instrumentsLink string, showMoreSelector string, instrumentsTableSelector string) *Trading212Service {
	return &Trading212Service{
		instrumentsLink:          instrumentsLink,
		showMoreSelector:         showMoreSelector,
		instrumentsTableSelector: instrumentsTableSelector,
	}
}

func (s *Trading212Service) GetUpdatedSymbols(data *proto_models.Symbols, srv trading212_service.Trading212Service_GetUpdatedSymbolsServer) error {
	ctx := srv.Context()
	alaskalog.Logger.Infoln("got request")
	for {
		externalData := pullAndParseTrading212Data(s.instrumentsLink, s.showMoreSelector, s.instrumentsTableSelector)
		if externalData == nil {
			ctx.Done()
		}

		res := make(<-chan *proto_models.Symbol)
		if data.Symbols == nil {
			res = externalData
		} else {
			res = generateResult(externalData, data)
		}

		for r := range res {
			if err := srv.Send(r); err != nil {
				alaskalog.Logger.Warnf("send error: %v", err)
				break
			}
			print(" sent: ", r.Identifier)
		}

		select{

		}
	}
}

func (s *Trading212Service) GetSymbols(data *trading212_service.GetRequest, srv trading212_service.Trading212Service_GetSymbolsServer) error {
	res := pullAndParseTrading212Data(s.instrumentsLink, s.showMoreSelector, s.instrumentsTableSelector)

	if res == nil {
		return errors.New("couldn't get t212 data")
	}

	for symbol := range res {
		if err := srv.Send(symbol); err != nil {
			alaskalog.Logger.Warnf("error while sending: %v", err)
			return err
		}
	}
	println("finished sending response")
	return nil
}

func generateResult(newSymbolsChan <-chan *proto_models.Symbol, oldSymbols *proto_models.Symbols) <-chan *proto_models.Symbol {
	if oldSymbols.Symbols == nil {
		return newSymbolsChan
	} else {
		var newSymbolsData []*proto_models.Symbol

		for s := range newSymbolsChan {
			newSymbolsData = append(newSymbolsData, s)
		}

		res := make(chan *proto_models.Symbol)
		oldDeletedSyms := make(chan *proto_models.Symbol)
		newAndUpdatedSyms := make(chan *proto_models.Symbol)

		go func() {
			defer close(oldDeletedSyms)
			defer println("set deleted syms")
			var wg sync.WaitGroup
			for _, oldSym := range oldSymbols.Symbols {
				wg.Add(1)
				go func(oldSym *proto_models.Symbol) {
					defer wg.Done()
					if ok, _ := common.ContainsSymbol(oldSym.ISIN, oldSym.Identifier, newSymbolsData); !ok {
						oldSym.DeletedAt = timestamppb.Now()
						oldDeletedSyms <- oldSym
					}
				}(oldSym)
			}
			wg.Wait()
		}()
		go func() {
			defer close(newAndUpdatedSyms)
			defer println("inserting and updating")
			var wg sync.WaitGroup
			for i := 0; i < common.MaxConcurrency; i++ {
				wg.Add(1)
				go func(wg *sync.WaitGroup) {
					defer wg.Done()
					for _, newSym := range newSymbolsData {
						if ok, oldSym := common.ContainsSymbol(newSym.ISIN, newSym.Identifier, oldSymbols.Symbols); !ok {
							newSym.CreatedAt = timestamppb.Now()
							newAndUpdatedSyms <- newSym
						} else {
							shouldUpdate := false
							if oldSym.Name != newSym.Name {
								shouldUpdate = true
								oldSym.Name = newSym.Name
							}
							if oldSym.MarketName != newSym.MarketName {
								shouldUpdate = true
								oldSym.MarketName = newSym.MarketName
							}
							if oldSym.MarketHoursGMT != newSym.MarketHoursGMT {
								shouldUpdate = true
								oldSym.MarketHoursGMT = newSym.MarketHoursGMT
							}
							if oldSym.Currency.Code != newSym.Currency.Code {
								shouldUpdate = true
								oldSym.Currency = &proto_models.Currency{Code: newSym.Currency.Code}
							}
							if oldSym.MinimumOrderQuantity != newSym.MinimumOrderQuantity {
								shouldUpdate = true
								oldSym.MinimumOrderQuantity = newSym.MinimumOrderQuantity
							}
							if shouldUpdate {
								t := timestamppb.New(time.Now())
								oldSym.UpdatedAt = t
							}
							newAndUpdatedSyms <- oldSym
						}
					}
				}(&wg)
			}
			wg.Wait()
		}()

		go func() {
			for s := range oldDeletedSyms {
				res <- s
			}
			for s := range newAndUpdatedSyms {
				res <- s
			}
		}()

		return res
	}
}

func pullAndParseTrading212Data(instrumentsLink, showMoreSelector, instrumentsTableSelector string) <-chan *proto_models.Symbol {
	ctx, c := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(alaskalog.Logger.Infof),
	)
	defer c()

	ctx, c = context.WithTimeout(ctx, 15*time.Second)
	defer c()

	var htmlRes string
	err :=  chromedp.Run(ctx,
		chromedp.Navigate(instrumentsLink),
		chromedp.WaitVisible(showMoreSelector),
		chromedp.Click(showMoreSelector),
		chromedp.InnerHTML(instrumentsTableSelector, &htmlRes))
	if err != nil {
		alaskalog.Logger.Warnf("failed to get 212 webpage: %v", err)
		return nil
	}

	return parseHtmlToProtoSyms(htmlRes)
}

func parseHtmlToProtoSyms(htmlRes string) <-chan *proto_models.Symbol {
	parsedProtoSyms := make(chan *proto_models.Symbol)
	rows := walkTable(htmlRes)

	go func() {
		defer close(parsedProtoSyms)
		defer println("closing parsed syms")
		var wg sync.WaitGroup
		for i := 0; i < common.MaxConcurrency; i++ {
			wg.Add(1)
			go func(wg *sync.WaitGroup) {
				defer wg.Done()
				for row := range rows {
					sym := getSymbolData(row)
					parsedProtoSyms <- &sym
				}
			}(&wg)
		}
		wg.Wait()
	}()

	return parsedProtoSyms
}

func getSymbolData(row []string) proto_models.Symbol {
	instrumentName := strings.TrimSpace(row[0])
	companyName := strings.TrimSpace(row[1])
	currencyCode := strings.TrimSpace(row[2])
	isin := strings.TrimSpace(row[3])
	minTradedQuantity, _ := strconv.ParseFloat(strings.TrimSpace(row[4]), 32)
	roundedMinQuantity := float32(math.Round(minTradedQuantity*1000)/1000)
	marketName := strings.TrimSpace(row[5])
	marketHours := strings.TrimSpace(row[6])

	return proto_models.Symbol{
		ISIN:                 isin,
		Identifier:           instrumentName,
		Name:                 companyName,
		Currency:             &proto_models.Currency{
			Code: currencyCode,
		},
		MinimumOrderQuantity: roundedMinQuantity,
		MarketName:           marketName,
		MarketHoursGMT:       marketHours,
	}
}

func walkTable(htmlRes string) <-chan []string {
	res := make(chan []string)

	go func() {
		defer close(res)
		z := html.NewTokenizer(strings.NewReader(htmlRes))
		for tt := z.Next();
			tt != html.ErrorToken;
			tt = z.Next() {
			t := z.Token()
			switch t.Type {
			case html.StartTagToken:
				if t.Data == "div" {
					for _, a := range t.Attr {
						if a.Key == "id" && strings.Contains(a.Val, "equity-row-") {
							// TODO: Figure out better logic.
							z.Next()
							z.Next()
							instrumentName := z.Token().Data
							z.Next()
							z.Next()
							z.Next()
							companyName := z.Token().Data
							z.Next()
							z.Next()
							z.Next()
							currencyCode := z.Token().Data
							z.Next()
							z.Next()
							z.Next()
							isinNode := z.Token().Data
							z.Next()
							z.Next()
							z.Next()
							minTradedQuantityNode := z.Token().Data
							z.Next()
							z.Next()
							z.Next()
							marketName := z.Token().Data
							z.Next()
							z.Next()
							z.Next()
							marketHours := z.Token().Data
							res <- []string{
								instrumentName,
								companyName,
								currencyCode,
								isinNode,
								minTradedQuantityNode,
								marketName,
								marketHours,
							}
						}
					}
				}
			}
		}
	}()

	return res
}