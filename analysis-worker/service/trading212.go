package service

import (
	"context"
	"fmt"
	"github.com/chromedp/chromedp"
	"github.com/dystopia-systems/alaskalog"
	"github.com/gofrs/uuid"
	"github.com/vectorman1/analysis/analysis-worker/common"
	"github.com/vectorman1/analysis/analysis-worker/generated/proto_models"
	"github.com/vectorman1/analysis/analysis-worker/generated/trading212_service"
	"golang.org/x/net/html"
	"golang.org/x/sync/errgroup"
	"io"
	"log"
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
	config *common.Config
	trading212_service.UnimplementedTrading212ServiceServer
}

func (s Trading212Service) New(instrumentsLink string, showMoreSelector string, instrumentsTableSelector string, cfg *common.Config) *Trading212Service {
	return &Trading212Service{
		instrumentsLink:          instrumentsLink,
		showMoreSelector:         showMoreSelector,
		instrumentsTableSelector: instrumentsTableSelector,
		config: cfg,
	}
}

func (s Trading212Service) RecalculateSymbols(srv trading212_service.Trading212Service_RecalculateSymbolsServer) error {
	ctx := srv.Context()
	alaskalog.Logger.Infoln("got request")
	var oldSymbols []*proto_models.Symbol

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		req, err := srv.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			alaskalog.Logger.Warnf("recv error: %v", err)
			return err
		}

		oldSymbols = append(oldSymbols, req)
	}

	externalData, err  := pullAndParseTrading212Data(s.instrumentsLink, s.showMoreSelector, s.instrumentsTableSelector, s.config)
	if err != nil {
		return err
	}

	alaskalog.Logger.Infoln(len(externalData), len(oldSymbols))
	res := generateResult(externalData, oldSymbols)

	for _, r := range res {
		if err := srv.Send(r); err != nil {
			alaskalog.Logger.Warnf("send error: %v", err)
			break
		}
	}

	return nil
}

func generateResult(newSymbols []*proto_models.Symbol, oldSymbols []*proto_models.Symbol) map[string]*trading212_service.RecalculateSymbolsResponse {
	unique := make(map[string]*trading212_service.RecalculateSymbolsResponse)

	// generate create, update and ignore responses
	for _, newSym := range newSymbols {
		if ok, oldSym := common.ContainsSymbol(newSym.Uuid, oldSymbols); !ok {
			if unique[newSym.Uuid] == nil {
				unique[newSym.Uuid] =
					&trading212_service.RecalculateSymbolsResponse{
					Type:   trading212_service.RecalculateSymbolsResponse_CREATE,
					Symbol: newSym,
				}
				continue
			} else {
				log.Println("collision on create: ", newSym)
				log.Println("existing: ", unique[newSym.Uuid].Type, unique[newSym.Uuid])
			}
		} else {
			// check if any fields from the new symbol are different from the old
			shouldUpdate := false
			if oldSym.Name != newSym.Name {
				shouldUpdate = true
			} else if oldSym.MinimumOrderQuantity != newSym.MinimumOrderQuantity {
				shouldUpdate = true
			} else if oldSym.MarketName != newSym.MarketName {
				shouldUpdate = true
			} else if oldSym.MarketHoursGmt != newSym.MarketHoursGmt {
				shouldUpdate = true
			}

			// if any fields are updated, send and update response, otherwise, send it back and ignore it
			if shouldUpdate {
				if unique[newSym.Uuid] == nil {
					unique[newSym.Uuid] =
						&trading212_service.RecalculateSymbolsResponse{
							Type:   trading212_service.RecalculateSymbolsResponse_UPDATE,
							Symbol: newSym,
						}
					continue
				} else {
					log.Println("collision on update: ", newSym)
					log.Println("existing: ", unique[newSym.Uuid].Type, unique[newSym.Uuid])
				}
			} else {
				if unique[newSym.Uuid] == nil {
					unique[newSym.Uuid] =
						&trading212_service.RecalculateSymbolsResponse{
							Type:   trading212_service.RecalculateSymbolsResponse_IGNORE,
							Symbol: newSym,
						}
					continue
				} else {
					log.Println("collision on ignore: ", newSym)
					log.Println("existing: ", unique[newSym.Uuid].Type, unique[newSym.Uuid])
				}
			}
		}
	}

	// generate delete responses
	for _, oldSym := range oldSymbols {
		if ok, _ := common.ContainsSymbol(oldSym.Uuid, newSymbols); !ok {
			if unique[oldSym.Uuid] == nil {
				unique[oldSym.Uuid] = &trading212_service.RecalculateSymbolsResponse{
					Type:   trading212_service.RecalculateSymbolsResponse_DELETE,
					Symbol: oldSym,
				}
			} else {
				log.Println("collision: ", oldSym)
				log.Println("existing: ", unique[oldSym.Uuid].Type, unique[oldSym.Uuid])
			}
		}
	}

	return unique
}

