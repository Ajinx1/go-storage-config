package db_registry

type ReportDataSource struct {
	ID           uint   `gorm:"primaryKey" json:"id"`
	Code         string `gorm:"uniqueIndex;not null" json:"code"`
	Name         string `gorm:"not null" json:"name"`
	DatabaseName string `gorm:"not null" json:"database_name"`
	SchemaName   string `json:"schema_name"`
	Active       bool   `gorm:"default:true" json:"active"`
}

