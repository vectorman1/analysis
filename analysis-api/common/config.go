package common

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"github.com/dystopia-systems/alaskalog"
)

type Config struct {
	JwtSigningSecret     string `json:"jwt_signing_secret"`

	// gRPC grpc-server start parameters section
	// gRPC is TCP port to listen by gRPC grpc-server
	GRPCPort string `json:"grpc_port"`

	// HTTP/REST gateway start parameters section
	// HTTPPort is TCP port to listen by HTTP/REST gateway
	HTTPPort string `json:"http_port"`

	// DB Datastore parameters section
	// DatastoreDBHost is host of database
	DatastoreDBHost string `json:"datastore_db_host"`
	// DatastoreDBUser is username to connect to database
	DatastoreDBUser string `json:"datastore_db_user"`
	// DatastoreDBPassword password to connect to database
	DatastoreDBPassword string `json:"datastore_db_password"`
	// DatastoreDBSchema is schema of database
	DatastoreDBSchema string `json:"datastore_db_schema"`

	// DatabaseMaxConnections is the maximum amount of connection pool connections to the database
	DatabaseMaxConnections int `json:"database_max_connections"`

	// Log parameters section
	// LogLevel is global log level: Debug(-1), Info(0), Warn(1), Error(2), DPanic(3), Panic(4), Fatal(5)
	LogLevel int `json:"log_level"`
	// LogTimeFormat is print time format for logger-grpc e.g. 2006-01-02T15:04:05Z07:00
	LogTimeFormat string `json:"log_time_format"`
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