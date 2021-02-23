package common

import (
	"crypto/tls"
	"github.com/vectorman1/analysis/analysis-worker/generated/proto_models"
	"google.golang.org/grpc/credentials"
	"net/http"
)

func ContainsSymbol(uuid string, arr []*proto_models.Symbol) (bool, *proto_models.Symbol) {
	for _, v := range arr {
		if v.Uuid == uuid {
			return true, v
		}
	}
	return false, nil
}

func LoadTLSCredentials() (credentials.TransportCredentials, error) {
	c := &SarumanClient{httpClient: &http.Client{}}
	certPath, err := c.GetServerCert()
	if err != nil {
		return nil, err
	}

	keyPath, err := c.GetServerKey()
	if err != nil {
		return nil, err
	}

	serverCert, err := tls.LoadX509KeyPair(certPath, keyPath)
	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth: tls.NoClientCert,
	}

	return credentials.NewTLS(config), nil
}
