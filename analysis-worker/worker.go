package main

import (
	"github.com/dystopia-systems/alaskalog"
	"github.com/vectorman1/analysis/analysis-worker/common"
	"github.com/vectorman1/analysis/analysis-worker/generated/trading212_service"
	"github.com/vectorman1/analysis/analysis-worker/service"
	"google.golang.org/grpc"
	"log"
	"net"
)

func main() {
	cfg, err := common.GetConfig()
	if err != nil {
		log.Fatalf("failed to get config: %v", err)
	}
	log.Println("loaded config")

	tls, err := common.LoadTLSCredentials()
	if err != nil {
		log.Fatalf("failed to read TLS config: %v", err)
	}
	log.Println("loaded TLS config")

	log.Println("Starting server...")
	lis, err := net.Listen("tcp", "localhost:6969")
	if err != nil {
		log.Fatalf("Failed to listen on port :6969: %v", err)
	}
	log.Println("listening on port :6969")

	s := grpc.NewServer(
		grpc.Creds(tls))
	trading212_service.RegisterTrading212ServiceServer(s,
		service.Trading212Service{}.New(
			common.TRADING212_INSTRUMENTS_LINK,
			common.TRADING212_SHOW_ALL_BUTTON_SELECTOR,
			common.TRADING212_ALL_INSTRUMENTS_SELECTOR,
			cfg))

	log.Println("initialized grpc server")

	if err := s.Serve(lis); err != nil {
		alaskalog.Logger.Fatalf("server failed with: %v", err)
	}
}
