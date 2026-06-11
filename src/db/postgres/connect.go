package postgres

import (
	"strconv"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectFromEnv(theConfig Config) (*gorm.DB, error) {
	return Connect(theConfig)
}

func Connect(config Config) (*gorm.DB, error) {
	config = config.normalize()
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

	sqlDB.SetMaxOpenConns(config.MaxOpenConns)
	sqlDB.SetMaxIdleConns(config.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(config.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(config.ConnMaxIdleTime)

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
