package db

import (
	"github.com/jackc/pgx"
	"github.com/vectorman1/analysis/analysis-api/common"
)

func GetConnPool(config *common.Config) (*pgx.ConnPool, error) {
	cfg := pgx.ConnConfig{
		Host:     config.DatastoreDBHost,
		User:     config.DatastoreDBUser,
		Password: config.DatastoreDBPassword,
		Database: config.DatastoreDBSchema,
	}

	poolConfig := pgx.ConnPoolConfig{
		ConnConfig:     cfg,
		MaxConnections: config.DatabaseMaxConnections,
	}

	return pgx.NewConnPool(poolConfig)
}
