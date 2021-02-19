package db

import (
	"github.com/dystopia-systems/alaskalog"
	"github.com/jackc/pgx"
	"github.com/vectorman1/analysis/analysis-api/common"
	"github.com/vectorman1/analysis/analysis-api/entity"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func GetGormDb(config *common.Config, dryRun bool) (*gorm.DB, error) {
	dsn := "host=localhost user=harb password=HueHue123 dbname=analysis port=5432 sslmode=disable TimeZone=Europe/Sofia"
	db, err := gorm.Open(postgres.New(
		postgres.Config{
			DSN:                  dsn,
			PreferSimpleProtocol: true,
		}),
		&gorm.Config{
			SkipDefaultTransaction: true,
		})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func Migrate(config *common.Config) error {
	db, err := GetGormDb(config, false)
	if err != nil {
		alaskalog.Logger.Fatalf("error initializing gorm db: %v", err)
	}
	d, _ := db.DB()
	defer d.Close()

	alaskalog.Logger.Infoln("Starting migration...")

	err = db.AutoMigrate(&entity.Currency{})
	if err != nil {
		return err
	}
	err = db.AutoMigrate(&entity.Symbol{})
	if err != nil {
		return err
	}
	err = db.AutoMigrate(&entity.Report{})
	if err != nil {
		return err
	}
	err = db.AutoMigrate(&entity.History{})
	if err != nil {
		return err
	}
	err = db.AutoMigrate(&entity.Signal{})
	if err != nil {
		return err
	}
	err = db.AutoMigrate(&entity.Strategy{})
	if err != nil {
		return err
	}

	alaskalog.Logger.Infoln("Migration was successful.")
	return nil
}

var maxDBConnections = 10

func GetPgDb(config *common.Config) (*pgx.ConnPool, error) {
	alaskalog.Logger.Infoln("Opening PostgreSQL connection...")
	cfg := pgx.ConnConfig{
		Host:     "localhost",
		User:     "harb",
		Password: "HueHue123",
		Database: "analysis",
	}

	poolConfig := pgx.ConnPoolConfig{
		ConnConfig:     cfg,
		MaxConnections: maxDBConnections,
	}

	return pgx.NewConnPool(poolConfig)
}
