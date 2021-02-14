package main

import (
	"analysis-api/common"
	"analysis-api/db"
	"analysis-api/infrastructure"
	"github.com/dystopia-systems/alaskalog"
)

func main() {
	c, err := common.GetConfig()
	if err != nil {
		alaskalog.Logger.Fatalf("error getting configuration: %v", err)
	}
	d, err := db.InitDb(c)
	if err != nil {
		alaskalog.Logger.Fatalf("error initializing db: %v", err)
	}
	err = db.Migrate(d)
	if err != nil {
		alaskalog.Logger.Fatalf("error migrating db: %v", err)
	}

	yahooService := infrastructure.NewYahooService()

	yahooService.GetQuote("AAPL", "2021-01-01", "2021-02-14")

	if err != nil {
		alaskalog.Logger.Fatalf("error creating from external: %v", err)
	}
}
