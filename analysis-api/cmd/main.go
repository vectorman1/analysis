package main

import (
	"context"
	"fmt"
	"os"

	"github.com/vectorman1/analysis/analysis-api/common"
	"github.com/vectorman1/analysis/analysis-api/db"
	logger_grpc "github.com/vectorman1/analysis/analysis-api/middleware/logger-grpc"
	grpc_server "github.com/vectorman1/analysis/analysis-api/server/grpc-server"
	rest_server "github.com/vectorman1/analysis/analysis-api/server/rest-server"
	"github.com/vectorman1/analysis/analysis-api/service"
)

// RunServer runs gRPC grpc-server and HTTP gateway
func RunServer() error {
	ctx := context.Background()
	config, err := common.GetConfig()
	if err != nil {
		return err
	}

	// get configuration
	if len(config.GRPCPort) == 0 {
		return fmt.Errorf("invalid TCP port for gRPC grpc-server: '%s'", config.GRPCPort)
	}
	if len(config.HTTPPort) == 0 {
		return fmt.Errorf("invalid TCP port for HTTP gateway: '%s'", config.HTTPPort)
	}
	// initialize logger-grpc
	if err := logger_grpc.Init(config.LogLevel, config.LogTimeFormat); err != nil {
		return fmt.Errorf("failed to initialize logger-grpc: %v", err)
	}

	dbConnPool, err := db.GetConnPool(config)
	if err != nil {
		return fmt.Errorf("failed to create conn pool: %v", err)
	}

	conn, err := dbConnPool.Acquire()
	if err != nil {
		return fmt.Errorf("failed to open database: %v", err)
	}
	defer conn.Close()

	// connect to worker rpc server
	rpcClient := &common.Rpc{}
	_, err = rpcClient.Initialize()
	if err != nil {
		return fmt.Errorf("failed to connect to worker rpc server: %v", err)
	}

	symbolsRepository := db.NewSymbolRepository(dbConnPool)
	currencyRepository := db.NewCurrencyRepository(dbConnPool)

	symbolsServiceServer := service.NewSymbolsServiceServer(rpcClient, symbolsRepository, currencyRepository)

	grpcServer := grpc_server.NewGRPCServer(ctx, config.GRPCPort, symbolsServiceServer)

	// run HTTP gateway
	go func() {
		_ = rest_server.RunServer(ctx, config.GRPCPort, config.HTTPPort)
	}()

	return grpcServer.Run()
}

func main() {
	if err := RunServer(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
