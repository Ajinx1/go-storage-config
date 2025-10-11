package eureka

import (
	"fmt"
	"time"

	"go.uber.org/zap"
)

type Client struct {
	conn *EurekaConn
}

func NewClient(conn *EurekaConn) *Client {
	return &Client{conn: conn}
}

func (c *Client) Register() error {
	err := c.conn.Client.RegisterInstance(c.conn.ServiceName, c.conn.Instance)
	if err != nil {
		fmt.Println("Failed to register with Eureka", zap.Error(err))
		return err
	}
	fmt.Println("Registered with Eureka", zap.String("service", c.conn.ServiceName))
	return nil
}

func (c *Client) StartHeartbeats() {
	go func() {
		for {
			time.Sleep(30 * time.Second)
			err := c.conn.Client.SendHeartbeat(c.conn.Instance.App, c.conn.Instance.InstanceID)
			if err != nil {
				fmt.Println("Failed to send heartbeat", zap.Error(err))
			} else {
				fmt.Println("Heartbeat sent to Eureka")
			}
		}
	}()
}

func (c *Client) Deregister() {
	c.conn.Instance.Status = "DOWN"
	err := c.conn.Client.UnregisterInstance(c.conn.ServiceName, c.conn.Instance.InstanceID)
	if err != nil {
		fmt.Println("Failed to deregister from Eureka", zap.Error(err))
		return
	}
	fmt.Println("Deregistered from Eureka", zap.String("service", c.conn.ServiceName))
}
