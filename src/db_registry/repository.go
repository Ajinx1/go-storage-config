package db_registry

import "gorm.io/gorm"

func loadActiveSources(db *gorm.DB) ([]ReportDataSource, error) {
	var data []ReportDataSource
	err := db.
		Where("active = ?", true).
		Find(&data).
		Error
	return data, err
}
