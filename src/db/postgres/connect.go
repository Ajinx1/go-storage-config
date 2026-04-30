package postgres

import (
	"strconv"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectFromEnv(theConfig Config) (*gorm.DB, error) {
	config := LoadPostgresConfigFromEnv(theConfig)
	return Connect(config)
}

func Connect(config Config) (*gorm.DB, error) {
	dsn := getDSN(config)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		PrepareStmt: true,
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxOpenConns(30)
	sqlDB.SetMaxIdleConns(15)
	sqlDB.SetConnMaxLifetime(20 * time.Minute)
	sqlDB.SetConnMaxIdleTime(10 * time.Minute)

	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func getDSN(cfg Config) string {
	return "host=" + cfg.Host +
		" port=" + strconv.Itoa(cfg.Port) +
		" user=" + cfg.User +
		" password=" + cfg.Password +
		" dbname=" + cfg.DBName +
		" sslmode=" + cfg.SSLMode
}
