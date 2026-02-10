package db_registry

import (
	"sync"

	"gorm.io/gorm"
)

type DBFactory func(databaseName string) (*gorm.DB, error)

type Registry struct {
	mu       sync.RWMutex
	dbs      map[string]*gorm.DB
	sources  map[string]ReportDataSource
	dbCreate DBFactory
}
