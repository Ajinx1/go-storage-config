package eureka

import "github.com/ArthurHlt/go-eureka-client/eureka"

type EurekaConfig struct {
	URL         string
	ServiceName string
	Port        string
}

type EurekaConn struct {
	Client      *eureka.Client
	Instance    *eureka.InstanceInfo
	ServiceName string
}
