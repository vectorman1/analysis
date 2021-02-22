package cors_rest

import (
	"log"
	"net/http"

	"github.com/vectorman1/analysis/analysis-api/common"

	"github.com/gorilla/handlers"
)

// GetCORS returns a handler with configured CORS
func GetCORS() func(http.Handler) http.Handler {
	config, err := common.GetConfig()
	if err != nil {
		log.Fatalf("couldn't get config: %v", err)
		return nil
	}
	originsAllowed := handlers.AllowedOrigins([]string{})
	if config.Environment == common.Development {
		originsAllowed = handlers.AllowedOrigins([]string{"*"})
	} else if config.Environment == common.Production {
		originsAllowed = handlers.AllowedOrigins([]string{"*.dystopia.systems"})
	}

	return getCORS(originsAllowed)
}

// getCORS returns the gorilla CORS handler
func getCORS(origins handlers.CORSOption) func(http.Handler) http.Handler {
	headers := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	methods := handlers.AllowedMethods([]string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions})

	return handlers.CORS(headers, methods, origins)
}
