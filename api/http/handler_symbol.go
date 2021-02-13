package http

import "analysis-api/service"
import "github.com/labstack/echo"

type SymbolHandler struct {
	SymbolService *service.SymbolService
}

func NewSymbolHandler(e *echo.Echo, symbolService *service.SymbolService) {
	handler := SymbolHandler{
		SymbolService: symbolService,
	}
}
