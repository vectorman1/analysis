package service

import (
	"context"
	"errors"
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
		externalData := pullAndParseTrading212Data(s.instrumentsLink, s.showMoreSelector, s.instrumentsTableSelector)
		if externalData == nil {
			ctx.Done()
		}

		res := make(<-chan *proto.Symbol)
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

func (s *Trading212Service) GetSymbols(data *proto.GetRequest, srv proto.Trading212Service_GetSymbolsServer) error {
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

func generateResult(newSymbolsChan <-chan *proto.Symbol, oldSymbols *proto.Symbols) <-chan *proto.Symbol {
	if oldSymbols.Symbols == nil {
		return newSymbolsChan
	} else {
		var newSymbolsData []*proto.Symbol

		for s := range newSymbolsChan {
			newSymbolsData = append(newSymbolsData, s)
		}

		res := make(chan *proto.Symbol)
		oldDeletedSyms := make(chan *proto.Symbol)
		newAndUpdatedSyms := make(chan *proto.Symbol)

		go func() {
			defer close(oldDeletedSyms)
			defer println("set deleted syms")
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

func pullAndParseTrading212Data(instrumentsLink, showMoreSelector, instrumentsTableSelector string) <-chan *proto.Symbol {
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

func parseHtmlToProtoSyms(htmlRes string) <-chan *proto.Symbol {
	parsedProtoSyms := make(chan *proto.Symbol)
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

func getSymbolData(row []string) proto.Symbol {
	instrumentName := strings.TrimSpace(row[0])
	companyName := strings.TrimSpace(row[1])
	currencyCode := strings.TrimSpace(row[2])
	isin := strings.TrimSpace(row[3])
	minTradedQuantity, _ := strconv.ParseFloat(strings.TrimSpace(row[4]), 32)
	roundedMinQuantity := float32(math.Round(minTradedQuantity*1000)/1000)
	marketName := strings.TrimSpace(row[5])
	marketHours := strings.TrimSpace(row[6])

	return proto.Symbol{
		ISIN:                 isin,
		Identifier:           instrumentName,
		Name:                 companyName,
		Currency:             &proto.Currency{
			Code: currencyCode,
		},
		MinimumOrderQuantity: roundedMinQuantity,
		MarketName:           marketName,
		MarketHoursGMT:       marketHours,
	}
}

func getExternalSymbols(rows <-chan []string) <-chan *model.ExternalSymbol {
	res := make(chan *model.ExternalSymbol)
	go func() {
		defer close(res)
		var wg sync.WaitGroup
		for row := range rows {
			wg.Add(1)
			go func(row []string) {
				defer wg.Done()
				instrumentName := row[0]
				companyName := row[1]
				currencyCode := row[2]
				isin := row[3]
				minTradedQuantity := row[4]
				marketName := row[5]
				marketHours := row[6]

				i := &model.ExternalSymbol{
					Instrument:        strings.TrimSpace(instrumentName),
					Company:           strings.TrimSpace(companyName),
					CurrencyCode:      strings.TrimSpace(currencyCode),
					ISIN:              strings.TrimSpace(isin),
					MinTradedQuantity: strings.TrimSpace(minTradedQuantity),
					MarketName:        strings.TrimSpace(marketName),
					MarketHoursGMT:    strings.TrimSpace(marketHours),
				}
				res <- i
			}(row)
		}
		wg.Wait()
	}()

	return res
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