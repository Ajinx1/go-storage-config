package redis

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

func LoadRedisConfig(theConfig RedisConfig) *RedisConfig {

	return &RedisConfig{
		Host:     theConfig.Host,
		Port:     theConfig.Port,
		Password: theConfig.Password,
		DB:       theConfig.DB,
	}
}
