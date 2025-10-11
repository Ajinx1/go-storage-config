package eureka

import (
	"fmt"
)

func InitEureka(theConfig EurekaConfig) (*Client, error) {
	conn, err := ConnectFromEnv(theConfig)
	if err != nil {
		return nil, fmt.Errorf("eureka connection failed: %v", err)
	}
	return NewClient(conn), nil
}
