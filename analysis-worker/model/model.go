package model

type ExternalSymbol struct {
	Instrument        string
	Company           string
	CurrencyCode      string
	ISIN              string
	MinTradedQuantity string
	MarketName        string
	MarketHoursGMT    string
}

