package infrastructure

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/vectorman1/analysis/analysis-api/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
)

type RpcClient interface {
	grpc.ClientConnInterface
	GetConnection() error
	LoadTLSCredentials() error
}

type Rpc struct {
	RpcClient
	credentials *credentials.TransportCredentials
	connection  *grpc.ClientConn
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
	pemServerCA, err := ioutil.ReadFile("infrastructure/proto/certs/ca-cert.pem")
	if err != nil {
		return err
	}

	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(pemServerCA) {
		return common.RPCTLSError
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
	r.connection = conn
	return nil
}
