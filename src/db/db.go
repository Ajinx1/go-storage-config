package db

import (
	"errors"

	"github.com/Ajinx1/go-storage-config/src/db/postgres"
	"github.com/Ajinx1/go-storage-config/src/db/sqlserver"
	"github.com/Ajinx1/go-storage-config/src/utils"

	"gorm.io/gorm"
)

const (
	Postgres  = "postgres"
	SQLServer = "sqlserver"
)

func Connect(driver string, theConfig interface{}) (*gorm.DB, error) {
	utils.LoadEnv()

	switch driver {
	case Postgres:
		if cfg, ok := theConfig.(postgres.Config); ok {
			return postgres.ConnectFromEnv(cfg)
		}
	case SQLServer:
		if cfg, ok := theConfig.(sqlserver.Config); ok {
			return sqlserver.ConnectFromEnv(cfg)
		}
	}
	return nil, errors.New("unsupported or invalid database config for driver: " + driver)
}
