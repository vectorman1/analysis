package infrastructure

import (
	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
	"time"
)
import "github.com/markcheno/go-quote"

type YahooService interface {
	GetQuote(symbol string, startDate string, endDate string)
}

type yahooService struct {
	YahooService
}

func NewYahooService() *yahooService {
	return &yahooService{}
}

func (s *yahooService) GetQuote(symbol string, startDate string, endDate string) {
	q, _ := quote.NewQuoteFromYahoo(symbol, startDate, endDate, quote.Daily, true)
	ts := techan.NewTimeSeries()

	for i, v := range q.Date {
		period := techan.NewTimePeriod(time.Unix(v.UnixNano(), 0), time.Hour*24)

		candle := techan.NewCandle(period)
		candle.OpenPrice = big.NewDecimal(q.Open[i])
		candle.ClosePrice = big.NewDecimal(q.Close[i])
		candle.MaxPrice = big.NewDecimal(q.High[i])
		candle.MinPrice = big.NewDecimal(q.Low[i])
		candle.Volume = big.NewDecimal(q.Volume[i])

		ts.AddCandle(candle)
	}

	closePriceIndicator := techan.NewClosePriceIndicator(ts)
	sma5 := techan.NewSimpleMovingAverage(closePriceIndicator, 5)
	ema5 := techan.NewEMAIndicator(closePriceIndicator, 5)

	println(`ema`, ema5.Calculate(len(ts.Candles)-1).FormattedString(4))
	println(`sma`, sma5.Calculate(len(ts.Candles)-1).FormattedString(4))
}
