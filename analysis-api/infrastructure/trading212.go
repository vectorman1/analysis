package infrastructure

import (
	"context"
	"github.com/chromedp/chromedp"
	"github.com/dystopia-systems/alaskalog"
	"golang.org/x/net/html"
	"strings"
	"time"
)

type Trading212 interface {
	GetPublishedSymbols() ([]ExternalSymbol, error)
}

type ExternalSymbol struct {
	Instrument        string
	Company           string
	CurrencyCode      string
	ISIN              string
	MinTradedQuantity string
	MarketName        string
	MarketHoursGMT    string
}

type Trading212Service struct {
	instrumentsLink          string
	showMoreSelector         string
	instrumentsTableSelector string
}

func NewTrading212Service(instrumentsLink string, showMoreSelector string, instrumentsTableSelector string) *Trading212Service {
	return &Trading212Service{
		instrumentsLink:          instrumentsLink,
		showMoreSelector:         showMoreSelector,
		instrumentsTableSelector: instrumentsTableSelector,
	}
}

func (s *Trading212Service) GetPublishedSymbols() ([]ExternalSymbol, error) {
	ctx, c := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(alaskalog.Logger.Infof),
	)
	defer c()

	ctx, c = context.WithTimeout(ctx, 10*time.Second)
	defer c()

	var htmlRes string
	err := chromedp.Run(ctx,
		chromedp.Navigate(s.instrumentsLink),
		chromedp.WaitVisible(s.showMoreSelector),
		chromedp.Click(s.showMoreSelector),
		chromedp.InnerHTML(s.instrumentsTableSelector, &htmlRes))
	if err != nil {
		return nil, err
	}

	externalInstruments, err := parseInstrumentsTable(htmlRes)
	if err != nil {
		return nil, err
	}

	return externalInstruments, nil
}

func parseInstrumentsTable(s string) ([]ExternalSymbol, error) {
	node, err := html.Parse(strings.NewReader(s))
	if err != nil {
		return nil, err
	}

	var res []ExternalSymbol
	res = traverseNodes(&res, node)

	return res, nil
}

func traverseNodes(res *[]ExternalSymbol, n *html.Node) []ExternalSymbol {
	if n == nil {
		return *res
	} else {
		if n.Data == "div" {
			for _, a := range n.Attr {
				if a.Key == "id" && strings.Contains(a.Val, "equity-row-") {
					instrumentName := n.FirstChild
					companyName := instrumentName.NextSibling
					currencyCode := companyName.NextSibling
					isin := currencyCode.NextSibling
					minTradedQuantity := isin.NextSibling
					marketName := minTradedQuantity.NextSibling
					marketHours := marketName.NextSibling

					i := ExternalSymbol{
						Instrument:        strings.TrimSpace(instrumentName.FirstChild.Data),
						Company:           strings.TrimSpace(companyName.FirstChild.Data),
						CurrencyCode:      strings.TrimSpace(currencyCode.FirstChild.Data),
						ISIN:              strings.TrimSpace(isin.FirstChild.Data),
						MinTradedQuantity: strings.TrimSpace(minTradedQuantity.FirstChild.Data),
						MarketName:        strings.TrimSpace(marketName.FirstChild.Data),
						MarketHoursGMT:    strings.TrimSpace(marketHours.FirstChild.Data),
					}

					*res = append(*res, i)
				}
			}
		}

		traverseNodes(res, n.FirstChild)
		traverseNodes(res, n.NextSibling)
	}

	return *res
}
