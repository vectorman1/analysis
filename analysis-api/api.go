package main

import (
	"github.com/dystopia-systems/alaskalog"
	"github.com/labstack/echo"
	"github.com/vectorman1/analysis/analysis-api/common"
	"github.com/vectorman1/analysis/analysis-api/db"
	_db "github.com/vectorman1/analysis/analysis-api/db"
	_delivery "github.com/vectorman1/analysis/analysis-api/delivery"
	"github.com/vectorman1/analysis/analysis-api/infrastructure"
	"github.com/vectorman1/analysis/analysis-api/service"
)

func main() {
	alaskalog.Configure()
	c, err := common.GetConfig()
	if err != nil {
		alaskalog.Logger.Fatalf("error getting configuration: %v", err)
	}

	err = db.Migrate(c)
	if err != nil {
		alaskalog.Logger.Fatalf("error migrating db: %v", err)
	}

	pgDb, err := db.GetPgDb(c)
	if err != nil {
		alaskalog.Logger.Fatalf("error initializing pure db: %v", err)
	}

	rpcClient := &infrastructure.Rpc{}
	_, err = rpcClient.Initialize()
	if err != nil {
		alaskalog.Logger.Fatalf("failed to connect to rpcClient server: %v", err)
	}

	e := echo.New()
	e.HideBanner = true

	symbolRepo := _db.NewSymbolRepository(pgDb)
	currencyRepo := _db.NewCurrencyRepository(pgDb)

	symbolService := service.NewSymbolService(symbolRepo, currencyRepo)

	trading212Service := infrastructure.Trading212Service{}.New(
		common.TRADING212_INSTRUMENTS_LINK,
		common.TRADING212_SHOW_ALL_BUTTON_SELECTOR,
		common.TRADING212_ALL_INSTRUMENTS_SELECTOR,
		symbolService,
		rpcClient)

	_delivery.NewSymbolHandler(e, symbolService, trading212Service)

	alaskalog.Logger.Fatal(e.Start(":7070"))
}
