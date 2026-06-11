package db_registry

import "gorm.io/gorm"

func (r *Registry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, db := range r.dbs {
		sqlDB, err := db.DB()
		if err != nil {
			continue
		}

		if err := sqlDB.Close(); err != nil {
			return err
		}
	}

	r.dbs = make(map[string]*gorm.DB)

	return nil
}
