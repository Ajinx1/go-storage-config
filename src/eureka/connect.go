package eureka

import (
	"fmt"
)

func ConnectFromEnv(theConfig EurekaConfig) (*EurekaConn, error) {
	cfg := LoadEurekaConfig(theConfig)

	conn, err := NewConnection(cfg)
	if err != nil {
		return nil, fmt.Errorf("eureka connection failed: %v", err)
	}

	return conn, nil
}

func LoadEurekaConfig(theConfig EurekaConfig) *EurekaConfig {
	return &EurekaConfig{
		URL:         theConfig.URL,
		ServiceName: theConfig.ServiceName,
		Port:        theConfig.Port,
	}
}
