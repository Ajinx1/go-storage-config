package db_registry

import "gorm.io/gorm"

func Init(metaDB *gorm.DB, factory DBFactory) (*Registry, error) {
	sources, err := loadActiveSources(metaDB)
	if err != nil {
		return nil, err
	}

	srcMap := make(map[string]ReportDataSource)
	for _, s := range sources {
		if s.DatabaseName != "" {
			srcMap[s.DatabaseName] = s
		}
	}

	return &Registry{
		dbs:      make(map[string]*gorm.DB),
		sources:  srcMap,
		dbCreate: factory,
	}, nil
}
