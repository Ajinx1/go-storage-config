package db_registry

import (
	"fmt"

	"gorm.io/gorm"
)

func (r *Registry) Get(code string) (*gorm.DB, error) {

	r.mu.RLock()
	db, ok := r.dbs[code]
	r.mu.RUnlock()

	if ok {
		return db, nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if db, ok := r.dbs[code]; ok {
		return db, nil
	}

	src, ok := r.sources[code]
	if !ok {
		return nil, fmt.Errorf("unknown datasource: %s", code)
	}

	conn, err := r.dbCreate(src)
	if err != nil {
		return nil, fmt.Errorf(
			"failed to connect to datasource %s (%s): %w",
			src.Code,
			src.DatabaseName,
			err,
		)
	}

	r.dbs[code] = conn

	return conn, nil
}
