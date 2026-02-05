package db_registry

import "gorm.io/gorm"

func loadActiveSources(db *gorm.DB) ([]DataSource, error) {
	var data []DataSource
	err := db.
		Where("active = ?", true).
		Find(&data).
		Error
	return data, err
}
