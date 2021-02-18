package common

import (
	"crypto/tls"
	"github.com/vectorman1/analysis/analysis-worker/proto"
	"golang.org/x/net/html"
	"google.golang.org/grpc/credentials"
	"sync"
)

func ContainsSymbol(isin string, identifier string, arr []*proto.Symbol) (bool, *proto.Symbol) {
	for _, v := range arr {
		if v.Identifier == identifier &&
			v.ISIN == isin {
			return true, v
		}
	}
	return false, nil
}

func LoadTLSCredentials() (credentials.TransportCredentials, error) {
	serverCert, err := tls.LoadX509KeyPair("proto/certs/server-cert.pem", "proto/certs/server-key.pem")
	if err != nil {
		return nil, err
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth: tls.NoClientCert,
	}

	return credentials.NewTLS(config), nil
}

type WorkList struct {
	m sync.Mutex
	Output chan *html.Node
	Input chan *html.Node
}

func (w *WorkList) Flatten(n *html.Node) {
	stack := []*html.Node{n}

	for {
		if stack[len(stack)-1].NextSibling == nil && stack[len(stack)-1].FirstChild == nil {
			stack = stack[:len(stack)-1]
		}

		if len(stack) == 0 {
			break
		}

		stack[len(stack)-1] = n.FirstChild
		stack = append(stack, n.NextSibling)
	}
}