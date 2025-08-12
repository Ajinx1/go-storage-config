package db

import (
	"fmt"

	rdb "github.com/you/go-storage-config/src/db/redis"
)

func InitRedis(theConfig rdb.RedisConfig) (*rdb.Client, error) {
	conn, err := rdb.ConnectFromEnv(theConfig)
	if err != nil {
		return nil, fmt.Errorf("redis connection failed: %v", err)
	}
	return rdb.NewClient(conn), nil
}
