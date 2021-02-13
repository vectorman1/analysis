package common

import (
	"encoding/json"
	"github.com/dystopia-systems/alaskalog"
	"io/ioutil"
	"net/http"
	"os"
)

type Config struct {
	MySqlConnectionString string `json:"my_sql_connection_string"`
	JwtSigningSecret      string `json:"jwt_signing_secret"`
}

func GetConfig() (*Config, error) {
	alaskalog.Logger.Infoln("Getting configuration from Saruman...")

	client := &http.Client{}
	url := os.Getenv("SARUMAN_URL")
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	req.Header.Set("Api-Key", os.Getenv("SARUMAN_API_KEY"))
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(body, &config)
	if err != nil {
		return nil, err
	}
	alaskalog.Logger.Infoln("Configuration loaded.")

	return &config, nil
}
