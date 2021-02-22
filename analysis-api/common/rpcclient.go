package common

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type RpcClient interface {
	grpc.ClientConnInterface
	GetConnection() error
	LoadTLSCredentials() error
}

type Rpc struct {
	RpcClient
	credentials *credentials.TransportCredentials
	Connection  *grpc.ClientConn
}

func (r *Rpc) Initialize() (*Rpc, error) {
	err := r.LoadTLSCredentials()
	if err != nil {
		return nil, err
	}
	err = r.GetConnection()
	if err != nil {
		return nil, err
	}
	return r, nil
}

func (r *Rpc) LoadTLSCredentials() error {
	pemServerCA, err := ioutil.ReadFile("certs/ca-cert.pem")
	if err != nil {
		return err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return fmt.Errorf("unable to load TLS configuration for worker client")
	}

	config := tls.Config{
		RootCAs: certPool,
	}
	var creds credentials.TransportCredentials
	creds = credentials.NewTLS(&config)
	r.credentials = &creds
	return nil
}

func (r *Rpc) GetConnection() error {
	conn, err := grpc.Dial(":6969", grpc.WithTransportCredentials(*r.credentials))
	if err != nil {
		return err
	}
	r.Connection = conn
	return nil
}
