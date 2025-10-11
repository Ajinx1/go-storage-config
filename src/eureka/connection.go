package eureka

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/ArthurHlt/go-eureka-client/eureka"
	"go.uber.org/zap"
)

func NewConnection(cfg *EurekaConfig) (*EurekaConn, error) {
	client := eureka.NewClient([]string{cfg.URL})

	ipAddr, err := getLocalIP()
	if err != nil {
		fmt.Println("Could not get local IP", zap.Error(err))
		ipAddr = "127.0.0.1"
	}

	fmt.Printf("ip is: %s\n", ipAddr)

	hostName := fmt.Sprintf("%s-%s", strings.ToLower(cfg.ServiceName), ipAddr)
	instanceID := fmt.Sprintf("%s:%s:%s", strings.ToLower(cfg.ServiceName), ipAddr, cfg.Port)

	instance := eureka.NewInstanceInfo(
		hostName,
		strings.ToUpper(cfg.ServiceName),
		ipAddr,
		mustAtoi(cfg.Port),
		30,
		false,
	)

	instance.InstanceID = instanceID
	instance.Status = "UP"
	instance.VipAddress = cfg.ServiceName
	instance.IpAddr = ipAddr
	instance.HostName = ipAddr
	instance.Port = &eureka.Port{
		Port:    mustAtoi(cfg.Port),
		Enabled: true,
	}
	instance.DataCenterInfo = &eureka.DataCenterInfo{
		Class: "com.netflix.appinfo.InstanceInfo$DefaultDataCenterInfo",
		Name:  "MyOwn",
	}
	instance.Metadata = &eureka.MetaData{
		Map: map[string]string{
			"instanceId": instanceID,
		},
	}
	instance.HealthCheckUrl = fmt.Sprintf("http://%s:%s/api/v1/health", ipAddr, cfg.Port)
	instance.StatusPageUrl = instance.HealthCheckUrl

	return &EurekaConn{
		Client:      client,
		Instance:    instance,
		ServiceName: cfg.ServiceName,
	}, nil
}

func getLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", fmt.Errorf("no valid IP found")
}

func mustAtoi(s string) int {
	v, _ := strconv.Atoi(s)
	return v
}
