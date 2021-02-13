package db

import (
	"analysis-api/common"
	"analysis-api/entity"
	"github.com/dystopia-systems/alaskalog"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDb(config *common.Config) (*gorm.DB, error) {
	alaskalog.Logger.Infoln("Opening MySQL connection...")
	db, err := gorm.Open(mysql.Open(config.MySqlConnectionString), &gorm.Config{
		Logger: logger.New(nil,
			logger.Config{
				LogLevel: logger.Silent,
			}),
	})
	if err != nil {
		return nil, err
	}

	alaskalog.Logger.Infoln("MySQL connection was successful")
	return db, nil
}

func Migrate(db *gorm.DB) error {
	alaskalog.Logger.Infoln("Starting migration...")

	err := db.AutoMigrate(&entity.Symbol{})
	if err != nil {
		return err
	}
	err = db.AutoMigrate(&entity.Currency{})
	if err != nil {
		return err
	}

	alaskalog.Logger.Infoln("Migration was successful.")
	return nil
}
