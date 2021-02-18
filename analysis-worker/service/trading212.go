package service

import (
	"context"
	"github.com/chromedp/chromedp"
	"github.com/dystopia-systems/alaskalog"
	"github.com/vectorman1/analysis/analysis-worker/common"
	"github.com/vectorman1/analysis/analysis-worker/model"
	"github.com/vectorman1/analysis/analysis-worker/proto"
	"golang.org/x/net/html"
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
	proto.UnimplementedTrading212ServiceServer
}

func (s Trading212Service) New(instrumentsLink string, showMoreSelector string, instrumentsTableSelector string) *Trading212Service {
	return &Trading212Service{
		instrumentsLink:          instrumentsLink,
		showMoreSelector:         showMoreSelector,
		instrumentsTableSelector: instrumentsTableSelector,
	}
}

func (s *Trading212Service) GetUpdatedSymbols(data *proto.Symbols, srv proto.Trading212Service_GetUpdatedSymbolsServer) error {
	ctx := srv.Context()
	alaskalog.Logger.Infoln("got request")
	for {
		externalData, externalDataErr := pullAndParseTrading212Data(s.instrumentsLink, s.showMoreSelector, s.instrumentsTableSelector)
		if externalDataErr != nil {
			alaskalog.Logger.Warnf("failed reading external data: %v", externalDataErr)
			ctx.Done()
		}

		res := generateResult(externalData, data)

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer alaskalog.Logger.Infoln("sent data back to client")
			for r := range res {
				if err := srv.Send(r); err != nil {
					alaskalog.Logger.Warnf("send error: %v", err)
					return
				}
				print(" sent: ", r.Identifier)
			}
		}()


		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		return nil
	}
}

func generateResult(newSymbolsChan []*proto.Symbol, oldSymbols *proto.Symbols) <-chan *proto.Symbol {
	var newSymbolsData []*proto.Symbol
	var mainWg sync.WaitGroup

	res := make(chan *proto.Symbol)

	mainWg.Add(1)
	go func() {
		defer mainWg.Done()
		for _, s := range newSymbolsChan {
			newSymbolsData = append(newSymbolsData, s)
		}
		println("new symbols loaded")
	}()

	mainWg.Wait()

	go func() {
		defer close(res)
		defer println("generated new result")
		oldDeletedSyms := make(chan *proto.Symbol)
		newAndUpdatedSyms := make(chan *proto.Symbol)

		// set symbols which are no longer active
		go func() {
			defer close(oldDeletedSyms)
			var wg sync.WaitGroup
			for _, oldSym := range oldSymbols.Symbols {
				wg.Add(1)
				go func(oldSym *proto.Symbol) {
					defer wg.Done()
					if ok, _ := common.ContainsSymbol(oldSym.ISIN, oldSym.Identifier, newSymbolsData); !ok {
						oldSym.DeletedAt = time.Now().Unix()
						oldDeletedSyms <- oldSym
					}
				}(oldSym)
			}
			wg.Wait()
		}()

		// insert new and update old symbols
		go func() {
			defer close(newAndUpdatedSyms)
			var wg sync.WaitGroup
			for _, newSym := range newSymbolsData {
				wg.Add(1)
				go func(newSym *proto.Symbol) {
					defer wg.Done()
					if ok, oldSym := common.ContainsSymbol(newSym.ISIN, newSym.Identifier, oldSymbols.Symbols); !ok {
						newSym.CreatedAt = time.Now().Unix()
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
							oldSym.Currency = &proto.Currency{Code: newSym.Currency.Code}
						}
						if oldSym.MinimumOrderQuantity != newSym.MinimumOrderQuantity {
							shouldUpdate = true
							oldSym.MinimumOrderQuantity = newSym.MinimumOrderQuantity
						}
						if shouldUpdate {
							oldSym.UpdatedAt = time.Now().Unix()
						}
						newAndUpdatedSyms <- oldSym
					}
				}(newSym)
			}
			wg.Wait()
		}()

		for s := range oldDeletedSyms {
			res <- s
		}
		for s := range newAndUpdatedSyms {
			res <- s
		}
	}()

	return res
}

func pullAndParseTrading212Data(instrumentsLink, showMoreSelector, instrumentsTableSelector string) ([]*proto.Symbol, error) {
	ctx, c := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(alaskalog.Logger.Infof),
	)
	defer c()

	ctx, c = context.WithTimeout(ctx, 10*time.Second)
	defer c()

	var htmlRes string
	err := chromedp.Run(ctx,
		chromedp.Navigate(instrumentsLink),
		chromedp.WaitVisible(showMoreSelector),
		chromedp.Click(showMoreSelector),
		chromedp.InnerHTML(instrumentsTableSelector, &htmlRes))
	if err != nil {
		return nil, err
	}

	instrumentRows := walkTable(htmlRes)
	externalSymbols := getExternalSymbols(instrumentRows)

	var res []*proto.Symbol
	var wg sync.WaitGroup
	for ext := range externalSymbols {
		wg.Add(1)
		go func(ext model.ExternalSymbol, wg *sync.WaitGroup) {
			defer wg.Done()
			print("getting symbolData for ", ext.Instrument)
			sym := getSymbolData(ext)
			res = append(res, sym)
		}(ext, &wg)
	}
	wg.Wait()
	return res, nil
}

func getSymbolData(ext model.ExternalSymbol) *proto.Symbol {
	minQuantity, _ := strconv.ParseFloat(ext.MinTradedQuantity, 32)
	roundedMinQuantity := float32(math.Round(minQuantity*1000/1000))
	return &proto.Symbol{
		ISIN:                 ext.ISIN,
		Identifier:           ext.Company,
		Name:                 ext.Company,
		Currency:             &proto.Currency{
			Code: ext.CurrencyCode,
		},
		MinimumOrderQuantity: roundedMinQuantity,
		MarketName:           ext.MarketName,
		MarketHoursGMT:       ext.MarketHoursGMT,
	}
}

func getExternalSymbols(rows <-chan []string) <-chan model.ExternalSymbol {
	res := make(chan model.ExternalSymbol)
	go func() {
		defer close(res)
		for row := range rows {
			instrumentName := row[0]
			companyName := row[1]
			currencyCode := row[2]
			isin := row[3]
			minTradedQuantity := row[4]
			marketName := row[5]
			marketHours := row[6]

			i := model.ExternalSymbol{
				Instrument:        strings.TrimSpace(instrumentName),
				Company:           strings.TrimSpace(companyName),
				CurrencyCode:      strings.TrimSpace(currencyCode),
				ISIN:              strings.TrimSpace(isin),
				MinTradedQuantity: strings.TrimSpace(minTradedQuantity),
				MarketName:        strings.TrimSpace(marketName),
				MarketHoursGMT:    strings.TrimSpace(marketHours),
			}

			res <- i
		}
	}()

	return res
}

func walkTable(htmlRes string) <-chan []string {
	res := make(chan []string)
	go func() {
		defer close(res)
		z := html.NewTokenizer(strings.NewReader(htmlRes))
		for z.Next() != html.ErrorToken {
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