package postgres

import (
	"strconv"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(config Config) (*gorm.DB, error) {
	dsn := getDSN(config)
	return gorm.Open(postgres.Open(dsn), &gorm.Config{})
}

func ConnectFromEnv(theConfig Config) (*gorm.DB, error) {
	config := LoadPostgresConfigFromEnv(theConfig)
	return Connect(config)
}

func getDSN(cfg Config) string {
	return "host=" + cfg.Host +
		" port=" + strconv.Itoa(cfg.Port) +
		" user=" + cfg.User +
		" password=" + cfg.Password +
		" dbname=" + cfg.DBName +
		" sslmode=" + cfg.SSLMode
}