func pullAndParseTrading212Data(instrumentsLink, showMoreSelector, instrumentsTableSelector string, cfg *common.Config) ([]*proto_models.Symbol, error) {
	ctx, c := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(alaskalog.Logger.Infof),
	)
	defer c()

	ctx, c = context.WithTimeout(ctx, 30*time.Second)
	defer c()

	var htmlRes string
	err :=  chromedp.Run(ctx,
		chromedp.Navigate(instrumentsLink),
		chromedp.WaitVisible(showMoreSelector),
		chromedp.Click(showMoreSelector),
		chromedp.InnerHTML(instrumentsTableSelector, &htmlRes))
	if err != nil {
		alaskalog.Logger.Warnf("failed to get 212 webpage: %v", err)
		return nil, err
	}

	return parseHtmlToProtoSyms(htmlRes, cfg)
}

func parseHtmlToProtoSyms(htmlRes string, cfg *common.Config) ([]*proto_models.Symbol, error) {
	var parsedProtoSyms []*proto_models.Symbol
	var wg sync.WaitGroup

	rows, err := walkTable(htmlRes)
	if err != nil {
		return nil, err
	}

	// get results from parser worker
	parsedSymsChan := make(chan *proto_models.Symbol)
	go func(wg *sync.WaitGroup) {
		for sym := range parsedSymsChan {
			parsedProtoSyms = append(parsedProtoSyms, sym)
			wg.Done()
		}
	}(&wg)

	ctx, c1 := context.WithTimeout(context.Background(), time.Second)
	defer c1()

	g, _ := errgroup.WithContext(ctx)

	// spawn goroutine for each row
	for _, row := range rows {
		wg.Add(1)
		trow := row
		g.Go(func() error {
			sym, err := getSymbolData(trow, cfg)
			if err != nil {
				return err
			}
			parsedSymsChan <- sym
			return nil
		})
	}

	// check for any errors
	err = g.Wait()
	if err != nil {
		return nil, err
	}

	// wait for the results to be added to the array
	wg.Wait()

	return parsedProtoSyms, nil
}

// getSymbolData reads a row from the table and parses it into a proto struct
func getSymbolData(row []string, cfg *common.Config) (*proto_models.Symbol, error) {
	instrumentName := strings.TrimSpace(row[0])
	companyName := strings.TrimSpace(row[1])
	currencyCode := strings.TrimSpace(row[2])
	isin := strings.TrimSpace(row[3])
	minTradedQuantity, _ := strconv.ParseFloat(strings.TrimSpace(row[4]), 32)
	roundedMinQuantity := float32(math.Round(minTradedQuantity*1000)/1000)
	marketName := strings.TrimSpace(row[5])
	marketHours := strings.TrimSpace(row[6])

	ns, err := uuid.FromString(cfg.SymbolsNamespace)
	if err != nil {
		return nil, err
	}

	str := fmt.Sprintf("%s,%s,%s", isin, instrumentName, marketName)
	u := uuid.NewV5(ns, str)
	us := u.String()

	return &proto_models.Symbol{
		Uuid: us,
		Isin:                 isin,
		Identifier:           instrumentName,
		Name:                 companyName,
		Currency:             &proto_models.Currency{
			Code: currencyCode,
		},
		MinimumOrderQuantity: roundedMinQuantity,
		MarketName:           marketName,
		MarketHoursGmt:       marketHours,
	}, nil
}

// walkTable recursively walks the table of instruments received and returns it as a splice of splices
func walkTable(htmlRes string) ([][]string, error) {
	doc, err := html.Parse(strings.NewReader(htmlRes))
	if err != nil {
		return nil, err
	}

	var symbolRows [][]string

	visitNode := func(n *html.Node) {
		if n.Type == html.ElementNode &&  n.Data == "div" {
			for _, a := range n.Attr {
				if a.Key == "id" && strings.Contains(a.Val, "equity-row-") {
					var row []string
					instrumentName := n.FirstChild
					companyName := instrumentName.NextSibling
					currencyCode := companyName.NextSibling
					isin := currencyCode.NextSibling
					minTradedQuantity := isin.NextSibling
					marketName := minTradedQuantity.NextSibling
					marketHours := marketName.NextSibling

					row = append(row,
						[]string{
							instrumentName.FirstChild.Data,
							companyName.FirstChild.Data,
							currencyCode.FirstChild.Data,
							isin.FirstChild.Data,
							minTradedQuantity.FirstChild.Data,
							marketName.FirstChild.Data,
							strings.TrimSpace(marketHours.FirstChild.Data)}...)
					symbolRows = append(symbolRows, row)
				}
			}
		}
	}

	forEachNode(doc, visitNode, nil)

	return symbolRows, nil
}

// Copied from gopl.io/ch5/outline2.
func forEachNode(n *html.Node, pre, post func(n *html.Node)) {
	if pre != nil {
		pre(n)
	}
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		forEachNode(c, pre, post)
	}
	if post != nil {
		post(n)
	}
}