package rest_server

import (
	"context"

	cors_rest "github.com/vectorman1/analysis/analysis-api/middleware/cors-rest"

	logger_rest "github.com/vectorman1/analysis/analysis-api/middleware/logger-rest"

	"github.com/vectorman1/analysis/analysis-api/middleware/http_rest"

	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/vectorman1/analysis/analysis-api/generated/symbol_service"
	logger_grpc "github.com/vectorman1/analysis/analysis-api/middleware/logger-grpc"
	tracer_rest "github.com/vectorman1/analysis/analysis-api/middleware/tracer-rest"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// RunServer runs HTTP/REST gateway
func RunServer(ctx context.Context, grpcPort, httpPort string) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	errorHandler := runtime.WithErrorHandler(http_rest.HandleMuxError)
	mux := runtime.NewServeMux(errorHandler)
	opts := []grpc.DialOption{grpc.WithInsecure()}

	if err := symbol_service.RegisterSymbolServiceHandlerFromEndpoint(ctx, mux, "localhost:"+grpcPort, opts); err != nil {
		logger_grpc.Log.Fatal("failed to start HTTP gateway", zap.String("reason", err.Error()))
	}

	// configure CORS headers options
	corsH := cors_rest.GetCORS()

	srv := &http.Server{
		Addr: ":" + httpPort,
		Handler: corsH(tracer_rest.AddRequestID(
			logger_rest.AddLogger(logger_grpc.Log, mux))),
	}

	// graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			// sig is a ^C, handle it
		}

		_, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		_ = srv.Shutdown(ctx)
	}()

	log.Println("starting HTTP/REST gateway...")
	return srv.ListenAndServe()
}
