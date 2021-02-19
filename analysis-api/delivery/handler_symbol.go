package delivery

import (
	"github.com/vectorman1/analysis/analysis-api/common"
	"github.com/vectorman1/analysis/analysis-api/infrastructure"
	"github.com/vectorman1/analysis/analysis-api/service"
	"net/http"
	"strconv"
)
import "github.com/labstack/echo"

type SymbolHandler struct {
	symbolService     *service.SymbolService
	trading212Service *infrastructure.Trading212Service
}

func NewSymbolHandler(e *echo.Echo, symbolService *service.SymbolService, trading212Service *infrastructure.Trading212Service) {
	handler := SymbolHandler{
		symbolService:     symbolService,
		trading212Service: trading212Service,
	}

	e.GET(Symbols, handler.GetSymbols)
	e.POST(SymbolsUpdate, handler.PostSymbolsUpdate)
	e.POST(SymbolsRecreate, handler.PostSymbolsRecreate)
}

func (h *SymbolHandler) GetSymbols(c echo.Context) error {
	pageSizeQ := c.QueryParam("pageSize")
	pageNumberQ := c.QueryParam("pageNumber")
	orderQ := c.QueryParam("order")
	ascQ := c.QueryParam("asc")

	pageSize, err := strconv.Atoi(pageSizeQ)
	if err != nil {
		return c.JSON(common.GetErrorResponse(common.InvalidModelError))
	}
	pageNumber, err := strconv.Atoi(pageNumberQ)
	if err != nil {
		return c.JSON(common.GetErrorResponse(common.InvalidModelError))
	}
	if orderQ == "" {
		return c.JSON(common.GetErrorResponse(common.InvalidModelError))
	}
	asc, err := strconv.ParseBool(ascQ)
	if err != nil {
		return c.JSON(common.GetErrorResponse(err))
	}

	res, err := h.symbolService.Get(pageSize, pageNumber, orderQ, asc)
	if err != nil {
		return c.JSON(common.GetErrorResponse(err))
	}

	return c.JSON(http.StatusOK, res)
}

func (h *SymbolHandler) PostSymbolsUpdate(c echo.Context) error {

	return nil
}

func (h *SymbolHandler) PostSymbolsRecreate(c echo.Context) error {
	ok, err := h.trading212Service.GetSymbols()
	if err != nil {
		return c.JSON(common.GetErrorResponse(err))
	}
	return c.JSON(http.StatusCreated, ok)
}
