package common

import (
	"google.golang.org/grpc/grpclog"
	"io/ioutil"
	"net/http"
	"os"
)

type SarumanClient struct {
	httpClient *http.Client
}

// TODO refactor
func (c *SarumanClient) GetServerCert() (string, error) {
	grpclog.Infoln("Getting configuration from Saruman...")

	url := "https://saruman-api.glamav.systems/api/v1/config/analysis-worker-cert/"
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Api-Key", os.Getenv("SARUMAN_API_KEY"))
	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	key, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	path := "server-cert.pem"
	err = ioutil.WriteFile(path, key, 0644)
	if err != nil {
		return "", err
	}

	return path, nil
}

// TODO refactor
func (c *SarumanClient) GetServerKey() (string, error) {
	grpclog.Infoln("Getting configuration from Saruman...")

	url := "https://saruman-api.glamav.systems/api/v1/config/analysis-worker-key/"
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Api-Key", os.Getenv("SARUMAN_API_KEY"))
	res, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	key, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	path := "server-key.pem"
	err = ioutil.WriteFile(path, key, 0644)
	if err != nil {
		return "", err
	}

	return path, nil
}
