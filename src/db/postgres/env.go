package postgres

func LoadPostgresConfigFromEnv(theConfig Config) Config {

	return Config{
		Host:     theConfig.Host,
		Port:     theConfig.Port,
		User:     theConfig.User,
		Password: theConfig.Password,
		DBName:   theConfig.DBName,
		SSLMode:  theConfig.SSLMode,
	}
}
