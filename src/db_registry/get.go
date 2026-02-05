package db_registry

import (
	"fmt"

	"gorm.io/gorm"
)

func (r *Registry) Get(dbName string) (*gorm.DB, error) {
	r.mu.RLock()
	db, ok := r.dbs[dbName]
	r.mu.RUnlock()
	if ok {
		return db, nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if db, ok := r.dbs[dbName]; ok {
		return db, nil
	}

	src, ok := r.sources[dbName]
	if !ok {
		return nil, fmt.Errorf("unknown datasource: %s", dbName)
	}

	conn, err := r.dbCreate(src.DatabaseName)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to connect to datasource %s (%s): %w",
			src.Code,
			src.DatabaseName,
			err,
		)
	}

	r.dbs[dbName] = conn
	return conn, nil
}
