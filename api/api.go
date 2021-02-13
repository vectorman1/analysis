package main

import (
	"analysis-api/common"
	"analysis-api/db"
	"analysis-api/infrastructure"
	"analysis-api/service"
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

	currencyRepository := db.NewCurrencyRepository(d)
	symbolRepository := db.NewSymbolRepository(d)

	symbolService := service.NewSymbolService(symbolRepository, currencyRepository)

	trading212Service := infrastructure.NewTrading212Service(
		common.TRADING212_INSTRUMENTS_LINK,
		common.TRADING212_SHOW_ALL_BUTTON_SELECTOR,
		common.TRADING212_ALL_INSTRUMENTS_SELECTOR,
	)

	es, err := trading212Service.GetPublishedSymbols()
	if err != nil {
		alaskalog.Logger.Fatalf("error getting external: %v", err)
	}
	err = symbolService.CreateBatchFromExternal(es)

	if err != nil {
		alaskalog.Logger.Fatalf("error creating from external: %v", err)
	}
}
